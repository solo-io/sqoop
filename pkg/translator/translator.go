package translator

import (
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	matchers "github.com/solo-io/gloo/projects/gloo/pkg/api/v1/core/matchers"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/sqoop/pkg/api/v1"
	TODO "github.com/solo-io/sqoop/pkg/todo"
)

// trnslate a snapshot of schemas and resolvermaps to:
// a 1 proxy for the snapshot, assigned to the sqoop sidecar
func Translate(writeNamespace string, snap *v1.ApiSnapshot, resourceErrs reporter.ResourceErrors) *gloov1.Proxy {
	ourRoutes := routesForResolverMaps(snap.ResolverMaps, resourceErrs)

	var routes []*gloov1.Route

	for _, r := range ourRoutes {
		routes = append(routes, &gloov1.Route{
			Matchers: []*matchers.Matcher{
				{
					PathSpecifier: &matchers.Matcher_Exact{
						Exact: r.path,
					},
					Methods: []string{"POST"},
				},
			},
			Action: &gloov1.Route_RouteAction{RouteAction: r.action},
		})
	}

	return &gloov1.Proxy{
		Metadata: core.Metadata{
			Name:      "sqoop-proxy",
			Namespace: writeNamespace,
			Labels: map[string]string{
				"created_by": "sqoop",
			},
		},
		Listeners: []*gloov1.Listener{
			{
				// TODO (ilackarms): make this section configurable
				Name:        "sqoop-listener",
				BindAddress: TODO.SqoopSidecarBindAddr,
				BindPort:    TODO.SqoopSidecarBindPort,
				ListenerType: &gloov1.Listener_HttpListener{
					HttpListener: &gloov1.HttpListener{
						VirtualHosts: []*gloov1.VirtualHost{
							{
								Name:    "sqoop-vhost",
								Domains: []string{"*"},
								Routes:  routes,
							},
						},
					},
				},
				// TODO(ilackarms / yuval-k): decide if we need ssl for connecting to sidecar
				SslConfigurations: nil,
			},
		},
	}
}
