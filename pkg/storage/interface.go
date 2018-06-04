package storage

import "github.com/solo-io/qloo/pkg/api/types/v1"

// Interface is interface to the storage backend
type Interface interface {
	V1() V1
}

type V1 interface {
	Register() error
	Schemas() Schemas
	ResolverMaps() ResolverMaps
}

type Schemas interface {
	Create(*v1.Schema) (*v1.Schema, error)
	Update(*v1.Schema) (*v1.Schema, error)
	Delete(name string) error
	Get(name string) (*v1.Schema, error)
	List() ([]*v1.Schema, error)
	Watch(handlers ...SchemaEventHandler) (*Watcher, error)
}

type ResolverMaps interface {
	Create(*v1.ResolverMap) (*v1.ResolverMap, error)
	Update(*v1.ResolverMap) (*v1.ResolverMap, error)
	Delete(name string) error
	Get(name string) (*v1.ResolverMap, error)
	List() ([]*v1.ResolverMap, error)
	Watch(...ResolverMapEventHandler) (*Watcher, error)
}
