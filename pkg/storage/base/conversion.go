package base

import (
	"fmt"
	"strconv"

	"github.com/gogo/protobuf/proto"
	"github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"github.com/solo-io/sqoop/pkg/api/types/v1"
)

func key(rootPath, itemName string) string {
	return rootPath + "/" + itemName
}

func toKVPair(rootPath string, item *StorableItem) (*api.KVPair, error) {
	data, err := item.GetBytes()
	if err != nil {
		return nil, errors.Wrap(err, "getting bytes to store")
	}
	var modifyIndex uint64
	if item.GetResourceVersion() != "" {
		if i, err := strconv.Atoi(item.GetResourceVersion()); err == nil {
			modifyIndex = uint64(i)
		}
	}
	return &api.KVPair{
		Key:         key(rootPath, item.GetName()),
		Value:       data,
		Flags:       uint64(item.GetTypeFlag()),
		ModifyIndex: modifyIndex,
	}, nil
}

func setResourceVersion(item *StorableItem, p *api.KVPair) {
	resourceVersion := fmt.Sprintf("%v", p.ModifyIndex)
	item.SetResourceVersion(resourceVersion)
}

func itemFromKVPair(p *api.KVPair) (*StorableItem, error) {
	item := &StorableItem{}
	switch StorableItemType(p.Flags) {
	case StorableItemTypeSchema:
		var schema v1.Schema
		err := proto.Unmarshal(p.Value, &schema)
		if err != nil {
			return nil, errors.Wrap(err, "unmarshalling value as schema")
		}
		item.Schema = &schema
	case StorableItemTypeResolverMap:
		var resolverMap v1.ResolverMap
		err := proto.Unmarshal(p.Value, &resolverMap)
		if err != nil {
			return nil, errors.Wrap(err, "unmarshalling value as virtual service")
		}
		item.ResolverMap = &resolverMap
	}
	setResourceVersion(item, p)
	return item, nil
}
