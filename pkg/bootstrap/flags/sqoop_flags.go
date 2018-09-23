package flags

import (
	"github.com/solo-io/sqoop/pkg/bootstrap"
	"github.com/spf13/cobra"
)

func AddSqoopFlags(cmd *cobra.Command, opts *bootstrap.Options) {
	// TODO ingress.bind-adress
	cmd.PersistentFlags().StringVar(&opts.VirtualServiceName, "sqoop.virtualservice", "sqoop-routes", "the "+
		"name of the virtual service Sqoop will use to store its routes")
	cmd.PersistentFlags().StringVar(&opts.RoleName, "sqoop.role", "sqoop", "the "+
		"name of the mesh role to assign to Sqoop when communicating with Gloo")
	cmd.PersistentFlags().StringVar(&opts.ProxyAddr, "sqoop.proxy-addr", "localhost:8080", "the "+
		"address (hostname:port) of the Sqoop proxy")
	cmd.PersistentFlags().StringVar(&opts.BindAddr, "sqoop.bind-addr", ":9090", "the "+
		"address for the Sqoop server to listen on")
}
