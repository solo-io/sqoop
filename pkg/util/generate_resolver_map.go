package util

import (
	"github.com/vektah/gqlgen/neelance/schema"
	"github.com/solo-io/qloo/pkg/api/types/v1"
)

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

func MetaType(typeName string) bool {
	for _, mt := range metaTypes {
		if typeName == mt {
			return true
		}
	}
	return false
}

func GenerateResolverMapSkeleton(sch *schema.Schema) (*v1.ResolverMap) {
	types := make(map[string]*v1.TypeResolver)
	for _, t := range sch.Types {
		if MetaType(t.TypeName()) {
			continue
		}
		fields := make(map[string]*v1.Resolver)
		switch t := t.(type) {
		case *schema.Object:
			for _, f := range t.Fields {
				fields[f.Name] = &v1.Resolver{
					Resolver: nil,
				}
			}
		}
		if len(fields) == 0 {
			continue
		}
		types[t.TypeName()] = &v1.TypeResolver{Fields: fields}
	}
	return &v1.ResolverMap{
		Types: types,
	}
}
