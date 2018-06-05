package gloo

import (
	"bytes"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/solo-io/qloo/pkg/exec"
	"github.com/solo-io/qloo/pkg/api/types/v1"
	"github.com/solo-io/qloo/pkg/util"
	"text/template"
	"fmt"
)

type ResolverFactory struct {
	proxyAddr string
}

func NewResolverFactory(proxyAddr string) *ResolverFactory {
	return &ResolverFactory{
		proxyAddr: proxyAddr,
	}
}

func (rf *ResolverFactory) CreateResolver(path ResolverPath, glooResolver *v1.GlooResolver) (exec.RawResolver, error) {
	requestBodyTemplate := glooResolver.RequestTemplate
	responseBodyTemplate := glooResolver.ResponseTemplate
	contentType := glooResolver.ContentType
	if contentType == "" {
		contentType = "application/json"
	}
	var (
		requestTemplate  *template.Template
		responseTemplate *template.Template
		err              error
	)

	if requestBodyTemplate != "" {
		requestTemplate, err = util.Template(requestBodyTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "parsing request body template failed")
		}
	}
	if responseBodyTemplate != "" {
		responseTemplate, err = util.Template(responseBodyTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "parsing response body template failed")
		}
	}

	return rf.newResolver(path, contentType, requestTemplate, responseTemplate), nil
}

func (rf *ResolverFactory) newResolver(path ResolverPath, contentType string, requestTemplate, responseTemplate *template.Template) exec.RawResolver {
	return func(params exec.Params) ([]byte, error) {
		body := &bytes.Buffer{}
		if requestTemplate != nil {
			buf, err := util.ExecTemplate(requestTemplate, params)
			if err != nil {
				// TODO: sanitize
				return nil, errors.Wrapf(err, "executing request template for params %v", params)
			}
			body = buf
		}
		url := "http://" + rf.proxyAddr + path.Path()
		res, err := http.Post(url, contentType, body)
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
		if err := responseTemplate.Execute(buf, input); err != nil {
			return nil, errors.Wrapf(err, "executing response template for response %v", input)
		}
		return buf.Bytes(), nil
	}
}

type ResolverPath struct {
	TypeName  string
	FieldName string
}

func (p ResolverPath) Path() string {
	return fmt.Sprintf("/%v.%v", p.TypeName, p.FieldName)
}
