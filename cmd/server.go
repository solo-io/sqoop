package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/vektah/gqlgen/graphql"
	"github.com/vektah/gqlgen/handler"
	"github.com/solo-io/qloo/pkg/dynamic"
	"github.com/solo-io/qloo/test"
	"github.com/vektah/gqlgen/example/starwars"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/solo-io/gloo/pkg/storage"
	"github.com/solo-io/gloo/pkg/api/types/v1"
	"html/template"
	"bytes"
	"io/ioutil"
)

var starWarsSchema = test.StarWarsSchema

func main() {
	http.Handle("/", handler.Playground("Starwars", "/query"))
	execSchema, resolvers := dynamic.MakeExecutableSchema(starWarsSchema)
	addResolvers(resolvers)
	http.Handle("/query", handler.GraphQL(execSchema,
		handler.ResolverMiddleware(func(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
			rc := graphql.GetResolverContext(ctx)
			fmt.Println("Entered", rc.Object, rc.Field.Name)
			res, err = next(ctx)
			fmt.Println("Left", rc.Object, rc.Field.Name, "=>", res, err)
			return res, err
		}),
	))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

var baseResolvers = starwars.NewResolver()

type UserInput struct {
	TypeToResolve    string
	FieldToResolve   string
	BodyTemplate     string
	ResponseTemplate string
	ContentType      string
}

func addResolvers(resolvers *dynamic.ResolverMap, factory *GlooResolverFactory) {
	resolver, err := factory.Resolver()
	resolvers.RegisterResolver("Query", "hero", resolver)
	resolvers.RegisterResolver("Query", "human", func(params dynamic.Params) ([]byte, error) {
		v, err := baseResolvers.Query_human(context.TODO(), params.Arg("id").(string))
		if err != nil {
			return nil, err
		}
		return json.Marshal(v)
	})
	resolvers.RegisterResolver("Human", "name", func(params dynamic.Params) ([]byte, error) {
		if params.Source == nil {
			return nil, errors.Errorf("source was nil")
		}
		name := params.Source.Data.Get("name").(*dynamic.String).Data
		return []byte(name), nil
	})
	// overriding resolver
	resolvers.RegisterResolver("Human", "appearsIn", func(params dynamic.Params) ([]byte, error) {
		return []byte("[\"EMPIRE\"]"), nil
	})
	resolvers.RegisterResolver("Human", "friends", func(params dynamic.Params) ([]byte, error) {
		fieldVal := params.Source.Data.Get("friendIds").(*dynamic.InternalOnly).Data
		ids := fieldVal.([]interface{})
		var friends []interface{}
		for _, id := range ids {
			v, err := baseResolvers.Query_character(context.TODO(), id.(string))
			if err != nil {
				return nil, err
			}
			friends = append(friends, v)
		}
		return json.Marshal(friends)
	})
	resolvers.RegisterResolver("Droid", "friends", func(params dynamic.Params) ([]byte, error) {
		fieldVal := params.Source.Data.Get("friendIds").(*dynamic.InternalOnly).Data
		ids := fieldVal.([]interface{})
		var friends []interface{}
		for _, id := range ids {
			v, err := baseResolvers.Query_character(context.TODO(), id.(string))
			if err != nil {
				return nil, err
			}
			friends = append(friends, v)
		}
		return json.Marshal(friends)
	})
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
		requestTemplate *template.Template
		responseTemplate *template.Template
		err error
	)

	if requestBodyTemplate != "" {
		requestTemplate, err = template.New("requestBody").Parse(requestBodyTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "parsing request body template failed")
		}
	}
	if responseBodyTemplate != "" {
		responseTemplate, err = template.New("responseBody").Parse(responseBodyTemplate)
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
		if responseTemplate == nil {
			return data, nil
		}

		// requires output to be json object
		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, errors.Wrap(err, "failed to parse response as json object. " +
				"response templates may only be used with JSON responses")
		}
		input := struct{
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
