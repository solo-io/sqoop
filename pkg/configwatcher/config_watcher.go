package configwatcher

import (
	"fmt"
	"sort"
	"sync"

	"github.com/pkg/errors"

	"github.com/gogo/protobuf/proto"
	"github.com/mitchellh/hashstructure"
	"github.com/solo-io/gloo/pkg/log"
	"github.com/solo-io/qloo/pkg/api/types/v1"
	"github.com/solo-io/qloo/pkg/storage"
)

type configWatcher struct {
	watchers []*storage.Watcher
	configs  chan *v1.Config
	errs     chan error
}

func NewConfigWatcher(storageClient storage.Interface) (*configWatcher, error) {
	if err := storageClient.V1().Register(); err != nil && !storage.IsAlreadyExists(err) {
		return nil, fmt.Errorf("failed to register to storage backend: %v", err)
	}

	configs := make(chan *v1.Config, 1)
	// do a first time read
	cache := &v1.Config{
		Schemas:      nil,
		ResolverMaps: nil,
	}

	syncSchemas := func(updatedList []*v1.Schema, _ *v1.Schema) {
		sort.SliceStable(updatedList, func(i, j int) bool {
			return updatedList[i].GetName() < updatedList[j].GetName()
		})

		oldHash, newHash := hashSchemas(cache.Schemas), hashSchemas(updatedList)
		if oldHash == newHash {
			return
		}
		log.GreyPrintf("\nold hash: %v\nnew hash: %v", oldHash, newHash)

		cache.Schemas = updatedList
		configs <- proto.Clone(cache).(*v1.Config)
	}
	schemaWatcher, err := storageClient.V1().Schemas().Watch(&storage.SchemaEventHandlerFuncs{
		AddFunc:    syncSchemas,
		UpdateFunc: syncSchemas,
		DeleteFunc: syncSchemas,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create watcher for schemas")
	}

	syncResolverMaps := func(updatedList []*v1.ResolverMap, _ *v1.ResolverMap) {
		sort.SliceStable(updatedList, func(i, j int) bool {
			return updatedList[i].GetName() < updatedList[j].GetName()
		})

		oldHash, newHash := hashResolverMaps(cache.ResolverMaps), hashResolverMaps(updatedList)
		if oldHash == newHash {
			return
		}
		log.GreyPrintf("\nold hash: %v\nnew hash: %v", oldHash, newHash)

		cache.ResolverMaps = updatedList
		configs <- proto.Clone(cache).(*v1.Config)
	}
	resolverMapWatcher, err := storageClient.V1().ResolverMaps().Watch(&storage.ResolverMapEventHandlerFuncs{
		AddFunc:    syncResolverMaps,
		UpdateFunc: syncResolverMaps,
		DeleteFunc: syncResolverMaps,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create watcher for virtualservices")
	}

	return &configWatcher{
		watchers: []*storage.Watcher{resolverMapWatcher, schemaWatcher},
		configs:  configs,
		errs:     make(chan error),
	}, nil
}

func (w *configWatcher) Run(stop <-chan struct{}) {
	done := &sync.WaitGroup{}
	for _, watcher := range w.watchers {
		done.Add(1)
		go func(watcher *storage.Watcher, stop <-chan struct{}, errs chan error) {
			watcher.Run(stop, errs)
			done.Done()
		}(watcher, stop, w.errs)
	}
	done.Wait()
}

func (w *configWatcher) Config() <-chan *v1.Config {
	return w.configs
}

func (w *configWatcher) Error() <-chan error {
	return w.errs
}

func hashSchemas(schemas []*v1.Schema) uint64 {
	// shave off status and resource version
	for _, item := range schemas {
		item.Status = nil
		if item.Metadata != nil {
			item.Metadata.ResourceVersion = ""
		}
	}
	h, err := hashstructure.Hash(schemas, nil)
	if err != nil {
		panic(err)
	}
	return h
}

func hashResolverMaps(resolverMaps []*v1.ResolverMap) uint64 {
	// shave off status and resource version
	for _, item := range resolverMaps {
		item.Status = nil
		if item.Metadata != nil {
			item.Metadata.ResourceVersion = ""
		}
	}
	h, err := hashstructure.Hash(resolverMaps, nil)
	if err != nil {
		panic(err)
	}
	return h
}
