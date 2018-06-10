package resolvermap

import (
	"github.com/solo-io/qloo/pkg/api/types/v1"
	"github.com/solo-io/qloo/pkg/qlooctl"
	"github.com/spf13/cobra"
	"github.com/pkg/errors"
	"github.com/solo-io/qloo/pkg/storage"
	"github.com/solo-io/qloo/pkg/storage/file"
)

var (
	resolverFile string
	schemaName   string
)

var resolverMapRegisterCmd = &cobra.Command{
	Use:   "register TypeName FieldName -f resolver.yaml [-s schema-name]",
	Short: "Register a resolver for a field in your Schema",
	Long: `Sets the resolver for a field in your schema. TypeName.FieldName will always be resolved using this resolver

Resolvers must be defined in yaml format. See the documentation at https://qloo.solo.io/v1/resolver_map/#qloo.api.v1.Resolver for the API specification for QLoo Resolvers`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 || args[0] == "" || args[1] == "" {
			return errors.Errorf("must specify args TypeName and FieldName")
		}
		if resolverFile == "" {
			return errors.Errorf("must provide the path to a resolver definition yaml file")
		}
		msg, err := registerResolver(schemaName, args[0], args[1], resolverFile)
		if err != nil {
			return err
		}
		return qlooctl.PrintAsYaml(msg)
	},
}

func init() {
	resolverMapRegisterCmd.PersistentFlags().StringVarP(&resolverFile, "file", "f", "", "path to a file "+
		"containing the resolver definition")
	resolverMapRegisterCmd.PersistentFlags().StringVarP(&schemaName, "schema", "s", "", "name of the "+
		"schema to connect this resolver to. this is required if more than one schema contains a definition for the "+
		"type name.")
	resolverMapCmd.AddCommand(resolverMapRegisterCmd)
}

func registerResolver(schemaName, typeName, fieldName, fileName string) (*v1.ResolverMap, error) {
	var resolver v1.Resolver
	if err := file.ReadFileInto(fileName, &resolver); err != nil {
		return nil, err
	}

	cli, err := qlooctl.MakeClient()
	if err != nil {
		return nil, err
	}
	existingResolverMap, err := getResolverMapFor(cli, schemaName, typeName, fieldName)
	if err != nil {
		return nil, err
	}

	existingResolverMap.Types[typeName].Fields[fieldName] = &resolver

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
