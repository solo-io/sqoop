package main

import (
	"github.com/spf13/cobra"
	"fmt"
	"github.com/pkg/errors"
)

var resolverMapDeleteCmd = &cobra.Command{
	Use:   "delete [NAME]",
	Short: "delete a resolver map by its name",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 || args[0] == "" {
			return errors.Errorf("must provide name")
		}
		err := deleteResolverMap(args[0])
		if err != nil {
			return err
		}
		fmt.Println("delete succesful")
	},
}

func init() {
	resolverMapCmd.AddCommand(resolverMapDeleteCmd)
}

func deleteResolverMap(name string) error {
	cli, err := makeClient()
	if err != nil {
		return err
	}
	return cli.V1().ResolverMaps().Delete(name)
}
