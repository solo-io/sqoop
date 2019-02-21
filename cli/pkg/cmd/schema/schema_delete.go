package schema

import (
	"fmt"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/sqoop/cli/pkg/common"
	"github.com/solo-io/sqoop/cli/pkg/helpers"
	"github.com/solo-io/sqoop/cli/pkg/options"
	"github.com/spf13/cobra"
)

func Delete(opts *options.Options, optionsFunc... cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [NAME]",
		Short: "delete a schema by its name",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := common.SetResourceName(&opts.Metadata, args)
			if err != nil {
				return err
			}
			if err := deleteSchema(opts); err != nil {
				return err
			}
			fmt.Println("schema deleted successfully")
			return nil
		},
		Args: common.RequiredNameArg,
	}


	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func deleteSchema(opts *options.Options) error {
	client, err := helpers.SchemaClient()
	if err != nil {
		return err
	}
	return client.Delete(opts.Metadata.Namespace, opts.Metadata.Name, clients.DeleteOpts{})
}