package flags

import (
	"github.com/spf13/cobra"
	"github.com/solo-io/qloo/pkg/bootstrap"
)

func AddQLooFlags(cmd *cobra.Command, opts *bootstrap.Options) {
	// TODO ingress.bind-adress
	cmd.PersistentFlags().StringVar(&opts.VirtualServiceName, "qloo.virtualservice", "qloo-routes", "the " +
		"name of the virtual service QLoo will use to store its routes")
	cmd.PersistentFlags().StringVar(&opts.RoleName, "qloo.role", "qloo-role", "the " +
		"name of the mesh role to assign to QLoo when communicating with Gloo")
	cmd.PersistentFlags().StringVar(&opts.ProxyAddr, "qloo.proxy-addr", "localhost:8080", "the " +
		"address (hostname:port) of the QLoo proxy")
}
