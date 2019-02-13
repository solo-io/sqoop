package util

import (
	"bytes"
	"encoding/json"
	"text/template"

	"github.com/solo-io/sqoop/pkg/engine/exec"
)

func Template(tmplString string) (*template.Template, error) {
	return template.New("qloo_template").Funcs(templateFuncs).Parse(tmplString)
}

func ExecTemplate(tmpl *template.Template, params exec.Params) (*bytes.Buffer, error) {
	buf := bytes.Buffer{}
	err := tmpl.Execute(&buf, templateParams(params))
	return &buf, err
}

var templateFuncs = template.FuncMap{
	"marshal": func(v interface{}) (string, error) {
		a, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(a), nil
	},
}

type params struct {
	Args   map[string]interface{}
	Parent map[string]interface{}
}

func templateParams(p exec.Params) params {
	var parent map[string]interface{}
	if parentObject, isObject := p.Parent.GoValue().(map[string]interface{}); isObject {
		parent = parentObject
	}
	return params{
		Args:   p.Args,
		Parent: parent,
	}
}
