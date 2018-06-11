package operator

import (
	"github.com/pkg/errors"
	"github.com/solo-io/gloo/pkg/api/types/v1"
	"github.com/solo-io/gloo/pkg/storage"
	qloov1 "github.com/solo-io/qloo/pkg/api/types/v1"
)

type GlooOperator struct {
	gloo               storage.Interface
	virtualServiceName string
	roleName           string
	cachedRoutes       []route
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

// apply routes to Gloo and clear cache
func (operator *GlooOperator) ConfigureGloo() error {
	desiredVirtualService, err := operator.desiredVirtualService(operator.cachedRoutes)
	if err != nil {
		return errors.Wrap(err, "invalid resolver routes")
	}
	existingVirtualService, err := operator.gloo.V1().VirtualServices().Get(operator.virtualServiceName)
	if err != nil {
		_, err := operator.gloo.V1().VirtualServices().Create(desiredVirtualService)
		return err
	}
	if routesEqual(existingVirtualService.Routes, desiredVirtualService.Routes) {
		return nil
	}
	desiredVirtualService.Metadata.ResourceVersion = existingVirtualService.Metadata.ResourceVersion
	_, err = operator.gloo.V1().VirtualServices().Update(desiredVirtualService)

	// clear cache
	operator.cachedRoutes = nil

	return err
}

func (operator *GlooOperator) ApplyResolvers(resolverMap *qloov1.ResolverMap) {
	operator.cachedRoutes = append(operator.cachedRoutes, buildRoutes(resolverMap)...)
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

func (operator *GlooOperator) desiredVirtualService(resolverRoutes []route) (*v1.VirtualService, error) {
	var routes []*v1.Route
	for _, rr := range resolverRoutes {
		route, err := resolverRoute(rr)
		if err != nil {
			return nil, errors.Wrap(err, "creating route for resolver")
		}
		routes = append(routes, route)
	}
	return &v1.VirtualService{
		Name:     operator.virtualServiceName,
		Domains:  []string{"*"},
		Routes:   routes,
		Roles:    []string{operator.roleName},
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
