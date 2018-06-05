package template

import (
	"github.com/solo-io/qloo/pkg/exec"
	"github.com/pkg/errors"
	"github.com/solo-io/qloo/pkg/api/types/v1"
	"github.com/solo-io/qloo/pkg/util"
)

func NewTemplateResolver(resolver *v1.TemplateResolver) (exec.RawResolver, error) {
	tmpl, err := util.Template(resolver.InlineTemplate)
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
