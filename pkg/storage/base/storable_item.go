package base

import (
	"github.com/gogo/protobuf/proto"
	gloov1 "github.com/solo-io/gloo/pkg/api/types/v1"
	"github.com/solo-io/qloo/pkg/api/types/v1"
	"github.com/solo-io/qloo/pkg/storage"
)

type StorableItem struct {
	Schema      *v1.Schema
	ResolverMap *v1.ResolverMap
}

func (item *StorableItem) GetName() string {
	switch {
	case item.Schema != nil:
		return item.Schema.GetName()
	case item.ResolverMap != nil:
		return item.ResolverMap.GetName()
	default:
		panic("virtual service, role, fileor schema must be set")
	}
}

func (item *StorableItem) GetResourceVersion() string {
	switch {
	case item.Schema != nil:
		if item.Schema.GetMetadata() == nil {
			return ""
		}
		return item.Schema.GetMetadata().GetResourceVersion()
	case item.ResolverMap != nil:
		if item.ResolverMap.GetMetadata() == nil {
			return ""
		}
		return item.ResolverMap.GetMetadata().GetResourceVersion()
	default:
		panic("virtual service, role, fileor schema must be set")
	}
}

func (item *StorableItem) SetResourceVersion(rv string) {
	switch {
	case item.Schema != nil:
		if item.Schema.GetMetadata() == nil {
			item.Schema.Metadata = &gloov1.Metadata{}
		}
		item.Schema.Metadata.ResourceVersion = rv
	case item.ResolverMap != nil:
		if item.ResolverMap.GetMetadata() == nil {
			item.ResolverMap.Metadata = &gloov1.Metadata{}
		}
		item.ResolverMap.Metadata.ResourceVersion = rv
	default:
		panic("virtual service, role, fileor schema must be set")
	}
}

func (item *StorableItem) GetBytes() ([]byte, error) {
	switch {
	case item.Schema != nil:
		return proto.Marshal(item.Schema)
	case item.ResolverMap != nil:
		return proto.Marshal(item.ResolverMap)
	default:
		panic("virtual service, role, fileor schema must be set")
	}
}

func (item *StorableItem) GetTypeFlag() StorableItemType {
	switch {
	case item.Schema != nil:
		return StorableItemTypeSchema
	case item.ResolverMap != nil:
		return StorableItemTypeResolverMap
	default:
		panic("virtual service, role, fileor schema must be set")
	}
}

type StorableItemType uint64

const (
	StorableItemTypeSchema StorableItemType = iota
	StorableItemTypeResolverMap
)

type StorableItemEventHandler struct {
	SchemaEventHandler      storage.SchemaEventHandler
	ResolverMapEventHandler storage.ResolverMapEventHandler
}
