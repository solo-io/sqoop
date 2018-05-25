package dynamic

import (
	"github.com/vektah/gqlgen/neelance/schema"
)

// store all the user resolvers
type ResolverMap struct {
	// resolvers by type
	Types map[string]*TypeResolver
}

type TypeResolver struct {
	// resolve each field of the type
	Fields map[string]ResolverFunc
}

// todo
type ResolverFunc func(args map[string]interface{}) (interface{}, error)

func NewResolverMap(sch *schema.Schema, inputResolvers map[string]ResolverFunc) *ResolverMap {
	typeMap := make(map[string]*TypeResolver)
	for _, t := range sch.Types {
		if metaType(t.TypeName()) {
			continue
		}
		fields := make(map[string]ResolverFunc)
		switch t := t.(type) {
		case *schema.Object:
			for _, f := range t.Fields {
				res, ok := inputResolvers[t.Name+"."+f.Name]
				if ok {
					fields[f.Name] = res
				}
			}
		case *schema.Interface:
			for _, f := range t.Fields {
				res, ok := inputResolvers[t.Name+"."+f.Name]
				if ok {
					fields[f.Name] = res
				}
			}
		case *schema.Union:
			for _, o := range t.PossibleTypes {
				res, ok := inputResolvers[t.Name+"."+o.Name]
				if ok {
					fields[o.Name] = res
				}
			}
		}
		if len(fields) == 0 {
			continue
		}
		typeMap[t.TypeName()] = &TypeResolver{Fields: fields}
	}
	return &ResolverMap{
		Types: typeMap,
	}
}

func (rm *ResolverMap) GetResolver(typeName, field string) ResolverFunc {
	typeResolver, ok := rm.Types[typeName]
	if !ok {
		return emptyResolver
	}
	fieldResolver, ok := typeResolver.Fields[field]
	if !ok {
		return emptyResolver
	}
	return fieldResolver
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
