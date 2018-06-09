package resolvermap

import (
	"github.com/spf13/cobra"
	"github.com/solo-io/qloo/pkg/api/types/v1"
	"github.com/pkg/errors"
	"fmt"
	"github.com/solo-io/qloo/pkg/storage/file"
	"github.com/solo-io/qloo/pkg/qlooctl"
)

var resolverMapUpdateOpts struct {
	FromFile string
}

var resolverMapUpdateCmd = &cobra.Command{
	Use:   "update NAME --from-file <path/to/your/qloo/resolver map>",
	Short: "upload a resolver map to QLoo from a local QLoo ResolverMap yaml file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.Errorf("requires exactly 1 argument")
		}
		if err := updateResolverMap(args[0], resolverMapUpdateOpts.FromFile); err != nil {
			return err
		}
		fmt.Println("resolver map updated successfully")
		return nil
	},
}

func init() {
	resolverMapUpdateCmd.PersistentFlags().StringVarP(&resolverMapUpdateOpts.FromFile, "from-file", "f", "", "path to a "+
		"graphql resolver map file from which to update the QLoo resolver map object")
	resolverMapCmd.AddCommand(resolverMapUpdateCmd)
}

func updateResolverMap(name, filename string) error {
	cli, err := qlooctl.MakeClient()
	if err != nil {
		return err
	}
	if name == "" {
		return errors.Errorf("schema name must be set")
	}
	if filename == "" {
		return errors.Errorf("filename must be set")
	}

	existing, err := cli.V1().ResolverMaps().Get(name)
	if err != nil {
		return err
	}

	if name == "" {
		return errors.Errorf("resolver map name must be set")
	}
	if filename == "" {
		return errors.Errorf("filename must be set")
	}
	var resolverMap v1.ResolverMap
	if err := file.ReadFileInto(filename, &resolverMap); err != nil {
		return err
	}
	resolverMap.Metadata = existing.Metadata
	_, err = cli.V1().ResolverMaps().Update(&resolverMap)
	return err
}
