package translator

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/sqoop/pkg/api/v1"
	"github.com/solo-io/sqoop/pkg/engine/exec"
	"github.com/vektah/gqlgen/neelance/schema"
)

// TODO(ilackarms)
func GenerateResolverMapSkeleton(meta core.Metadata, sch *schema.Schema) *v1.ResolverMap {
	types := make(map[string]*v1.TypeResolver)
	for _, t := range sch.Types {
		if exec.MetaType(t.TypeName()) {
			continue
		}
		fields := make(map[string]*v1.FieldResolver)
		switch t := t.(type) {
		case *schema.Object:
			for _, f := range t.Fields {
				fields[f.Name] = &v1.FieldResolver{
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
		Metadata: meta,
		Types:    types,
	}
}
