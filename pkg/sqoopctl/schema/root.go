package schema

import (
	"github.com/solo-io/sqoop/pkg/sqoopctl"
	"github.com/spf13/cobra"
)

var schemaCmd = &cobra.Command{
	Use: "schema",
	Aliases: []string{
		"schemas", "s",
	},
	Short: "Create, read, update, and delete GraphQL schemas for Sqoop",
	Long:  `Use these commands to register a GraphQL schema with Sqoop`,
}

func init() {
	sqoopctl.RootCmd.AddCommand(schemaCmd)
}
