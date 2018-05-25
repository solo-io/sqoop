package dynamic

import (
	"context"
	"github.com/pkg/errors"
	"github.com/vektah/gqlgen/graphql"
)

type Resolver struct {}

func NewResolver() *Resolver {
	return &Resolver{}
}

func (r *Resolver) Query(ctx context.Context, field graphql.CollectedField) (interface{}, error) {
	return field, nil
}

func (r *Resolver) Mutation(ctx context.Context, field graphql.CollectedField) (interface{}, error) {
	return field, nil
}

func (r *Resolver) Subscribe(ctx context.Context, field graphql.CollectedField) (interface{}, error) {
	return nil, errors.Errorf("not implmeneted")
}

func resolveField(ctx context.Context, field graphql.CollectedField) (interface{}, error) {
	
	return field, nil
}

//type Node interface {
//	Resolve(ctx context.Context, args map[string]interface{}) (interface{}, error)
//}
//
//type Leaf func(ctx context.Context, args map[string]interface{}) (interface{}, error)
//
//func (f Leaf) Resolve(ctx context.Context, args map[string]interface{}) (interface{}, error) {
//	if f == nil {
//		return nil, errors.Errorf("nil field func")
//	}
//	return f(ctx, args)
//}
//
//type Map map[string]Node
//
//func (m Map) Resolve(ctx context.Context, args map[string]interface{}) (interface{}, error) {
//	if m == nil {
//		return nil, errors.Errorf("nil map")
//	}
//	results := make(map[string]interface{})
//	for k, v := range args {
//		child, ok := m[k]
//		if !ok {
//			return nil, errors.Errorf("no entry in resolver map %v for %s", m, k)
//		}
//		subArgs, ok := v.(map[string]interface{})
//		if !ok {
//			return nil, errors.Errorf("invalid entry in args map: %v", v)
//		}
//		result, err := child.Resolve(ctx, subArgs)
//		if err != nil {
//			return nil, errors.Wrapf(err, "resolver for %v failed", k)
//		}
//		results[k] = result
//	}
//	return results, nil
//}
