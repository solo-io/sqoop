package operator

import (
	"github.com/pkg/errors"
	"github.com/solo-io/gloo/pkg/api/types/v1"
	"github.com/solo-io/gloo/pkg/storage"
	qloov1 "github.com/solo-io/qloo/pkg/api/types/v1"
)

const listenerPort = uint32(8080)

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
	// VirtualService
	desiredVirtualService, err := operator.desiredVirtualService(operator.cachedRoutes)
	if err != nil {
		return errors.Wrap(err, "invalid resolver routes")
	}
	existingVirtualService, err := operator.gloo.V1().VirtualServices().Get(operator.virtualServiceName)
	if err != nil {
		if _, err := operator.gloo.V1().VirtualServices().Create(desiredVirtualService); err != nil {
			return err
		}
	} else {
		if !routesEqual(existingVirtualService.Routes, desiredVirtualService.Routes) {
			desiredVirtualService.Metadata.ResourceVersion = existingVirtualService.Metadata.ResourceVersion
			if _, err = operator.gloo.V1().VirtualServices().Update(desiredVirtualService); err != nil {
				return err
			}
		}
	}

	// Role
	desiredRole := operator.desiredRole()
	existingRole, err := operator.gloo.V1().Roles().Get(operator.roleName)
	if err != nil {
		if _, err := operator.gloo.V1().Roles().Create(desiredRole); err != nil {
			return err
		}
	} else {
		if !listenersEqual(existingRole.Listeners, desiredRole.Listeners) {
			desiredRole.Metadata.ResourceVersion = existingRole.Metadata.ResourceVersion
			if _, err = operator.gloo.V1().Roles().Update(desiredRole); err != nil {
				return err
			}
		}
	}

	// clear cache
	operator.cachedRoutes = nil

	return nil
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

func listenersEqual(list1, list2 []*v1.Listener) bool {
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
		Name:               operator.virtualServiceName,
		Domains:            []string{"*"},
		Routes:             routes,
		Metadata:           &v1.Metadata{},
		DisableForGateways: true,
	}, nil
}

func (operator *GlooOperator) desiredRole() *v1.Role {
	return &v1.Role{
		Name: operator.roleName,
		Listeners: []*v1.Listener{
			{
				Name:            "graphql-port",
				BindAddress:     "::",
				BindPort:        listenerPort,
				VirtualServices: []string{operator.virtualServiceName},
			},
		},
		Metadata: &v1.Metadata{},
	}
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
