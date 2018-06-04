package file

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/radovskyb/watcher"

	"time"

	"github.com/solo-io/qloo/pkg/api/types/v1"
	"github.com/solo-io/gloo/pkg/log"
	"github.com/solo-io/qloo/pkg/storage"
)

// TODO: evaluate efficiency of LSing a whole dir on every op
// so far this is preferable to caring what files are named
type resolverMapsClient struct {
	dir           string
	syncFrequency time.Duration
}

func (c *resolverMapsClient) Create(item *v1.ResolverMap) (*v1.ResolverMap, error) {
	// set resourceversion on clone
	resolverMapClone, ok := proto.Clone(item).(*v1.ResolverMap)
	if !ok {
		return nil, errors.New("internal error: output of proto.Clone was not expected type")
	}
	if resolverMapClone.Metadata == nil {
		resolverMapClone.Metadata = &v1.Metadata{}
	}
	resolverMapClone.Metadata.ResourceVersion = newOrIncrementResourceVer(resolverMapClone.Metadata.ResourceVersion)
	resolverMapFiles, err := c.pathsToResolverMaps()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read resolverMap dir")
	}
	// error if exists already
	for file, existingUps := range resolverMapFiles {
		if existingUps.Name == item.Name {
			return nil, storage.NewAlreadyExistsErr(errors.Errorf("resolverMap %v already defined in %s", item.Name, file))
		}
	}
	filename := filepath.Join(c.dir, item.Name+".yml")
	err = WriteToFile(filename, resolverMapClone)
	if err != nil {
		return nil, errors.Wrap(err, "failed creating file")
	}
	return resolverMapClone, nil
}

func (c *resolverMapsClient) Update(item *v1.ResolverMap) (*v1.ResolverMap, error) {
	if item.Metadata == nil || item.Metadata.ResourceVersion == "" {
		return nil, errors.New("resource version must be set for update operations")
	}
	resolverMapFiles, err := c.pathsToResolverMaps()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read resolverMap dir")
	}
	// error if exists already
	for file, existingUps := range resolverMapFiles {
		if existingUps.Name != item.Name {
			continue
		}
		if existingUps.Metadata != nil && lessThan(item.Metadata.ResourceVersion, existingUps.Metadata.ResourceVersion) {
			return nil, errors.Errorf("resource version outdated for %v", item.Name)
		}
		resolverMapClone, ok := proto.Clone(item).(*v1.ResolverMap)
		if !ok {
			return nil, errors.New("internal error: output of proto.Clone was not expected type")
		}
		resolverMapClone.Metadata.ResourceVersion = newOrIncrementResourceVer(resolverMapClone.Metadata.ResourceVersion)

		err = WriteToFile(file, resolverMapClone)
		if err != nil {
			return nil, errors.Wrap(err, "failed creating file")
		}

		return resolverMapClone, nil
	}
	return nil, errors.Errorf("resolverMap %v not found", item.Name)
}

func (c *resolverMapsClient) Delete(name string) error {
	resolverMapFiles, err := c.pathsToResolverMaps()
	if err != nil {
		return errors.Wrap(err, "failed to read resolverMap dir")
	}
	// error if exists already
	for file, existingUps := range resolverMapFiles {
		if existingUps.Name == name {
			return os.Remove(file)
		}
	}
	return errors.Errorf("file not found for resolverMap %v", name)
}

func (c *resolverMapsClient) Get(name string) (*v1.ResolverMap, error) {
	resolverMapFiles, err := c.pathsToResolverMaps()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read resolverMap dir")
	}
	// error if exists already
	for _, existingUps := range resolverMapFiles {
		if existingUps.Name == name {
			return existingUps, nil
		}
	}
	return nil, errors.Errorf("file not found for resolverMap %v", name)
}

func (c *resolverMapsClient) List() ([]*v1.ResolverMap, error) {
	resolverMapPaths, err := c.pathsToResolverMaps()
	if err != nil {
		return nil, err
	}
	var resolverMaps []*v1.ResolverMap
	for _, up := range resolverMapPaths {
		resolverMaps = append(resolverMaps, up)
	}
	return resolverMaps, nil
}

func (c *resolverMapsClient) pathsToResolverMaps() (map[string]*v1.ResolverMap, error) {
	files, err := ioutil.ReadDir(c.dir)
	if err != nil {
		return nil, errors.Wrap(err, "could not read dir")
	}
	resolverMaps := make(map[string]*v1.ResolverMap)
	for _, f := range files {
		path := filepath.Join(c.dir, f.Name())
		if !strings.HasSuffix(path, ".yml") && !strings.HasSuffix(path, ".yaml") {
			continue
		}
		var resolverMap v1.ResolverMap
		err := ReadFileInto(path, &resolverMap)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse .yml file as resolverMap")
		}
		resolverMaps[path] = &resolverMap
	}
	return resolverMaps, nil
}

func (u *resolverMapsClient) Watch(handlers ...storage.ResolverMapEventHandler) (*storage.Watcher, error) {
	w := watcher.New()
	w.SetMaxEvents(0)
	w.FilterOps(watcher.Create, watcher.Write, watcher.Remove)
	if err := w.AddRecursive(u.dir); err != nil {
		return nil, errors.Wrapf(err, "failed to add directory %v", u.dir)
	}

	return storage.NewWatcher(func(stop <-chan struct{}, errs chan error) {
		go func() {
			if err := w.Start(u.syncFrequency); err != nil {
				errs <- err
			}
		}()
		// start the watch with an "initial read" event
		current, err := u.List()
		if err != nil {
			errs <- err
			return
		}
		for _, h := range handlers {
			h.OnAdd(current, nil)
		}
		for {
			select {
			case event := <-w.Event:
				if err := u.onEvent(event, handlers...); err != nil {
					log.Warnf("event handle error in file-based config storage client: %v", err)
				}
			case err := <-w.Error:
				log.Warnf("watcher error in file-based config storage client: %v", err)
				return
			case err := <-errs:
				log.Warnf("failed to start file watcher: %v", err)
				return
			case <-stop:
				w.Close()
				return
			}
		}
	}), nil
}

func (u *resolverMapsClient) onEvent(event watcher.Event, handlers ...storage.ResolverMapEventHandler) error {
	log.Debugf("file event: %v [%v]", event.Path, event.Op)
	current, err := u.List()
	if err != nil {
		return err
	}
	if event.IsDir() {
		return nil
	}
	switch event.Op {
	case watcher.Create:
		for _, h := range handlers {
			var created v1.ResolverMap
			err := ReadFileInto(event.Path, &created)
			if err != nil {
				return err
			}
			h.OnAdd(current, &created)
		}
	case watcher.Write:
		for _, h := range handlers {
			var updated v1.ResolverMap
			err := ReadFileInto(event.Path, &updated)
			if err != nil {
				return err
			}
			h.OnUpdate(current, &updated)
		}
	case watcher.Remove:
		for _, h := range handlers {
			// can't read the deleted object
			// callers beware
			h.OnDelete(current, nil)
		}
	}
	return nil
}
