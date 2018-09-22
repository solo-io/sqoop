package resolvermap

import (
	"github.com/solo-io/sqoop/pkg/sqoopctl"
	"github.com/spf13/cobra"
)

var resolverMapCmd = &cobra.Command{
	Use: "resolvermap",
	Aliases: []string{
		"resolvermaps", "rms",
	},
	Short: "Create, read, update, and delete Sqoop Resolver Maps",
	Long:  `Use these commands to define resolvers for your GraphQL Schemas`,
}

func init() {
	sqoopctl.RootCmd.AddCommand(resolverMapCmd)
}
