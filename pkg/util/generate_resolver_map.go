package util

import (
	"github.com/solo-io/sqoop/pkg/api/types/v1"
	"github.com/solo-io/sqoop/pkg/exec"
	"github.com/vektah/gqlgen/neelance/schema"
)

func GenerateResolverMapSkeleton(name string, sch *schema.Schema) *v1.ResolverMap {
	types := make(map[string]*v1.TypeResolver)
	for _, t := range sch.Types {
		if exec.MetaType(t.TypeName()) {
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
		Name:  name,
		Types: types,
	}
}
