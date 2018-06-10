package flags

import (
	"github.com/solo-io/qloo/pkg/bootstrap"
	"github.com/spf13/cobra"
)

func AddQLooFlags(cmd *cobra.Command, opts *bootstrap.Options) {
	// TODO ingress.bind-adress
	cmd.PersistentFlags().StringVar(&opts.VirtualServiceName, "qloo.virtualservice", "qloo-routes", "the "+
		"name of the virtual service QLoo will use to store its routes")
	cmd.PersistentFlags().StringVar(&opts.RoleName, "qloo.role", "qloo", "the "+
		"name of the mesh role to assign to QLoo when communicating with Gloo")
	cmd.PersistentFlags().StringVar(&opts.ProxyAddr, "qloo.proxy-addr", "localhost:8080", "the "+
		"address (hostname:port) of the QLoo proxy")
	cmd.PersistentFlags().StringVar(&opts.BindAddr, "qloo.bind-addr", ":9090", "the "+
		"address for the QLoo server to listen on")
}
