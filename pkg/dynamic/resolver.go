package dynamic

import (
	"github.com/vektah/gqlgen/neelance/schema"
	"github.com/vektah/gqlgen/neelance/common"
	"github.com/pkg/errors"
	"github.com/solo-io/gloo/pkg/log"
)

// store all the user resolvers
type ResolverMap struct {
	// resolvers for all named types
	Types map[schema.NamedType]*TypeResolver
}

type TypeResolver struct {
	// resolve each field of the type
	Fields map[string]*FieldResolver
}

type FieldResolver struct {
	// type the field resolves to
	Type common.Type

	// how to resolve this field. should return Type
	ResolverFunc ResolverFunc
}

type ResolverFunc func(params Params) (Value, error)

func NewResolverMap(sch *schema.Schema, inputResolvers map[string]ResolverFunc) *ResolverMap {
	typeMap := make(map[schema.NamedType]*TypeResolver)
	for _, t := range sch.Types {
		if metaType(t.TypeName()) {
			continue
		}
		fields := make(map[string]*FieldResolver)
		switch t := t.(type) {
		case *schema.Object:
			for _, f := range t.Fields {
				inputKey := t.Name+"."+f.Name
				log.Printf("looking for resolver: %v", inputKey)
				res := inputResolvers[inputKey]
				if res == nil {
					continue
				}
				fields[f.Name] = &FieldResolver{Type: f.Type, ResolverFunc: res}
			}
		}
		if len(fields) == 0 {
			continue
		}
		typeMap[t] = &TypeResolver{Fields: fields}
	}
	return &ResolverMap{
		Types: typeMap,
	}
}

type Params struct {
	Source map[string]interface{}
	Args   map[string]interface{}
}

func (p Params) Arg(name string) interface{} {
	if len(p.Args) == 0 {
		return nil
	}
	return p.Args[name]
}

func (rm *ResolverMap) Resolve(typ schema.NamedType, field string, params Params) (Value, error) {
	fieldResolver, err := rm.getFieldResolver(typ, field)
	if err != nil {
		return nil, errors.Wrap(err, "resolver lookup")
	}
	data, err := fieldResolver.ResolverFunc(params)
	if err != nil {
		return nil, errors.Wrapf(err, "failed executing resolver for %v.%v", typ.String(), field)
	}
	result, err := convertResult(typ, data)
	if err != nil {
		return nil, errors.Wrap(err, "converting interface{} to result")
	}
	return result, nil
}

func convertResult(typ common.Type, data interface{}) (Value , error) {
	var result Value
	switch typ := typ.(type) {
	case *schema.Object:
		obj, ok := data.(map[string]Value)
		if !ok {
			return nil, errors.Errorf("resolver did not return expected type map[string]interface{}: %v", data)
		}
		result = &Object{
			Object: typ,
			Data: obj,
		}
	case *common.List:
		items, ok := data.([]interface{})
		if !ok {
			return nil, errors.Errorf("resolver did not return expected type map[string]interface{}: %v", data)
		}
		var list []Value
		for _, item := range items {
			val, err := convertResult(typ.OfType, item)
			if err != nil {
				return nil, errors.Wrap(err, "converting array element into result")
			}
			list = append(list, val)
		}
		result = &Array{
			List: typ,
			Data: list,
		}
	}
	return result, nil
}

func (rm *ResolverMap) getFieldResolver(typ schema.NamedType, field string) (*FieldResolver, error) {
	typeResolver, ok := rm.Types[typ]
	if !ok {
		return nil, errors.Errorf("type %v unknown", typ.TypeName())
	}
	fieldResolver, ok := typeResolver.Fields[field]
	if !ok {
		return nil, errors.Errorf("type %v does not contain field %v", typ.TypeName(), field)
	}
	return fieldResolver, nil
}

var metaTypes = []string{
	"Map",
	"Float",
	"ID",
	"Int",
	"Boolean",
	"String",
	"__Type",
	"__TypeKind",
	"__Directive",
	"__EnumValue",
	"__Schema",
	"__InputValue",
	"__DirectiveLocation",
	"__Field",
}

func metaType(typeName string) bool {
	for _, mt := range metaTypes {
		if typeName == mt {
			return true
		}
	}
	return false
}
