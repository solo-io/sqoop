package consul

import (
	"github.com/solo-io/qloo/pkg/api/types/v1"
	"github.com/solo-io/qloo/pkg/storage"
	"github.com/solo-io/qloo/pkg/storage/base"
)

type resolverMapsClient struct {
	base *base.ConsulStorageClient
}

func (c *resolverMapsClient) Create(item *v1.ResolverMap) (*v1.ResolverMap, error) {
	out, err := c.base.Create(&base.StorableItem{ResolverMap: item})
	if err != nil {
		return nil, err
	}
	return out.ResolverMap, nil
}

func (c *resolverMapsClient) Update(item *v1.ResolverMap) (*v1.ResolverMap, error) {
	out, err := c.base.Update(&base.StorableItem{ResolverMap: item})
	if err != nil {
		return nil, err
	}
	return out.ResolverMap, nil
}

func (c *resolverMapsClient) Delete(name string) error {
	return c.base.Delete(name)
}

func (c *resolverMapsClient) Get(name string) (*v1.ResolverMap, error) {
	out, err := c.base.Get(name)
	if err != nil {
		return nil, err
	}
	return out.ResolverMap, nil
}

func (c *resolverMapsClient) List() ([]*v1.ResolverMap, error) {
	list, err := c.base.List()
	if err != nil {
		return nil, err
	}
	var resolverMaps []*v1.ResolverMap
	for _, obj := range list {
		resolverMaps = append(resolverMaps, obj.ResolverMap)
	}
	return resolverMaps, nil
}

func (c *resolverMapsClient) Watch(handlers ...storage.ResolverMapEventHandler) (*storage.Watcher, error) {
	var baseHandlers []base.StorableItemEventHandler
	for _, h := range handlers {
		baseHandlers = append(baseHandlers, base.StorableItemEventHandler{ResolverMapEventHandler: h})
	}
	return c.base.Watch(baseHandlers...)
}
