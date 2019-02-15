package resolvermap

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
		Short: "delete a resolver map by its name",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := common.SetResourceName(&opts.Metadata, args)
			if err != nil {
				return err
			}
			if err := deleteResolverMap(opts); err != nil {
				return err
			}
			fmt.Println("resolvermap deleted successfully")
			return nil
		},
		Args: common.RequiredNameArg,
	}


	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func deleteResolverMap(opts *options.Options) error {
	client, err := helpers.ResolverMapClient()
	if err != nil {
		return err
	}
	return client.Delete(opts.Metadata.Namespace, opts.Metadata.Name, clients.DeleteOpts{})
}