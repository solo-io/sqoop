package translator

import (
	"fmt"
	"sort"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	v1 "github.com/solo-io/sqoop/pkg/api/v1"
)

type route struct {
	path   string
	action *gloov1.RouteAction
}

func RoutePath(resolverMap core.ResourceRef, typeName, fieldName string) string {
	return fmt.Sprintf("/%v.%v.%v.%v", resolverMap.Namespace, resolverMap.Name, typeName, fieldName)
}

func routesForResolverMaps(resolverMaps v1.ResolverMapList, resourceErrs reporter.ResourceErrors) []route {
	var routes []route
	for _, resolverMap := range resolverMaps {
		for typeName, typeResolver := range resolverMap.Types {
			for fieldName, fieldResolver := range typeResolver.Fields {
				glooResolver, ok := fieldResolver.Resolver.(*v1.FieldResolver_GlooResolver)
				if !ok {
					continue
				}
				if glooResolver.GlooResolver == nil {
					resourceErrs.AddError(resolverMap, errors.Errorf("gloo resolver cannot be nil"))
					continue
				}
				if glooResolver.GlooResolver.Action == nil {
					resourceErrs.AddError(resolverMap, errors.Errorf("resolver action cannot be nil"))
					continue
				}
				routes = append(routes, route{
					path:   RoutePath(resolverMap.Metadata.Ref(), typeName, fieldName),
					action: glooResolver.GlooResolver.Action,
				})
			}
		}
	}
	sort.SliceStable(routes, func(i, j int) bool {
		return routes[i].path < routes[j].path
	})
	return routes
}
