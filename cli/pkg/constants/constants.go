package constants

import "github.com/spf13/cobra"

var (
	SCHEMA_COMMAND = cobra.Command{
		Use:     "schema",
		Aliases: []string{"schemas"},
	}

	RESOLVER_MAP_COMMAND = cobra.Command{
		Use:     "resolvermap",
		Aliases: []string{"r", "rm", "resolvermaps"},
	}


)
