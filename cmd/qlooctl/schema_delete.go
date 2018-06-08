package main

import (
	"github.com/spf13/cobra"
	"fmt"
)

var schemaDeleteCmd = &cobra.Command{
	Use:   "delete [NAME]",
	Short: "return a schema by its name or list all schemas",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 || args[0] == "" {
			list, err := listSchemas()
			if err != nil {
				return err
			}
			for _, msg := range list {
				if err := printAsYaml(msg); err != nil {
					return err
				}
			}
			return nil
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
	cli, err := makeClient()
	if err != nil {
		return err
	}
	return cli.V1().Schemas().Delete(name)
}
