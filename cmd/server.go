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

func addResolvers(resolvers *dynamic.ResolverMap) {
	resolvers.RegisterResolver("Query", "human", func(params dynamic.Params) ([]byte, error) {
		v, err := baseResolvers.Query_human(context.TODO(), params.Arg("id").(string))
		if err != nil {
			return nil, err
		}
		return json.Marshal(v)
	})
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
