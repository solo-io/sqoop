package flagutils

import (
	"github.com/solo-io/sqoop/cli/pkg/options"
	"github.com/spf13/pflag"
)

func AddResolverMapFlags(set *pflag.FlagSet, opts *options.ResolverMap) {
	set.StringVarP(&opts.Upstream, "upstream", "u", "", "upstream where the function lives")
	set.StringVarP(&opts.Function, "function", "g", "", "function to use as resolver")
	set.StringVarP(&opts.RequestTemplate, "request-template", "b", "", "template to use for the request body")
	set.StringVarP(&opts.ResponseTemplate, "response-template", "r", "", "template to use for the response body")
	set.StringVarP(&opts.SchemaName, "schema", "s", "", "name of the "+
		"schema to connect this resolver to. this is required if more than one schema contains a definition for the "+
		"type name.")
}
