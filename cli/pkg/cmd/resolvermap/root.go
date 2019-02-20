package resolvermap

import (
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/sqoop/cli/pkg/constants"
	"github.com/solo-io/sqoop/cli/pkg/options"
	"github.com/spf13/cobra"
)

func ResolverMapCmd(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   constants.RESOLVER_MAP_COMMAND.Use,
		Short: constants.RESOLVER_MAP_COMMAND.Short,
	}

	cmd.AddCommand(
		Reset(opts),
		Register(opts),
	)

	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}
