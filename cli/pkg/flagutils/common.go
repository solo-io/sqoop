package flagutils

import (
	"github.com/solo-io/sqoop/cli/pkg/options"
	"github.com/spf13/pflag"
)

func AddOutputFlag(set *pflag.FlagSet, strptr *string) {
	set.StringVarP(strptr, "output", "o", "", "output format: (yaml, json, table)")
}

func AddFileFlag(set *pflag.FlagSet, strptr *string) {
	set.StringVarP(strptr, "file", "f", "", "file to be read or written to")
}

func AddInteractiveFlag(set *pflag.FlagSet, boolptr *bool) {
	set.BoolVarP(boolptr, "interactive", "i", false, "interactive mode")
}

func AddCommonFlags(set *pflag.FlagSet, opts *options.Top) {
	AddOutputFlag(set, &opts.Output)
	AddFileFlag(set, &opts.File)
	AddInteractiveFlag(set, &opts.Interactive)
}
