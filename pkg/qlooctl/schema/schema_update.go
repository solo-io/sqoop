package schema

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/solo-io/sqoop/pkg/api/types/v1"
	"github.com/solo-io/sqoop/pkg/sqoopctl"
	"github.com/spf13/cobra"
)

var schemaUpdateOpts struct {
	FromFile       string
	UseResolverMap string
}

var schemaUpdateCmd = &cobra.Command{
	Use:   "update NAME --from-file <path/to/your/graphql/schema>",
	Short: "upload a schema to Sqoop from a local GraphQL Schema file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.Errorf("requires exactly 1 argument")
		}
		if err := updateSchema(args[0], schemaUpdateOpts.FromFile, schemaUpdateOpts.UseResolverMap); err != nil {
			return err
		}
		fmt.Println("schema updated successfully")
		return nil
	},
}

func init() {
	schemaUpdateCmd.PersistentFlags().StringVarP(&schemaUpdateOpts.FromFile, "from-file", "f", "", "path to a "+
		"graphql schema file from which to update the Sqoop schema object")
	schemaUpdateCmd.PersistentFlags().StringVarP(&schemaUpdateOpts.UseResolverMap, "resolvermap", "r", "", "The name of a "+
		"ResolverMap to connect to this Schema. If none is specified, an empty ResolverMap will be generated for you, which "+
		"you can then configure with sqoopctl")
	schemaCmd.AddCommand(schemaUpdateCmd)
}

func updateSchema(name, filename, resolvermap string) error {
	cli, err := sqoopctl.MakeClient()
	if err != nil {
		return err
	}
	if name == "" {
		return errors.Errorf("schema name must be set")
	}
	if filename == "" {
		return errors.Errorf("filename must be set")
	}

	existing, err := cli.V1().Schemas().Get(name)
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
		Metadata:     existing.Metadata,
	}
	_, err = cli.V1().Schemas().Update(schema)
	return err
}
