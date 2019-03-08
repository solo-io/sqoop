package resolvermap

import (
	"fmt"

	"github.com/solo-io/sqoop/cli/pkg/options"
	"github.com/spf13/cobra"
)

func ensureResolverMapParams(opts *options.ResolverMap, args []string) error {
	if opts.SchemaName == "" {
		return fmt.Errorf("schema name cannot be empty")
	}
	if opts.Function == "" {
		return fmt.Errorf("function name cannot be empty")
	}
	if opts.Upstream == "" {
		return fmt.Errorf("upstream cannot be empty")
	}
	opts.TypeName = args[0]
	opts.FieldName = args[1]
	return nil
}

func ensureResolverMapArgs(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("2 args are required (typeName, fieldName) %d were found", len(args))
	}
	return nil
}
