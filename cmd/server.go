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
			v, err := baseResolvers.Query_human(context.TODO(), id.(string))
			if err != nil {
				return nil, err
			}
			friends = append(friends, v)
		}
		return json.Marshal(friends)
	})
}
