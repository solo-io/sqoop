package schema

import (
	"fmt"
	"io/ioutil"

	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/sqoop/cli/pkg/common"
	"github.com/solo-io/sqoop/cli/pkg/helpers"
	"github.com/solo-io/sqoop/cli/pkg/options"
	v1 "github.com/solo-io/sqoop/pkg/api/v1"
	"github.com/spf13/cobra"
)

func Create(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create NAME -f <path/to/your/graphql/schema>",
		Short: "upload a schema to Sqoop from a local GraphQL Schema file",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := common.SetResourceName(&opts.Metadata, args)
			if err != nil {
				return err
			}
			if err := createSchema(opts); err != nil {
				return err
			}
			fmt.Println("schema created successfully")
			return nil
		},
		Args: common.RequiredNameArg,
	}

	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func createSchema(opts *options.Options) error {
	if opts.Top.File == "" {
		return fmt.Errorf("schema file must be set")
	}
	client, err := helpers.SchemaClient()
	if err != nil {
		return err
	}
	inlineSchemaBytes, err := ioutil.ReadFile(opts.Top.File)
	if err != nil {
		return err
	}
	schema := &v1.Schema{
		Metadata:     opts.Metadata,
		InlineSchema: string(inlineSchemaBytes),
	}

	_, err = client.Write(schema, clients.WriteOpts{})
	return err
}
