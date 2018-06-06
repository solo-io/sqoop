package consul

import (
	"github.com/pkg/errors"

	"github.com/solo-io/qloo/pkg/api/types/v1"
	"github.com/solo-io/qloo/pkg/storage"
	"github.com/solo-io/qloo/pkg/storage/base"
)

type schemasClient struct {
	base *base.ConsulStorageClient
}

func (c *schemasClient) Create(item *v1.Schema) (*v1.Schema, error) {
	if item.Name == "" {
		return nil, errors.Errorf("name required")
	}
	out, err := c.base.Create(&base.StorableItem{Schema: item})
	if err != nil {
		return nil, err
	}
	return out.Schema, nil
}

func (c *schemasClient) Update(item *v1.Schema) (*v1.Schema, error) {
	if item.Name == "" {
		return nil, errors.Errorf("name required")
	}
	out, err := c.base.Update(&base.StorableItem{Schema: item})
	if err != nil {
		return nil, err
	}
	return out.Schema, nil
}

func (c *schemasClient) Delete(name string) error {
	return c.base.Delete(name)
}

func (c *schemasClient) Get(name string) (*v1.Schema, error) {
	out, err := c.base.Get(name)
	if err != nil {
		return nil, err
	}
	return out.Schema, nil
}

func (c *schemasClient) List() ([]*v1.Schema, error) {
	list, err := c.base.List()
	if err != nil {
		return nil, err
	}
	var schemas []*v1.Schema
	for _, obj := range list {
		schemas = append(schemas, obj.Schema)
	}
	return schemas, nil
}

func (c *schemasClient) Watch(handlers ...storage.SchemaEventHandler) (*storage.Watcher, error) {
	var baseHandlers []base.StorableItemEventHandler
	for _, h := range handlers {
		baseHandlers = append(baseHandlers, base.StorableItemEventHandler{SchemaEventHandler: h})
	}
	return c.base.Watch(baseHandlers...)
}
