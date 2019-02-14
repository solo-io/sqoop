package schema

import (
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/sqoop/cli/pkg/constants"
	"github.com/solo-io/sqoop/cli/pkg/flagutils"
	"github.com/solo-io/sqoop/cli/pkg/options"
	"github.com/spf13/cobra"
)

func SchemaCmd(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   constants.SCHEMA_COMMAND.Use,
		Short: constants.SCHEMA_COMMAND.Short,
	}

	cmd.AddCommand(
		Create(opts),
		Delete(opts),
		Update(opts),
	)

	flagutils.AddMetadataFlags(cmd.PersistentFlags(), &opts.Metadata)

	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}
