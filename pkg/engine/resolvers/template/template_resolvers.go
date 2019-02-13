package template

import (
	"github.com/pkg/errors"
	"github.com/solo-io/sqoop/pkg/engine/exec"
	"github.com/solo-io/sqoop/pkg/engine/util"
)

func NewTemplateResolver(inlineTemplate string) (exec.RawResolver, error) {
	tmpl, err := util.Template(inlineTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "parsing inline template")
	}
	return func(params exec.Params) ([]byte, error) {
		buf, err := util.ExecTemplate(tmpl, params)
		if err != nil {
			return nil, errors.Wrap(err, "executing inline template")
		}
		return buf.Bytes(), nil
	}, nil
}
