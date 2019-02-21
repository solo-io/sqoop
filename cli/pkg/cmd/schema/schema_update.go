package schema

import (
	"fmt"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/sqoop/cli/pkg/common"
	"github.com/solo-io/sqoop/cli/pkg/helpers"
	"github.com/solo-io/sqoop/cli/pkg/options"
	"github.com/solo-io/sqoop/pkg/api/v1"
	"github.com/spf13/cobra"
	"io/ioutil"
)

func Update(opts *options.Options, optionsFunc... cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update NAME -f <path/to/your/graphql/schema>",
		Short: "upload a schema to Sqoop from a local GraphQL Schema file",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := common.SetResourceName(&opts.Metadata, args)
			if err != nil {
				return err
			}
			if err := updateSchema(opts); err != nil {
				return err
			}
			fmt.Println("schema updated successfully")
			return nil
		},
	}


	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func updateSchema(opts *options.Options) error {
	if opts.Top.File == "" {
		return fmt.Errorf("schema file must be set")
	}
	client, err := helpers.SchemaClient()
	if err != nil {
		return err
	}
	existing, err := client.Read(opts.Metadata.Namespace, opts.Metadata.Name, clients.ReadOpts{})
	if err != nil {
		return err
	}
	inlineSchemaBytes, err := ioutil.ReadFile(opts.Top.File)
	if err != nil {
		return err
	}
	schema := &v1.Schema{
		Metadata: existing.Metadata,
		InlineSchema: string(inlineSchemaBytes),
	}

	_, err = client.Write(schema, clients.WriteOpts{OverwriteExisting: true})
	return err
}