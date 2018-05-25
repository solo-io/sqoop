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
)

func main() {
	http.Handle("/", handler.Playground("Starwars", "/query"))
	http.Handle("/query", handler.GraphQL(dynamic.MakeExecutableSchema(test.StarWarsSchema),
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
