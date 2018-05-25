package dynamic
//
//import (
//	"github.com/vektah/gqlgen/neelance/schema"
//	"context"
//	"github.com/vektah/gqlgen/graphql"
//	"github.com/vektah/gqlgen/neelance/query"
//)
//
////
////var resolvers = Map{w
////	"character": Map{
////		"name": Leaf(func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
////			return "Luke", nil
////		}),
////	},
////}
//
//type ResolverFunc func(ctx context.Context, args map[string]interface{}) (interface{}, error)
//
//type SchemaResolver struct {
//	*schema.Schema
//	*graphql.RequestContext
//	Resovlers map[string]ResolverFunc
//}
//
//func (r *SchemaResolver) Resolve(ctx context.Context, parent string, field graphql.CollectedField) (map[string]interface{}, error) {
//	fields := graphql.CollectFields(r.Doc, field.Selections, []string{}, ec.Variables)
//	results := make(map[string]interface{})
//	for _, sel := range field.Selections {
//		switch sel := sel.(type) {
//
//		}
//	}
//	resolver, ok := r.Resovlers[parent+"."+field.Name]
//	if !ok {
//		results[field.Name] = nil
//	}
//}
