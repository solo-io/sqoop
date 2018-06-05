package resolvers

import (
	"github.com/solo-io/qloo/pkg/api/types/v1"
	"github.com/solo-io/qloo/pkg/exec"
	"github.com/pkg/errors"
	"github.com/solo-io/qloo/pkg/resolvers/node"
	"github.com/solo-io/qloo/pkg/resolvers/template"
	"github.com/solo-io/qloo/pkg/resolvers/gloo"
)

type ResolverFactory struct {
	glooResolverFactory gloo.ResolverFactory
}

func (rf *ResolverFactory) CreateResolver(typeName, fieldName string, resolverMap *v1.ResolverMap) (exec.RawResolver, error) {
	if len(resolverMap.Types) == 0 {
		return nil, errors.Errorf("no types defined in resolver map %v", resolverMap.Name)
	}
	typeResolver, ok := resolverMap.Types[typeName]
	if !ok {
		return nil, errors.Errorf("type %v not found in resolver map %v", typeName, resolverMap.Name)
	}
	if len(typeResolver.Fields) == 0 {
		return nil, errors.Errorf("no fields defined for type %v in resolver map %v", typeName, resolverMap.Name)
	}
	fieldResolver, ok := typeResolver.Fields[fieldName]
	if !ok {
		return nil, errors.Errorf("field %v not found for type %v in resolver map %v",
			fieldName, typeResolver, resolverMap.Name)
	}
	switch resolver := fieldResolver.Resolver.(type) {
	case *v1.Resolver_NodejsResolver:
		return node.NewNodeResolver(resolver.NodejsResolver)
	case *v1.Resolver_TemplateResolver:
		return template.NewTemplateResolver(resolver.TemplateResolver)
	case *v1.Resolver_GlooResolver:
		path := gloo.ResolverPath{TypeName: typeName, FieldName: fieldName}
		return rf.glooResolverFactory.CreateResolver(path, resolver.GlooResolver)
	}
	panic("unknown resolver type")
}