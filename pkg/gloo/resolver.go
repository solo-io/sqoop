package gloo

import (
	"github.com/solo-io/qloo/pkg/dynamic"
	"bytes"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/solo-io/gloo/pkg/log"
	"text/template"
)

type ResolverFactory struct {
	ProxyAddr string
}

func (gr *ResolverFactory) MakeResolver(path, requestBodyTemplate, responseBodyTemplate, contentType string) (dynamic.RawResolver, error) {
	if contentType == "" {
		contentType = "application/json"
	}
	var (
		requestTemplate  *template.Template
		responseTemplate *template.Template
		err              error
	)

	if requestBodyTemplate != "" {
		requestTemplate, err = template.New("requestBody for " + path).Funcs(template.FuncMap{
			"marshal": func(v interface{}) (string, error) {
				a, err := json.Marshal(v)
				if err != nil {
					return "", err
				}
				log.GreyPrintf("%v", string(a))
				return string(a), nil
			},
		}).Parse(requestBodyTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "parsing request body template failed")
		}
	}
	if responseBodyTemplate != "" {
		responseTemplate, err = template.New("responseBody for " + path).Funcs(template.FuncMap{
			"marshal": func(v interface{}) (string, error) {
				a, err := json.Marshal(v)
				if err != nil {
					return "", err
				}
				log.Printf("parsed: %v", string(a))
				return string(a), nil
			},
		}).Parse(responseBodyTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "parsing response body template failed")
		}
	}

	return func(params dynamic.Params) ([]byte, error) {
		body := bytes.Buffer{}
		if requestTemplate != nil {
			if err := requestTemplate.Execute(&body, params); err != nil {
				// TODO: sanitize
				return nil, errors.Wrapf(err, "executing request template for params %v", params)
			}
		}
		str := body.String()
		log.Printf("body: %v", str)
		if params.Parent != nil {
			log.Printf("source: %v", params.Parent.GoValue())
		}

		url := "http://" + gr.ProxyAddr + path
		res, err := http.Post(url, contentType, &body)
		if err != nil {
			return nil, errors.Wrap(err, "performing http post")
		}

		defer res.Body.Close()
		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, errors.Wrap(err, "reading response body")
		}

		if res.StatusCode < 200 || res.StatusCode >= 300 {
			return nil, errors.Errorf("unexpected status code: %v (%s)", res.StatusCode, data)
		}
		// empty response
		if len(data) == 0 {
			return nil, nil
		}

		// no template, return raw
		if responseTemplate == nil {
			return data, nil
		}

		// requires output to be json object
		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, errors.Wrap(err, "failed to parse response as json object. "+
				"response templates may only be used with JSON responses")
		}
		input := struct {
			Result map[string]interface{}
		}{
			Result: result,
		}
		buf := &bytes.Buffer{}
		if err := requestTemplate.Execute(buf, input); err != nil {
			return nil, errors.Wrapf(err, "executing response template for response %v", input)
		}
		return buf.Bytes(), nil
	}, nil
}
