package schema

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/solo-io/qloo/pkg/api/types/v1"
	"github.com/solo-io/qloo/pkg/qlooctl"
	"github.com/spf13/cobra"
)

var schemaCreateOpts struct {
	FromFile       string
	UseResolverMap string
}

var schemaCreateCmd = &cobra.Command{
	Use:   "create NAME --from-file <path/to/your/graphql/schema>",
	Short: "upload a schema to QLoo from a local GraphQL Schema file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.Errorf("requires exactly 1 argument")
		}
		if err := createSchema(args[0], schemaCreateOpts.FromFile, schemaCreateOpts.UseResolverMap); err != nil {
			return err
		}
		fmt.Println("schema created successfully")
		return nil
	},
}

func Init(schemaCmd *cobra.Command) {
	schemaCreateCmd.PersistentFlags().StringVarP(&schemaCreateOpts.FromFile, "from-file", "f", "", "path to a "+
		"graphql schema file from which to create the QLoo schema object")
	schemaCreateCmd.PersistentFlags().StringVarP(&schemaCreateOpts.UseResolverMap, "resolvermap", "r", "", "The name of a "+
		"ResolverMap to connect to this Schema. If none is specified, an empty ResolverMap will be generated for you, which "+
		"you can then configure with qlooctl")
	schemaCmd.AddCommand(schemaCreateCmd)
}

func createSchema(name, filename, resolvermap string) error {
	if name == "" {
		return errors.Errorf("schema name must be set")
	}
	if filename == "" {
		return errors.Errorf("filename must be set")
	}
	cli, err := qlooctl.MakeClient()
	if err != nil {
		return err
	}
	inlineSchemaBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	schema := &v1.Schema{
		Name:         name,
		InlineSchema: string(inlineSchemaBytes),
		ResolverMap:  resolvermap,
	}
	_, err = cli.V1().Schemas().Create(schema)
	return err
}
