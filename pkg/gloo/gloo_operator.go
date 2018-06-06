package gloo

import (
	"github.com/solo-io/gloo/pkg/storage"
	"github.com/pkg/errors"
	"github.com/solo-io/gloo/pkg/api/types/v1"
	qloov1 "github.com/solo-io/qloo/pkg/api/types/v1"
)

type GlooOperator struct {
	gloo               storage.Interface
	virtualServiceName string
	roleName           string
}

func NewGlooOperator(gloo storage.Interface, virtualServiceName string, roleName string) *GlooOperator {
	return &GlooOperator{
		gloo:               gloo,
		virtualServiceName: virtualServiceName,
		roleName:           roleName,
	}
}

type route struct {
	path         string
	destinations []destination
}

type destination struct {
	upstreamName, functionName string
	weight                     uint32
}

func (client *GlooOperator) ApplyResolvers(resolverMap *qloov1.ResolverMap) error {
	routes := buildRoutes(resolverMap)
	desiredVirtualService, err := client.desiredVirtualService(routes)
	if err != nil {
		return errors.Wrap(err, "invalid resolver routes")
	}
	existingVirtualService, err := client.gloo.V1().VirtualServices().Get(client.virtualServiceName)
	if err != nil {
		_, err := client.gloo.V1().VirtualServices().Create(desiredVirtualService)
		return err
	}
	if routesEqual(existingVirtualService.Routes, desiredVirtualService.Routes) {
		return nil
	}
	desiredVirtualService.Metadata.ResourceVersion = existingVirtualService.Metadata.ResourceVersion
	_, err = client.gloo.V1().VirtualServices().Update(desiredVirtualService)
	return err
}

func routesEqual(list1, list2 []*v1.Route) bool {
	if len(list1) != len(list2) {
		return false
	}
	for i := range list1 {
		r1, r2 := list1[i], list2[i]
		if !r1.Equal(r2) {
			return false
		}
	}
	return true
}

func (client *GlooOperator) desiredVirtualService(resolverRoutes []route) (*v1.VirtualService, error) {
	var routes []*v1.Route
	for _, rr := range resolverRoutes {
		route, err := resolverRoute(rr)
		if err != nil {
			return nil, errors.Wrap(err, "creating route for resolver")
		}
		routes = append(routes, route)
	}
	return &v1.VirtualService{
		Name:     client.virtualServiceName,
		Domains:  []string{"*"},
		Routes:   routes,
		Roles:    []string{client.roleName},
		Metadata: &v1.Metadata{},
	}, nil
}

func resolverRoute(route route) (*v1.Route, error) {
	if len(route.destinations) == 0 {
		return nil, errors.Errorf("need at least 1 destination to create a resolver route")
	}

	var (
		singleDestination *v1.Destination
		multiDestination  []*v1.WeightedDestination
	)
	destinations := route.destinations

	switch {
	case len(destinations) == 1:
		singleDestination = &v1.Destination{
			DestinationType: &v1.Destination_Function{
				Function: &v1.FunctionDestination{
					UpstreamName: destinations[0].upstreamName,
					FunctionName: destinations[0].functionName,
				},
			},
		}
	case len(destinations) > 1:
		for _, dest := range destinations {
			multiDestination = append(multiDestination, &v1.WeightedDestination{
				Destination: &v1.Destination{
					DestinationType: &v1.Destination_Function{
						Function: &v1.FunctionDestination{
							UpstreamName: dest.upstreamName,
							FunctionName: dest.functionName,
						},
					},
				},
				Weight: dest.weight,
			})
		}
	}

	return &v1.Route{
		Matcher: &v1.Route_RequestMatcher{
			RequestMatcher: &v1.RequestMatcher{
				Path: &v1.RequestMatcher_PathExact{
					PathExact: route.path,
				},
				Verbs: []string{"POST"},
			},
		},
		MultipleDestinations: multiDestination,
		SingleDestination:    singleDestination,
	}, nil
}
