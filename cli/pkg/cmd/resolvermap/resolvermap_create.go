package resolvermap

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
	"sigs.k8s.io/yaml"
)

func Create(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create NAME --from-file <path/to/your/sqoop/resolver map>",
		Short: "upload a resolver map to Sqoop from a local Sqoop ResolverMap yaml file",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := common.SetResourceName(&opts.Metadata, args)
			if err != nil {
				return err
			}
			if err := createResolverMap(opts); err != nil {
				return err
			}
			fmt.Println("resolver map created successfully")
			return nil
		},
		Args: common.RequiredNameArg,
	}

	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func createResolverMap(opts *options.Options) error {
	if opts.Top.File == "" {
		return fmt.Errorf("resolver map file must be set")
	}
	client, err := helpers.ResolverMapClient()
	if err != nil {
		return err
	}
	var resolverMap v1.ResolverMap
	resolverMapBytes, err := ioutil.ReadFile(opts.Top.File)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(resolverMapBytes, &resolverMap); err != nil {
		return err
	}

	_, err = client.Write(&resolverMap, clients.WriteOpts{})
	return err
}
