package install

import (
"github.com/solo-io/gloo/projects/gloo/cli/pkg/constants"
"github.com/solo-io/go-utils/cliutils"
"github.com/solo-io/sqoop/cli/pkg/options"
"github.com/spf13/cobra"
)

func InstallCmd(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   constants.INSTALL_COMMAND.Use,
		Short: constants.INSTALL_COMMAND.Short,
		Long:  constants.INSTALL_COMMAND.Long,
	}
	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}
