package gloo

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"text/template"

	"github.com/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/sqoop/pkg/api/v1"
	"github.com/solo-io/sqoop/pkg/engine/exec"
	"github.com/solo-io/sqoop/pkg/engine/util"
	"github.com/solo-io/sqoop/pkg/translator"
)

type ResolverFactory struct {
	proxyAddr string
}

func NewResolverFactory(proxyAddr string) *ResolverFactory {
	return &ResolverFactory{
		proxyAddr: proxyAddr,
	}
}

// TODO(ilackarms): support more than just Body in request/response template
func (rf *ResolverFactory) CreateResolver(resolverMap core.ResourceRef, typeName, fieldName string, glooResolver *v1.GlooResolver) (exec.RawResolver, error) {
	requestBodyTemplate := glooResolver.RequestTemplate
	responseBodyTemplate := glooResolver.ResponseTemplate

	contentType := "application/json"
	if requestBodyTemplate != nil {
		// TODO(ilackarms): find package that allows us to convert from http1->http2 header style
		ct := requestBodyTemplate.Headers[":content-type"]
		if ct != "" {
			contentType = ct
		}
		ct = requestBodyTemplate.Headers["Content-Type"]
		if ct != "" {
			contentType = ct
		}
	}

	var (
		requestTemplate  *template.Template
		responseTemplate *template.Template
		err              error
	)

	if requestBodyTemplate != nil {
		requestTemplate, err = util.Template(requestBodyTemplate.Body)
		if err != nil {
			return nil, errors.Wrap(err, "parsing request body template failed")
		}
	}
	if responseBodyTemplate != nil {
		responseTemplate, err = util.Template(responseBodyTemplate.Body)
		if err != nil {
			return nil, errors.Wrap(err, "parsing response body template failed")
		}
	}

	return rf.newResolver(resolverMap, typeName, fieldName, contentType, requestTemplate, responseTemplate), nil
}

func (rf *ResolverFactory) newResolver(resolverMap core.ResourceRef, typeName, fieldName string, contentType string, requestTemplate, responseTemplate *template.Template) exec.RawResolver {
	return func(params exec.Params) ([]byte, error) {
		body := &bytes.Buffer{}

		switch {
		case requestTemplate != nil:
			buf, err := util.ExecTemplate(requestTemplate, params)
			if err != nil {
				// TODO: sanitize
				return nil, errors.Wrapf(err, "executing request template for params %v", params)
			}
			body = buf
		case len(params.Args) > 0:
			if err := json.NewEncoder(body).Encode(params.Args); err != nil {
				return nil, errors.Wrap(err, "failed to encode args")
			}
		}

		url := "http://" + rf.proxyAddr + translator.RoutePath(resolverMap, typeName, fieldName)
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

		var result interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, errors.Wrap(err, "failed to parse response as json object. "+
				"response templates may only be used with JSON responses")
		}
		buf := &bytes.Buffer{}
		if err := responseTemplate.Execute(buf, result); err != nil {
			return nil, errors.Wrapf(err, "executing response template for response %v", result)
		}
		return buf.Bytes(), nil
	}
}
