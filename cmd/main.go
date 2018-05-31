package main

import (
	"context"
	"fmt"
	"net/http"
	"bytes"
	"io/ioutil"
	"os"

	"github.com/vektah/gqlgen/graphql"
	"github.com/vektah/gqlgen/handler"
	"github.com/solo-io/qloo/pkg/dynamic"
	"github.com/solo-io/qloo/test"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/solo-io/gloo/pkg/storage"
	"github.com/solo-io/gloo/pkg/api/types/v1"
	"html/template"
	"github.com/spf13/cobra"
	"github.com/solo-io/gloo/pkg/bootstrap"
	"github.com/solo-io/gloo/pkg/bootstrap/flags"
	"github.com/solo-io/gloo/pkg/bootstrap/configstorage"
)

var starWarsSchema = test.StarWarsSchema

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

func run() error {
	factory := &GlooResolverFactory{
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

	return http.ListenAndServe(":8080", nil)
}

type UserInput struct {
	TypeToResolve    string
	FieldToResolve   string
	RequestTemplate  string
	ResponseTemplate string
	ContentType      string
	Destinations     []Destination
}

var inputs = []UserInput{
	{
		TypeToResolve:  "Query",
		FieldToResolve: "hero",
		Destinations: []Destination{
			{
				UpstreamName: "myupstream",
				FunctionName: "getHero",
			},
		},
	},
	{
		TypeToResolve:   "Query",
		FieldToResolve:  "human",
		RequestTemplate: `{"id": {{ .Args["id"] }}}`,
		Destinations: []Destination{
			{
				UpstreamName: "myupstream",
				FunctionName: "getHuman",
			},
		},
	},
	{
		TypeToResolve:   "Human",
		FieldToResolve:  "friends",
		RequestTemplate: `{{ marshal .Source["friendIds"] }}`,
		Destinations: []Destination{
			{
				UpstreamName: "myupstream",
				FunctionName: "getHumanFriends",
			},
		},
	},
	{
		TypeToResolve:   "Droid",
		FieldToResolve:  "friends",
		RequestTemplate: `{{ marshal .Source["friendIds"] }}`,
		Destinations: []Destination{
			{
				UpstreamName: "myupstream",
				FunctionName: "getDroidFriends",
			},
		},
	},
}

func pathName(graphqlType, field string) string {
	return fmt.Sprintf("/%v.%v", graphqlType, field)
}

func addResolvers(resolvers *dynamic.ResolverMap, factory *GlooResolverFactory, client *GlooClient, inputs []UserInput) error {
	var glooRoutes []Route
	for _, in := range inputs {
		path := pathName(in.TypeToResolve, in.FieldToResolve)
		resolver, err := factory.Resolver(path, in.RequestTemplate, in.ResponseTemplate, in.ContentType)
		if err != nil {
			return errors.Wrap(err, "generating resolver from inputs")
		}
		if err := resolvers.RegisterResolver(in.TypeToResolve, in.FieldToResolve, resolver); err != nil {
			return errors.Wrap(err, "attaching resolver to schema")
		}
		glooRoutes = append(glooRoutes, Route{
			Path:         path,
			Destinations: in.Destinations,
		})
	}
	return client.SyncVirtualService(glooRoutes)

	//resolvers.RegisterResolver("Query", "hero", factory.MustResolver("/Query.hero", ))
	//resolvers.RegisterResolver("Query", "human", func(params dynamic.Params) ([]byte, error) {
	//	v, err := baseResolvers.Query_human(context.TODO(), params.Arg("id").(string))
	//	if err != nil {
	//		return nil, err
	//	}
	//	return json.Marshal(v)
	//})
	//resolvers.RegisterResolver("Human", "name", func(params dynamic.Params) ([]byte, error) {
	//	if params.Source == nil {
	//		return nil, errors.Errorf("source was nil")
	//	}
	//	name := params.Source.Data.Get("name").(*dynamic.String).Data
	//	return []byte(name), nil
	//})
	//// overriding resolver
	//resolvers.RegisterResolver("Human", "appearsIn", func(params dynamic.Params) ([]byte, error) {
	//	return []byte("[\"EMPIRE\"]"), nil
	//})
	//resolvers.RegisterResolver("Human", "friends", func(params dynamic.Params) ([]byte, error) {
	//	fieldVal := params.Source.Data.Get("friendIds").(*dynamic.InternalOnly).Data
	//	ids := fieldVal.([]interface{})
	//	var friends []interface{}
	//	for _, id := range ids {
	//		v, err := baseResolvers.Query_character(context.TODO(), id.(string))
	//		if err != nil {
	//			return nil, err
	//		}
	//		friends = append(friends, v)
	//	}
	//	return json.Marshal(friends)
	//})
	//resolvers.RegisterResolver("Droid", "friends", func(params dynamic.Params) ([]byte, error) {
	//	fieldVal := params.Source.Data.Get("friendIds").(*dynamic.InternalOnly).Data
	//	ids := fieldVal.([]interface{})
	//	var friends []interface{}
	//	for _, id := range ids {
	//		v, err := baseResolvers.Query_character(context.TODO(), id.(string))
	//		if err != nil {
	//			return nil, err
	//		}
	//		friends = append(friends, v)
	//	}
	//	return json.Marshal(friends)
	//})
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

type GlooResolverFactory struct {
	ProxyAddr string
}

func (gr *GlooResolverFactory) Resolver(path, requestBodyTemplate, responseBodyTemplate, contentType string) (dynamic.RawResolver, error) {
	if contentType == "" {
		contentType = "application/json"
	}
	var (
		requestTemplate  *template.Template
		responseTemplate *template.Template
		err              error
	)

	if requestBodyTemplate != "" {
		requestTemplate, err = template.New("requestBody").Funcs(template.FuncMap{
			"marshal": func(v interface{}) (template.JS, error) {
				a, err := json.Marshal(v)
				if err != nil {
					return "", err
				}
				return template.JS(a), nil
			},
		}).Parse(requestBodyTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "parsing request body template failed")
		}
	}
	if responseBodyTemplate != "" {
		responseTemplate, err = template.New("responseBody").Funcs(template.FuncMap{
			"marshal": func(v interface{}) (template.JS, error) {
				a, err := json.Marshal(v)
				if err != nil {
					return "", err
				}
				return template.JS(a), nil
			},
		}).Parse(responseBodyTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "parsing response body template failed")
		}
	}

	return func(params dynamic.Params) ([]byte, error) {
		body := &bytes.Buffer{}
		if requestTemplate != nil {
			if err := requestTemplate.Execute(body, params); err != nil {
				// TODO: sanitize
				return nil, errors.Wrapf(err, "executing request template for params %v", params)
			}
		}
		url := "http://" + gr.ProxyAddr + path
		res, err := http.Post(url, contentType, body)
		if err != nil {
			return nil, errors.Wrap(err, "performing http post")
		}
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			return nil, errors.Errorf("unexpected status code: %v", res.StatusCode)
		}
		// empty response
		if res.Body == nil {
			return nil, nil
		}
		defer res.Body.Close()
		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, errors.Wrap(err, "reading response body")
		}

		// no template, return raw
		if responseTemplate == nil {
			return data, nil
		}

		// requires output to be json object
		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, errors.Wrap(err, "failed to parse response as json object. "+
				"response templates may only be used with JSON responses")
		}
		input := struct {
			Result map[string]interface{}
		}{
			Result: result,
		}
		buf := &bytes.Buffer{}
		if err := requestTemplate.Execute(buf, input); err != nil {
			return nil, errors.Wrapf(err, "executing response template for response %v", input)
		}
		return buf.Bytes(), nil
	}, nil
}
