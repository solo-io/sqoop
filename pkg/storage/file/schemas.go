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

	gloov1 "github.com/solo-io/gloo/pkg/api/types/v1"
	"github.com/solo-io/gloo/pkg/log"
	"github.com/solo-io/qloo/pkg/api/types/v1"
	"github.com/solo-io/qloo/pkg/storage"
)

// TODO: evaluate efficiency of LSing a whole dir on every op
// so far this is preferable to caring what files are named
type schemasClient struct {
	dir           string
	syncFrequency time.Duration
}

func (c *schemasClient) Create(item *v1.Schema) (*v1.Schema, error) {
	if item.Name == "" {
		return nil, errors.Errorf("name required")
	}
	// set resourceversion on clone
	schemaClone, ok := proto.Clone(item).(*v1.Schema)
	if !ok {
		return nil, errors.New("internal error: output of proto.Clone was not expected type")
	}
	if schemaClone.Metadata == nil {
		schemaClone.Metadata = &gloov1.Metadata{}
	}
	schemaClone.Metadata.ResourceVersion = newOrIncrementResourceVer(schemaClone.Metadata.ResourceVersion)
	schemaFiles, err := c.pathsToSchemas()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read schema dir")
	}
	// error if exists already
	for file, existingUps := range schemaFiles {
		if existingUps.Name == item.Name {
			return nil, storage.NewAlreadyExistsErr(errors.Errorf("schema %v already defined in %s", item.Name, file))
		}
	}
	filename := filepath.Join(c.dir, item.Name+".yml")
	err = WriteToFile(filename, schemaClone)
	if err != nil {
		return nil, errors.Wrap(err, "failed creating file")
	}
	return schemaClone, nil
}

func (c *schemasClient) Update(item *v1.Schema) (*v1.Schema, error) {
	if item.Name == "" {
		return nil, errors.Errorf("name required")
	}
	if item.Metadata == nil || item.Metadata.ResourceVersion == "" {
		return nil, errors.New("resource version must be set for update operations")
	}
	schemaFiles, err := c.pathsToSchemas()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read schema dir")
	}
	// error if exists already
	for file, existingUps := range schemaFiles {
		if existingUps.Name != item.Name {
			continue
		}
		if existingUps.Metadata != nil && lessThan(item.Metadata.ResourceVersion, existingUps.Metadata.ResourceVersion) {
			return nil, errors.Errorf("resource version outdated for %v", item.Name)
		}
		schemaClone, ok := proto.Clone(item).(*v1.Schema)
		if !ok {
			return nil, errors.New("internal error: output of proto.Clone was not expected type")
		}
		schemaClone.Metadata.ResourceVersion = newOrIncrementResourceVer(schemaClone.Metadata.ResourceVersion)

		err = WriteToFile(file, schemaClone)
		if err != nil {
			return nil, errors.Wrap(err, "failed creating file")
		}

		return schemaClone, nil
	}
	return nil, errors.Errorf("schema %v not found", item.Name)
}

func (c *schemasClient) Delete(name string) error {
	schemaFiles, err := c.pathsToSchemas()
	if err != nil {
		return errors.Wrap(err, "failed to read schema dir")
	}
	// error if exists already
	for file, existingUps := range schemaFiles {
		if existingUps.Name == name {
			return os.Remove(file)
		}
	}
	return errors.Errorf("file not found for schema %v", name)
}

func (c *schemasClient) Get(name string) (*v1.Schema, error) {
	schemaFiles, err := c.pathsToSchemas()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read schema dir")
	}
	// error if exists already
	for _, existingUps := range schemaFiles {
		if existingUps.Name == name {
			return existingUps, nil
		}
	}
	return nil, errors.Errorf("file not found for schema %v", name)
}

func (c *schemasClient) List() ([]*v1.Schema, error) {
	schemaPaths, err := c.pathsToSchemas()
	if err != nil {
		return nil, err
	}
	var schemas []*v1.Schema
	for _, up := range schemaPaths {
		schemas = append(schemas, up)
	}
	return schemas, nil
}

func (c *schemasClient) pathsToSchemas() (map[string]*v1.Schema, error) {
	files, err := ioutil.ReadDir(c.dir)
	if err != nil {
		return nil, errors.Wrap(err, "could not read dir")
	}
	schemas := make(map[string]*v1.Schema)
	for _, f := range files {
		path := filepath.Join(c.dir, f.Name())
		if !strings.HasSuffix(path, ".yml") && !strings.HasSuffix(path, ".yaml") {
			continue
		}

		schema, err := pathToSchema(path)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse .yml file as schema")
		}

		schemas[path] = schema
	}
	return schemas, nil
}

func pathToSchema(path string) (*v1.Schema, error) {
	var schema v1.Schema
	err := ReadFileInto(path, &schema)
	if err != nil {
		return nil, err
	}
	if schema.Metadata == nil {
		schema.Metadata = &gloov1.Metadata{}
	}
	if schema.Metadata.ResourceVersion == "" {
		schema.Metadata.ResourceVersion = "1"
	}
	return &schema, nil
}

func (u *schemasClient) Watch(handlers ...storage.SchemaEventHandler) (*storage.Watcher, error) {
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

func (u *schemasClient) onEvent(event watcher.Event, handlers ...storage.SchemaEventHandler) error {
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
			created, err := pathToSchema(event.Path)
			if err != nil {
				return err
			}
			h.OnAdd(current, created)
		}
	case watcher.Write:
		for _, h := range handlers {
			updated, err := pathToSchema(event.Path)
			if err != nil {
				return err
			}
			h.OnUpdate(current, updated)
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
