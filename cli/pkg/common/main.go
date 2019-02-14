package common

import (
	"fmt"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/spf13/cobra"
)

func RequiredNameArg(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("name are required, no arg(s) found")
	}
	return nil
}

func SetResourceName(opts *core.Metadata, args []string) error {
	if len(args) > 0 {
		opts.Name = args[0]
		return nil
	}
	if opts.Name == "" {
		return fmt.Errorf("no resource name could be found")
	}
	return nil
}