package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"encoding/json"

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
)

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
		proxyAddr: "localhost:8080",
	}

	storageClient, err := configstorage.Bootstrap(opts)
	if err != nil {
		return err
	}

	client := &GlooClient{
		gloo:           storageClient,
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

func pathName(graphqlType, field string) string {
	return fmt.Sprintf("/%v.%v", graphqlType, field)
}

func addResolvers(resolvers *dynamic.ExecutableResolvers, factory *gloo.ResolverFactory, client *GlooClient, inputs []UserInput) error {
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
			if err := resolvers.SetResolver(in.TypeToResolve, in.FieldToResolve, resolver); err != nil {
				return errors.Wrap(err, "attaching resolver to schema")
			}
		case in.GlooResolverInput != nil:
			path := pathName(in.TypeToResolve, in.FieldToResolve)
			glooInputs := in.GlooResolverInput
			resolver, err := factory.CreateResolver(path, glooInputs.RequestTemplate, glooInputs.ResponseTemplate, glooInputs.ContentType)
			if err != nil {
				return errors.Wrap(err, "generating resolver from inputs")
			}
			if err := resolvers.SetResolver(in.TypeToResolve, in.FieldToResolve, resolver); err != nil {
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
