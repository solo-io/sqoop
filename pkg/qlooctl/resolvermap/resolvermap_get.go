package resolvermap

import (
	"github.com/solo-io/qloo/pkg/api/types/v1"
	"github.com/solo-io/qloo/pkg/qlooctl"
	"github.com/spf13/cobra"
)

var resolverMapGetCmd = &cobra.Command{
	Use:   "get [NAME]",
	Short: "return a resolver map by its name or list all resolver maps",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 || args[0] == "" {
			list, err := listResolverMaps()
			if err != nil {
				return err
			}
			for _, msg := range list {
				if err := qlooctl.PrintAsYaml(msg); err != nil {
					return err
				}
			}
			return nil
		}
		msg, err := getResolverMap(args[0])
		if err != nil {
			return err
		}
		return qlooctl.PrintAsYaml(msg)
	},
}

func init() {
	resolverMapCmd.AddCommand(resolverMapGetCmd)
}

func getResolverMap(name string) (*v1.ResolverMap, error) {
	cli, err := qlooctl.MakeClient()
	if err != nil {
		return nil, err
	}
	return cli.V1().ResolverMaps().Get(name)
}

func listResolverMaps() ([]*v1.ResolverMap, error) {
	cli, err := qlooctl.MakeClient()
	if err != nil {
		return nil, err
	}
	return cli.V1().ResolverMaps().List()
}
