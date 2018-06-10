package resolvermap

import (
	"github.com/solo-io/qloo/pkg/qlooctl"
	"github.com/spf13/cobra"
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
