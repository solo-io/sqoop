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
)

func main() {
	http.Handle("/", handler.Playground("Starwars", "/query"))
	http.Handle("/query", handler.GraphQL(dynamic.MakeExecutableSchema(test.StarWarsSchema, starWarsResolvers),
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

var starWarsResolvers = map[string]dynamic.ResolverFunc{
	"Query.character": func(params dynamic.Params) (interface{}, error) {
		v, err := baseResolvers.Query_character(context.TODO(), params.Arg("id").(string))
		if err != nil {
			return nil, err
		}
		return fromJson(v), nil
	},
	"Human.friends": func(params dynamic.Params) (interface{}, error) {
		var friends []map[string]interface{}
		ids := params.Source["friendIds"].([]string)
		for _, id := range ids {
			f, err := baseResolvers.Query_character(context.TODO(), id)
			if err != nil {
				return nil, err
			}
			friends = append(friends, fromJson(f))
		}
		return friends, nil
	},
}

// simulate parsing a json response
func fromJson(v interface{}) map[string]interface{} {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		panic(err)
	}
	return m
}
