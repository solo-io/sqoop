package resolvermap

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/solo-io/sqoop/pkg/api/types/v1"
	"github.com/solo-io/sqoop/pkg/sqoopctl"
	"github.com/solo-io/sqoop/pkg/storage/file"
	"github.com/spf13/cobra"
)

var resolverMapCreateOpts struct {
	FromFile string
}

var resolverMapCreateCmd = &cobra.Command{
	Use:   "create NAME --from-file <path/to/your/sqoop/resolver map>",
	Short: "upload a resolver map to Sqoop from a local Sqoop ResolverMap yaml file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.Errorf("requires exactly 1 argument")
		}
		if err := createResolverMap(args[0], resolverMapCreateOpts.FromFile); err != nil {
			return err
		}
		fmt.Println("resolver map created successfully")
		return nil
	},
}

func init() {
	resolverMapCreateCmd.PersistentFlags().StringVarP(&resolverMapCreateOpts.FromFile, "from-file", "f", "", "path to a "+
		"graphql resolver map file from which to create the Sqoop resolver map object")
	resolverMapCmd.AddCommand(resolverMapCreateCmd)
}

func createResolverMap(name, filename string) error {
	if name == "" {
		return errors.Errorf("resolver map name must be set")
	}
	if filename == "" {
		return errors.Errorf("filename must be set")
	}
	cli, err := sqoopctl.MakeClient()
	if err != nil {
		return err
	}
	var resolverMap v1.ResolverMap
	if err := file.ReadFileInto(filename, &resolverMap); err != nil {
		return err
	}
	_, err = cli.V1().ResolverMaps().Create(&resolverMap)
	return err
}
