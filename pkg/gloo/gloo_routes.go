package gloo

import (
	"github.com/solo-io/qloo/pkg/api/types/v1"
	"fmt"
	"sort"
)

func RoutePath(typeName, fieldName string) string {
	return fmt.Sprintf("/%v.%v", typeName, fieldName)
}

func buildRoutes(resolverMap *v1.ResolverMap) []route {
	var routes []route
	for typeName, typeResolver := range resolverMap.Types {
		for fieldName, fieldResolver := range typeResolver.Fields {
			glooResolver, ok := fieldResolver.Resolver.(*v1.Resolver_GlooResolver)
			if !ok {
				continue
			}
			routes = append(routes, route{
				path:         RoutePath(typeName, fieldName),
				destinations: destinationsForFunction(glooResolver.GlooResolver),
			})
		}
	}
	sort.SliceStable(routes, func(i, j int) bool {
		return routes[i].path < routes[j].path
	})
	return routes
}

func destinationsForFunction(resolver *v1.GlooResolver) []destination {
	switch function := resolver.Function.(type) {
	case *v1.GlooResolver_SingleFunction:
		return []destination{
			{
				upstreamName: function.SingleFunction.Upstream,
				functionName: function.SingleFunction.Function,
			},
		}
	case *v1.GlooResolver_MultiFunction:
		var dests []destination
		for _, weightedFunc := range function.MultiFunction.WeightedFunctions {
			dests = append(dests, destination{
				upstreamName: weightedFunc.Function.Upstream,
				functionName: weightedFunc.Function.Function,
				weight:       weightedFunc.Weight,
			})
		}
		return dests
	}
	panic("unknown function time")
}
