package consul

import (
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"github.com/solo-io/qloo/pkg/storage"
	"github.com/solo-io/qloo/pkg/storage/base"
)

//go:generate go run ${GOPATH}/src/github.com/solo-io/qloo/pkg/storage/generate/generate_clients.go -f ${GOPATH}/src/github.com/solo-io/qloo/pkg/storage/consul/client_template.go.tmpl -o ${GOPATH}/src/github.com/solo-io/qloo/pkg/storage/consul/
type Client struct {
	v1 *v1client
}

// TODO: support basic auth and tls
func NewStorage(cfg *api.Config, rootPath string, syncFrequency time.Duration) (storage.Interface, error) {
	cfg.WaitTime = syncFrequency

	// Get a new client
	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "creating consul client")
	}

	return &Client{
		v1: &v1client{
			schemas: &schemasClient{
				base: base.NewConsulStorageClient(rootPath+"/schemas", client),
			},
			resolverMaps: &resolverMapsClient{
				base: base.NewConsulStorageClient(rootPath+"/resolverMaps", client),
			},
		},
	}, nil
}

func (c *Client) V1() storage.V1 {
	return c.v1
}

type v1client struct {
	schemas       *schemasClient
	resolverMaps *resolverMapsClient
}

func (c *v1client) Register() error {
	return nil
}

func (c *v1client) Schemas() storage.Schemas {
	return c.schemas
}

func (c *v1client) ResolverMaps() storage.ResolverMaps {
	return c.resolverMaps
}
