package constants

import "github.com/spf13/cobra"

var (
	INSTALL_COMMAND = cobra.Command{
		Use:   "install",
		Short: "install sqoop on different platforms (into gloo-system namespace by default)",
		Long:  "choose which version of Sqoop to install.",
	}

	UNINSTALL_COMMAND = cobra.Command{
		Use:   "uninstall",
		Short: "uninstall sqoop and remove namespace (gloo-system by default)",
	}

	SCHEMA_COMMAND = cobra.Command{
		Use:     "schema",
		Short:   "interacting with sqoop schema resources",
		Aliases: []string{"schemas"},
	}

	RESOLVER_MAP_COMMAND = cobra.Command{
		Use:     "resolvermap",
		Short:   "interacting with sqoop resolver maps",
		Aliases: []string{"r", "rm", "resolvermaps"},
	}
)
