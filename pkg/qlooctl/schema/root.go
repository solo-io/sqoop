package schema

import (
	"github.com/spf13/cobra"
	"github.com/solo-io/qloo/pkg/qlooctl"
)

var schemaCmd = &cobra.Command{
	Use: "schema",
	Aliases: []string{
		"schemas", "s",
	},
	Short: "Create, read, update, and delete GraphQL schemas for QLoo",
	Long:  `Use these commands to register a GraphQL schema with QLoo`,
}

func init() {
	qlooctl.RootCmd.AddCommand(schemaCmd)
}
