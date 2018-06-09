package resolvermap

import (
	"github.com/spf13/cobra"
	"github.com/solo-io/qloo/pkg/qlooctl"
)


var resolverMapCmd = &cobra.Command{
	Use: "resolvermap",
	Aliases: []string{
		"resolvermaps", "rms",
	},
	Short: "Create, read, update, and delete QLoo Resolver Maps",
	Long:  `Use these commands to define resolvers for your GraphQL Schemas`,
}

func init() {
	qlooctl.RootCmd.AddCommand(resolverMapCmd)
}
