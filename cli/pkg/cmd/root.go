package cmd

import (
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/sqoop/cli/pkg/cmd/install"
	"github.com/solo-io/sqoop/cli/pkg/cmd/resolvermap"
	"github.com/solo-io/sqoop/cli/pkg/cmd/schema"
	"github.com/solo-io/sqoop/cli/pkg/flagutils"
	"github.com/solo-io/sqoop/cli/pkg/options"
	"github.com/spf13/cobra"
)

func App(version string, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {

	opts := &options.Options{}

	app := &cobra.Command{
		Use:   "sqoopctl",
		Short: "Interact with Sqoop's storage API from the command line. \nFor more information, visit https://sqoop.solo.io.",
		Version: version,
	}
	pflags := app.PersistentFlags()
	flagutils.AddCommonFlags(pflags, &opts.Top)

	app.AddCommand(
		install.InstallCmd(opts),
		install.UninstallCmd(opts),
		resolvermap.ResolverMapCmd(opts),
		schema.SchemaCmd(opts),
	)

	// Complete additional passed in setup
	cliutils.ApplyOptions(app, optionsFunc)

	return app
}
