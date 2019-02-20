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
		Short: "Interact with Sqoop's storage API from the command line",
		Long: "As Sqoop features a storage-based API, direct communication with " +
			"the Sqoop server is not necessary. sqoopctl simplifies the administration of " +
			"Sqoop by providing an easy way to create, read, update, and delete Sqoop storage objects.\n\n" +
			"" +
			"The primary concerns of sqoopctl are Schemas and ResolverMaps. Schemas contain your GraphQL schema;" +
			" ResolverMaps define how your schema fields are resolved.\n\n" +
			"" +
			"Start by creating a schema using sqoopctl schema create --from-file <path/to/your/graphql/schema>",
		Run: func(cmd *cobra.Command, args []string) {
			panic("not implemented")
		},
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
