package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/vektah/gqlgen/graphql"
	"github.com/vektah/gqlgen/handler"
	"github.com/solo-io/qloo/pkg/dynamic"
	"github.com/solo-io/qloo/test"
	"github.com/pkg/errors"
	"github.com/solo-io/gloo/pkg/storage"
	"github.com/solo-io/gloo/pkg/api/types/v1"
	"github.com/spf13/cobra"
	"github.com/solo-io/gloo/pkg/bootstrap"
	"github.com/solo-io/gloo/pkg/bootstrap/flags"
	"github.com/solo-io/gloo/pkg/bootstrap/configstorage"
	"github.com/solo-io/qloo/pkg/gloo"
	"encoding/json"
)

var inputs = []UserInput{
	{
		TypeToResolve:  "Query",
		FieldToResolve: "hero",
		GlooResolverInput: &GlooResolverInput{
			Destinations: []Destination{
				{
					UpstreamName: "starwars-rest",
					FunctionName: "GetHero",
				},
			},
		},
	},
	{
		TypeToResolve:  "Query",
		FieldToResolve: "human",
		GlooResolverInput: &GlooResolverInput{
			RequestTemplate: `{"id": {{ index .Args "id" }}}`,
			Destinations: []Destination{
				{
					UpstreamName: "starwars-rest",
					FunctionName: "GetCharacter",
				},
			},
		},
	},
	{
		TypeToResolve:  "Human",
		FieldToResolve: "friends",
		GlooResolverInput: &GlooResolverInput{
			RequestTemplate: `{{ marshal (index .Parent "friend_ids") }}`,
			Destinations: []Destination{
				{
					UpstreamName: "starwars-rest",
					FunctionName: "GetCharacters",
				},
			},
		},
	},
	{
		TypeToResolve:  "Droid",
		FieldToResolve: "friends",
		GlooResolverInput: &GlooResolverInput{
			RequestTemplate: `{{ marshal (index .Parent "friend_ids") }}`,
			Destinations: []Destination{
				{
					UpstreamName: "starwars-rest",
					FunctionName: "GetCharacters",
				},
			},
		},
	},
	{
		TypeToResolve:  "Human",
		FieldToResolve: "appearsIn",
		ParentResolverInput: &ParentResolverInput{
			ParentField: "appears_in",
		},
	},
	{
		TypeToResolve:  "Droid",
		FieldToResolve: "appearsIn",
		ParentResolverInput: &ParentResolverInput{
			ParentField: "appears_in",
		},
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var opts bootstrap.Options

var rootCmd = &cobra.Command{
	Use:   "qloo",
	Short: "runs qloo",
	RunE: func(cmd *cobra.Command, args []string) error {
		return run()
	},
}

func init() {
	flags.AddConfigStorageOptionFlags(rootCmd, &opts)
	flags.AddFileFlags(rootCmd, &opts)
}

var starWarsSchema = test.StarWarsSchema

func run() error {
	factory := &gloo.ResolverFactory{
		ProxyAddr: "localhost:8080",
	}

	gloo, err := configstorage.Bootstrap(opts)
	if err != nil {
		return err
	}

	client := &GlooClient{
		gloo:           gloo,
		virtualService: "qloo",
		role:           "qloo",
	}

	execSchema, resolvers := dynamic.MakeExecutableSchema(starWarsSchema)

	if err := addResolvers(resolvers, factory, client, inputs); err != nil {
		return errors.Wrap(err, "failed to start")
	}
	http.Handle("/", handler.Playground("Starwars", "/query"))
	http.Handle("/query", handler.GraphQL(execSchema,
		handler.ResolverMiddleware(func(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
			rc := graphql.GetResolverContext(ctx)
			fmt.Println("Entered", rc.Object, rc.Field.Name)
			res, err = next(ctx)
			fmt.Println("Left", rc.Object, rc.Field.Name, "=>", res, err)
			return res, err
		}),
	))

	return http.ListenAndServe(":9090", nil)
}

type UserInput struct {
	TypeToResolve       string
	FieldToResolve      string
	GlooResolverInput   *GlooResolverInput
	ParentResolverInput *ParentResolverInput
}

type GlooResolverInput struct {
	RequestTemplate  string
	ResponseTemplate string
	ContentType      string
	Destinations     []Destination
}

type ParentResolverInput struct {
	ParentField string
}

func pathName(graphqlType, field string) string {
	return fmt.Sprintf("/%v.%v", graphqlType, field)
}

func addResolvers(resolvers *dynamic.ResolverMap, factory *gloo.ResolverFactory, client *GlooClient, inputs []UserInput) error {
	var glooRoutes []Route
	for _, in := range inputs {
		switch {
		case in.ParentResolverInput != nil:
			resolver := func(params dynamic.Params) ([]byte, error) {
				if params.Parent == nil {
					return nil, errors.Errorf("no parent to lookup field")
				}
				m, isMap := params.Parent.GoValue().(map[string]interface{})
				if !isMap {
					return nil, errors.Errorf("parent was not an object")
				}
				v, ok := m[in.ParentResolverInput.ParentField]
				if !ok {
					return nil, errors.Errorf("filed %v not found in parent", in.ParentResolverInput.ParentField)
				}
				return json.Marshal(v)
			}
			if err := resolvers.RegisterResolver(in.TypeToResolve, in.FieldToResolve, resolver); err != nil {
				return errors.Wrap(err, "attaching resolver to schema")
			}
		case in.GlooResolverInput != nil:
			path := pathName(in.TypeToResolve, in.FieldToResolve)
			glooInputs := in.GlooResolverInput
			resolver, err := factory.MakeResolver(path, glooInputs.RequestTemplate, glooInputs.ResponseTemplate, glooInputs.ContentType)
			if err != nil {
				return errors.Wrap(err, "generating resolver from inputs")
			}
			if err := resolvers.RegisterResolver(in.TypeToResolve, in.FieldToResolve, resolver); err != nil {
				return errors.Wrap(err, "attaching resolver to schema")
			}
			glooRoutes = append(glooRoutes, Route{
				Path:         path,
				Destinations: glooInputs.Destinations,
			})
		}
	}
	return client.SyncVirtualService(glooRoutes)
}

type GlooClient struct {
	gloo           storage.Interface
	virtualService string
	role           string
}

type Route struct {
	Path         string
	Destinations []Destination
}

type Destination struct {
	UpstreamName, FunctionName string
	Weight                     uint32
}

func (client *GlooClient) SyncVirtualService(resolverRoutes []Route) error {
	desiredVirtualService, err := client.desiredVirtualService(resolverRoutes)
	if err != nil {
		return errors.Wrap(err, "invalid resolver routes")
	}
	existingVirtualService, err := client.gloo.V1().VirtualServices().Get(client.virtualService)
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

func (client *GlooClient) desiredVirtualService(resolverRoutes []Route) (*v1.VirtualService, error) {
	var routes []*v1.Route
	for _, rr := range resolverRoutes {
		route, err := resolverRoute(rr)
		if err != nil {
			return nil, errors.Wrap(err, "creating route for resolver")
		}
		routes = append(routes, route)
	}
	return &v1.VirtualService{
		Name:     client.virtualService,
		Domains:  []string{"*"},
		Routes:   routes,
		Roles:    []string{client.role},
		Metadata: &v1.Metadata{},
	}, nil
}

func resolverRoute(route Route) (*v1.Route, error) {
	if len(route.Destinations) == 0 {
		return nil, errors.Errorf("need at least 1 destination to create a resolver route")
	}

	var (
		singleDestination *v1.Destination
		multiDestination  []*v1.WeightedDestination
	)
	destinations := route.Destinations

	switch {
	case len(destinations) == 1:
		singleDestination = &v1.Destination{
			DestinationType: &v1.Destination_Function{
				Function: &v1.FunctionDestination{
					UpstreamName: destinations[0].UpstreamName,
					FunctionName: destinations[0].FunctionName,
				},
			},
		}
	case len(destinations) > 1:
		for _, dest := range destinations {
			multiDestination = append(multiDestination, &v1.WeightedDestination{
				Destination: &v1.Destination{
					DestinationType: &v1.Destination_Function{
						Function: &v1.FunctionDestination{
							UpstreamName: dest.UpstreamName,
							FunctionName: dest.FunctionName,
						},
					},
				},
				Weight: dest.Weight,
			})
		}
	}

	return &v1.Route{
		Matcher: &v1.Route_RequestMatcher{
			RequestMatcher: &v1.RequestMatcher{
				Path: &v1.RequestMatcher_PathExact{
					PathExact: route.Path,
				},
				Verbs: []string{"POST"},
			},
		},
		MultipleDestinations: multiDestination,
		SingleDestination:    singleDestination,
	}, nil
}
