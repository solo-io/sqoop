package dynamic

import (
	"github.com/vektah/gqlgen/neelance/schema"
	"github.com/vektah/gqlgen/neelance/common"
)

// store all the user resolvers
type ResolverMap struct {
	// resolvers for all named types
	Types map[schema.NamedType]TypeResolver
}

type TypeResolver struct {
	// resolve each field of the type
	Fields map[string]FieldResolver
}

type FieldResolver struct {
	// type the field resolves to
	Type common.Type
	// how to resolve this field. should return Type
	ResolverFunc ResolverFunc
}

// todo
type ResolverFunc func(args map[string]interface{}) (interface{}, error)

func NewResolverMap(sch *schema.Schema, inputResolvers map[string]ResolverFunc) *ResolverMap {
	typeMap := make(map[schema.NamedType]TypeResolver)
	for _, t := range sch.Types {
		if metaType(t.TypeName()) {
			continue
		}
		fields := make(map[string]FieldResolver)
		switch t := t.(type) {
		case *schema.Object:
			for _, f := range t.Fields {
				res := inputResolvers[t.Name+"."+f.Name]
				fields[f.Name] = FieldResolver{Type: f.Type, ResolverFunc: res}
			}
		case *schema.Interface:
			for _, f := range t.Fields {
				res := inputResolvers[t.Name+"."+f.Name]
				fields[f.Name] = FieldResolver{Type: f.Type, ResolverFunc: res}
			}
		case *schema.Union:
			for _, o := range t.PossibleTypes {
				res := inputResolvers[t.Name+"."+o.Name]
				fields[o.Name] = FieldResolver{Type: o, ResolverFunc: res}
			}
		}
		if len(fields) == 0 {
			continue
		}
		typeMap[t] = TypeResolver{Fields: fields}
	}
	return &ResolverMap{
		Types: typeMap,
	}
}

func (rm *ResolverMap) Resolve(typ schema.NamedType, field string, args map[string]interface{}) (interface{}, error) {
	return rm.getFieldResolver(typ, field)(args)
}

func (rm *ResolverMap) getFieldResolver(typ schema.NamedType, field string) ResolverFunc {
	typeResolver, ok := rm.Types[typ]
	if !ok {
		return emptyResolver
	}
	fieldResolver, ok := typeResolver.Fields[field]
	if !ok {
		return emptyResolver
	}
	return fieldResolver.ResolverFunc
}

func emptyResolver(args map[string]interface{}) (interface{}, error) {
	return nil, nil
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
