package schema

import (
	"github.com/solo-io/qloo/pkg/api/types/v1"
	"github.com/solo-io/qloo/pkg/qlooctl"
	"github.com/spf13/cobra"
)

var schemaGetCmd = &cobra.Command{
	Use:   "get [NAME]",
	Short: "return a schema by its name or list all schemas",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 || args[0] == "" {
			list, err := listSchemas()
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
		msg, err := getSchema(args[0])
		if err != nil {
			return err
		}
		return qlooctl.PrintAsYaml(msg)
	},
}

func init() {
	schemaCmd.AddCommand(schemaGetCmd)
}

func getSchema(name string) (*v1.Schema, error) {
	cli, err := qlooctl.MakeClient()
	if err != nil {
		return nil, err
	}
	return cli.V1().Schemas().Get(name)
}

func listSchemas() ([]*v1.Schema, error) {
	cli, err := qlooctl.MakeClient()
	if err != nil {
		return nil, err
	}
	return cli.V1().Schemas().List()
}
