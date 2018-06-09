package schema

import (
	"github.com/spf13/cobra"
	"fmt"
	"github.com/vektah/gqlgen/neelance/errors"
)

var schemaDeleteCmd = &cobra.Command{
	Use:   "delete [NAME]",
	Short: "delete a schema by its name",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 || args[0] == "" {
			return errors.Errorf("must provide name")
		}
		err := deleteSchema(args[0])
		if err != nil {
			return err
		}
		fmt.Println("delete succesful")
	},
}

func init() {
	schemaCmd.AddCommand(schemaDeleteCmd)
}

func deleteSchema(name string) error {
	cli, err := qlooctl.MakeClient()
	if err != nil {
		return err
	}
	return cli.V1().Schemas().Delete(name)
}
