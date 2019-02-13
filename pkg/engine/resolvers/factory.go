package resolvers

import (
	"github.com/pkg/errors"
	v1 "github.com/solo-io/sqoop/pkg/api/v1"
	"github.com/solo-io/sqoop/pkg/engine/exec"
	"github.com/solo-io/sqoop/pkg/engine/resolvers/gloo"
	"github.com/solo-io/sqoop/pkg/engine/resolvers/node"
	"github.com/solo-io/sqoop/pkg/engine/resolvers/template"
)

type ResolverFactory struct {
	glooResolverFactory *gloo.ResolverFactory
	resolverMap         *v1.ResolverMap
}

func NewResolverFactory(proxyAddr string, resolverMap *v1.ResolverMap) *ResolverFactory {
	return &ResolverFactory{
		glooResolverFactory: gloo.NewResolverFactory(proxyAddr),
		resolverMap:         resolverMap,
	}
}

func (rf *ResolverFactory) CreateResolver(typeName, fieldName string) (exec.RawResolver, error) {
	if len(rf.resolverMap.Types) == 0 {
		return nil, errors.Errorf("no types defined in resolver map %v", rf.resolverMap.Metadata.Ref())
	}
	typeResolver, ok := rf.resolverMap.Types[typeName]
	if !ok {
		return nil, errors.Errorf("type %v not found in resolver map %v", typeName, rf.resolverMap.Metadata.Ref())
	}
	if len(typeResolver.Fields) == 0 {
		return nil, errors.Errorf("no fields defined for type %v in resolver map %v", typeName, rf.resolverMap.Metadata.Ref())
	}
	fieldResolver, ok := typeResolver.Fields[fieldName]
	if !ok {
		return nil, errors.Errorf("field %v not found for type %v in resolver map %v",
			fieldName, typeResolver, rf.resolverMap.Metadata.Ref())
	}
	switch resolver := fieldResolver.Resolver.(type) {
	case *v1.FieldResolver_NodejsResolver:
		return node.NewNodeResolver(resolver.NodejsResolver)
	case *v1.FieldResolver_TemplateResolver:
		return template.NewTemplateResolver(resolver.TemplateResolver.InlineTemplate)
	case *v1.FieldResolver_GlooResolver:
		return rf.glooResolverFactory.CreateResolver(rf.resolverMap.Metadata.Ref(), typeName, fieldName, resolver.GlooResolver)
	}
	// no resolver has been defined
	return nil, nil
}
