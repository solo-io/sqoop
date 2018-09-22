package node

import (
	"github.com/pkg/errors"
	"github.com/solo-io/sqoop/pkg/api/types/v1"
	"github.com/solo-io/sqoop/pkg/exec"
)

func NewNodeResolver(resolver *v1.NodeJSResolver) (exec.RawResolver, error) {
	return nil, errors.Errorf("nodejs resolvers currently unsupported")
}
