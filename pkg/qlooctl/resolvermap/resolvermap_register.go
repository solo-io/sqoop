package resolvermap

import (
	"github.com/pkg/errors"
	"github.com/solo-io/sqoop/pkg/api/types/v1"
	"github.com/solo-io/sqoop/pkg/sqoopctl"
	"github.com/solo-io/sqoop/pkg/storage"
	"github.com/spf13/cobra"
)

var (
	schemaName       string
	functionName     string
	upstreamName     string
	requestTemplate  string
	responseTemplate string
)

var resolverMapRegisterCmd = &cobra.Command{
	Use:   "register TypeName FieldName -f resolver.yaml [-s schema-name]",
	Short: "Register a resolver for a field in your Schema",
	Long: `Sets the resolver for a field in your schema. TypeName.FieldName will always be resolved using this resolver

Resolvers must be defined in yaml format. See the documentation at https://sqoop.solo.io/v1/resolver_map/#sqoop.api.v1.Resolver for the API specification for Sqoop Resolvers`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 || args[0] == "" || args[1] == "" {
			return errors.Errorf("must specify args TypeName and FieldName")
		}
		if upstreamName == "" || functionName == "" {
			return errors.Errorf("must provide an upstream and function to create a resolver")
		}
		msg, err := registerResolver(schemaName, args[0], args[1], upstreamName, functionName, requestTemplate, responseTemplate)
		if err != nil {
			return err
		}
		return sqoopctl.Print(msg)
	},
}

func init() {
	resolverMapRegisterCmd.PersistentFlags().StringVarP(&upstreamName, "upstream", "u", "", "upstream where the function lives")
	resolverMapRegisterCmd.PersistentFlags().StringVarP(&functionName, "function", "f", "", "function to use as resolver")
	resolverMapRegisterCmd.PersistentFlags().StringVarP(&requestTemplate, "request-template", "b", "", "template to use for the request body")
	resolverMapRegisterCmd.PersistentFlags().StringVarP(&responseTemplate, "response-template", "r", "", "template to use for the response body")
	resolverMapRegisterCmd.PersistentFlags().StringVarP(&schemaName, "schema", "s", "", "name of the "+
		"schema to connect this resolver to. this is required if more than one schema contains a definition for the "+
		"type name.")
	resolverMapCmd.AddCommand(resolverMapRegisterCmd)
}

func registerResolver(schemaName, typeName, fieldName, upstreamName, functionName, requestTemplate, responseTemplate string) (*v1.ResolverMap, error) {
	resolver := &v1.Resolver{
		Resolver: &v1.Resolver_GlooResolver{
			GlooResolver: &v1.GlooResolver{
				RequestTemplate:  requestTemplate,
				ResponseTemplate: responseTemplate,
				Function: &v1.GlooResolver_SingleFunction{
					SingleFunction: &v1.Function{
						Upstream: upstreamName,
						Function: functionName,
					},
				},
			},
		},
	}
	cli, err := sqoopctl.MakeClient()
	if err != nil {
		return nil, err
	}
	existingResolverMap, err := getResolverMapFor(cli, schemaName, typeName, fieldName)
	if err != nil {
		return nil, err
	}

	existingResolverMap.Types[typeName].Fields[fieldName] = resolver

	return cli.V1().ResolverMaps().Update(existingResolverMap)
}

func getResolverMapFor(cli storage.Interface, schemaName, typeName, fieldName string) (*v1.ResolverMap, error) {
	if schemaName != "" {
		schema, err := cli.V1().Schemas().Get(schemaName)
		if err != nil {
			return nil, err
		}
		if schema.ResolverMap == "" {
			return nil, errors.Errorf("schema %v does not have a resolver map defined", schemaName)
		}
		return cli.V1().ResolverMaps().Get(schema.ResolverMap)
	}
	resolverMaps, err := cli.V1().ResolverMaps().List()
	if err != nil {
		return nil, err
	}
	for _, rm := range resolverMaps {
		typResolver, ok := rm.Types[typeName]
		if !ok {
			continue
		}
		if _, ok := typResolver.Fields[fieldName]; !ok {
			continue
		}
		return rm, nil
	}
	return nil, errors.Errorf("cannot find a resolver map for type %v with field %v", typeName, fieldName)
}
