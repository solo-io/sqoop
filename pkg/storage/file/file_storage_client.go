package file

import (
	"os"
	"path/filepath"
	"time"

	"github.com/solo-io/sqoop/pkg/storage"
)

//go:generate go run ${GOPATH}/src/github.com/solo-io/sqoop/pkg/storage/generate/generate_clients.go -f ${GOPATH}/src/github.com/solo-io/sqoop/pkg/storage/file/client_template.go.tmpl -o ${GOPATH}/src/github.com/solo-io/sqoop/pkg/storage/file/

type Client struct {
	v1 *v1client
}

const schemasDir = "schemas"
const resolverMapsDir = "resolver_maps"

func NewStorage(dir string, syncFrequency time.Duration) (storage.Interface, error) {
	if dir == "" {
		dir = GlooDefaultDirectory
	}
	return &Client{
		v1: &v1client{
			schemas: &schemasClient{
				dir:           filepath.Join(dir, schemasDir),
				syncFrequency: syncFrequency,
			},
			resolverMaps: &resolverMapsClient{
				dir:           filepath.Join(dir, resolverMapsDir),
				syncFrequency: syncFrequency,
			},
		},
	}, nil
}

func (c *Client) V1() storage.V1 {
	return c.v1
}

type v1client struct {
	schemas      *schemasClient
	resolverMaps *resolverMapsClient
}

func (c *v1client) Register() error {
	err := os.MkdirAll(c.schemas.dir, 0755)
	if err != nil && err != os.ErrExist {
		return err
	}
	err = os.MkdirAll(c.resolverMaps.dir, 0755)
	if err != nil && err != os.ErrExist {
		return err
	}
	return nil
}

func (c *v1client) Schemas() storage.Schemas {
	return c.schemas
}

func (c *v1client) ResolverMaps() storage.ResolverMaps {
	return c.resolverMaps
}
