package crd

import (
	"time"

	"github.com/pkg/errors"
	"github.com/solo-io/qloo/pkg/api/types/v1"
	"github.com/solo-io/qloo/pkg/storage"
	crdclientset "github.com/solo-io/qloo/pkg/storage/crd/client/clientset/versioned"
	crdv1 "github.com/solo-io/qloo/pkg/storage/crd/solo.io/v1"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"

	"github.com/solo-io/qloo/pkg/storage/crud"
	kuberrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"
	"github.com/solo-io/gloo/pkg/log"
)

type schemasClient struct {
	crds    crdclientset.Interface
	apiexts apiexts.Interface
	// write and read objects to this namespace if not specified on the QlooObjects
	namespace     string
	syncFrequency time.Duration
}

func (c *schemasClient) Create(item *v1.Schema) (*v1.Schema, error) {
	return c.createOrUpdateSchemaCrd(item, crud.OperationCreate)
}

func (c *schemasClient) Update(item *v1.Schema) (*v1.Schema, error) {
	return c.createOrUpdateSchemaCrd(item, crud.OperationUpdate)
}

func (c *schemasClient) Delete(name string) error {
	return c.crds.QlooV1().Schemas(c.namespace).Delete(name, nil)
}

func (c *schemasClient) Get(name string) (*v1.Schema, error) {
	crdSchema, err := c.crds.QlooV1().Schemas(c.namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed performing get api request")
	}
	var returnedSchema v1.Schema
	if err := ConfigObjectFromCrd(
		crdSchema.ObjectMeta,
		crdSchema.Spec,
		crdSchema.Status,
		&returnedSchema); err != nil {
		return nil, errors.Wrap(err, "converting returned crd to schema")
	}
	return &returnedSchema, nil
}

func (c *schemasClient) List() ([]*v1.Schema, error) {
	crdList, err := c.crds.QlooV1().Schemas(c.namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed performing list api request")
	}
	var returnedSchemas []*v1.Schema
	for _, crdSchema := range crdList.Items {
		var returnedSchema v1.Schema
		if err := ConfigObjectFromCrd(
			crdSchema.ObjectMeta,
			crdSchema.Spec,
			crdSchema.Status,
			&returnedSchema); err != nil {
			return nil, errors.Wrap(err, "converting returned crd to schema")
		}
		returnedSchemas = append(returnedSchemas, &returnedSchema)
	}
	return returnedSchemas, nil
}

func (u *schemasClient) Watch(handlers ...storage.SchemaEventHandler) (*storage.Watcher, error) {
	lw := cache.NewListWatchFromClient(u.crds.QlooV1().RESTClient(), crdv1.SchemaCRD.Plural, u.namespace, fields.Everything())
	sw := cache.NewSharedInformer(lw, new(crdv1.Schema), u.syncFrequency)
	for _, h := range handlers {
		sw.AddEventHandler(&schemaEventHandler{handler: h, store: sw.GetStore()})
	}
	return storage.NewWatcher(func(stop <-chan struct{}, _ chan error) {
		sw.Run(stop)
	}), nil
}

func (c *schemasClient) createOrUpdateSchemaCrd(schema *v1.Schema, op crud.Operation) (*v1.Schema, error) {
	schemaCrd, err := ConfigObjectToCrd(c.namespace, schema)
	if err != nil {
		return nil, errors.Wrap(err, "converting qloo object to crd")
	}
	schemas := c.crds.QlooV1().Schemas(schemaCrd.GetNamespace())
	var returnedCrd *crdv1.Schema
	switch op {
	case crud.OperationCreate:
		returnedCrd, err = schemas.Create(schemaCrd.(*crdv1.Schema))
		if err != nil {
			if kuberrs.IsAlreadyExists(err) {
				return nil, storage.NewAlreadyExistsErr(err)
			}
			return nil, errors.Wrap(err, "kubernetes create api request")
		}
	case crud.OperationUpdate:
		// need to make sure we preserve labels
		currentCrd, err := schemas.Get(schemaCrd.GetName(), metav1.GetOptions{ResourceVersion: schemaCrd.GetResourceVersion()})
		if err != nil {
			return nil, errors.Wrap(err, "kubernetes get api request")
		}
		// copy labels
		schemaCrd.SetLabels(currentCrd.Labels)
		returnedCrd, err = schemas.Update(schemaCrd.(*crdv1.Schema))
		if err != nil {
			return nil, errors.Wrap(err, "kubernetes update api request")
		}
	}
	var returnedSchema v1.Schema
	if err := ConfigObjectFromCrd(
		returnedCrd.ObjectMeta,
		returnedCrd.Spec,
		returnedCrd.Status,
		&returnedSchema); err != nil {
		return nil, errors.Wrap(err, "converting returned crd to schema")
	}
	return &returnedSchema, nil
}

// implements the kubernetes ResourceEventHandler interface
type schemaEventHandler struct {
	handler storage.SchemaEventHandler
	store   cache.Store
}

func (eh *schemaEventHandler) getUpdatedList() []*v1.Schema {
	updatedList := eh.store.List()
	var updatedSchemaList []*v1.Schema
	for _, updated := range updatedList {
		schemaCrd, ok := updated.(*crdv1.Schema)
		if !ok {
			continue
		}
		var returnedSchema v1.Schema
		if err := ConfigObjectFromCrd(
			schemaCrd.ObjectMeta,
			schemaCrd.Spec,
			schemaCrd.Status,
			&returnedSchema); err != nil {
			log.Warnf("watch event: %v", errors.Wrap(err, "converting returned crd to schema"))
		}
		updatedSchemaList = append(updatedSchemaList, &returnedSchema)
	}
	return updatedSchemaList
}

func convertSchema(obj interface{}) (*v1.Schema, bool) {
	schemaCrd, ok := obj.(*crdv1.Schema)
	if !ok {
		return nil, ok
	}
	var returnedSchema v1.Schema
	if err := ConfigObjectFromCrd(
		schemaCrd.ObjectMeta,
		schemaCrd.Spec,
		schemaCrd.Status,
		&returnedSchema); err != nil {
		log.Warnf("watch event: %v", errors.Wrap(err, "converting returned crd to schema"))
		return nil, false
	}
	return &returnedSchema, true
}

func (eh *schemaEventHandler) OnAdd(obj interface{}) {
	schema, ok := convertSchema(obj)
	if !ok {
		return
	}
	eh.handler.OnAdd(eh.getUpdatedList(), schema)
}
func (eh *schemaEventHandler) OnUpdate(_, newObj interface{}) {
	newSchema, ok := convertSchema(newObj)
	if !ok {
		return
	}
	eh.handler.OnUpdate(eh.getUpdatedList(), newSchema)
}

func (eh *schemaEventHandler) OnDelete(obj interface{}) {
	schema, ok := convertSchema(obj)
	if !ok {
		return
	}
	eh.handler.OnDelete(eh.getUpdatedList(), schema)
}
