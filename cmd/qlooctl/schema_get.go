package main

import (
	"github.com/spf13/cobra"
	"github.com/solo-io/qloo/pkg/api/types/v1"
	"github.com/pkg/errors"
	"io/ioutil"
)

var schemaGetCmd = &cobra.Command{
	Use:   "get NAME",
	Short: "return a schema by its name",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.Errorf("requires exactly 1 argument")
		}
		msg, err := getSchema(args[0])
		if err != nil {
			return err
		}
		return printAsYaml(msg)
	},
}

func init() {
	schemaCmd.AddCommand(schemaGetCmd)
}

func getSchema(name string) (*v1.Schema, error) {
	cli, err := makeClient()
	if err != nil {
		return nil, err
	}
	return cli.V1().Schemas().Get(name)
}
