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

	"github.com/solo-io/gloo/pkg/log"
	"github.com/solo-io/qloo/pkg/storage/crud"
	kuberrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"
)

type resolverMapsClient struct {
	crds    crdclientset.Interface
	apiexts apiexts.Interface
	// write and read objects to this namespace if not specified on the QlooObjects
	namespace     string
	syncFrequency time.Duration
}

func (c *resolverMapsClient) Create(item *v1.ResolverMap) (*v1.ResolverMap, error) {
	return c.createOrUpdateResolverMapCrd(item, crud.OperationCreate)
}

func (c *resolverMapsClient) Update(item *v1.ResolverMap) (*v1.ResolverMap, error) {
	return c.createOrUpdateResolverMapCrd(item, crud.OperationUpdate)
}

func (c *resolverMapsClient) Delete(name string) error {
	return c.crds.QlooV1().ResolverMaps(c.namespace).Delete(name, nil)
}

func (c *resolverMapsClient) Get(name string) (*v1.ResolverMap, error) {
	crdResolverMap, err := c.crds.QlooV1().ResolverMaps(c.namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed performing get api request")
	}
	var returnedResolverMap v1.ResolverMap
	if err := ConfigObjectFromCrd(
		crdResolverMap.ObjectMeta,
		crdResolverMap.Spec,
		crdResolverMap.Status,
		&returnedResolverMap); err != nil {
		return nil, errors.Wrap(err, "converting returned crd to resolverMap")
	}
	return &returnedResolverMap, nil
}

func (c *resolverMapsClient) List() ([]*v1.ResolverMap, error) {
	crdList, err := c.crds.QlooV1().ResolverMaps(c.namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed performing list api request")
	}
	var returnedResolverMaps []*v1.ResolverMap
	for _, crdResolverMap := range crdList.Items {
		var returnedResolverMap v1.ResolverMap
		if err := ConfigObjectFromCrd(
			crdResolverMap.ObjectMeta,
			crdResolverMap.Spec,
			crdResolverMap.Status,
			&returnedResolverMap); err != nil {
			return nil, errors.Wrap(err, "converting returned crd to resolverMap")
		}
		returnedResolverMaps = append(returnedResolverMaps, &returnedResolverMap)
	}
	return returnedResolverMaps, nil
}

func (u *resolverMapsClient) Watch(handlers ...storage.ResolverMapEventHandler) (*storage.Watcher, error) {
	lw := cache.NewListWatchFromClient(u.crds.QlooV1().RESTClient(), crdv1.ResolverMapCRD.Plural, u.namespace, fields.Everything())
	sw := cache.NewSharedInformer(lw, new(crdv1.ResolverMap), u.syncFrequency)
	for _, h := range handlers {
		sw.AddEventHandler(&resolverMapEventHandler{handler: h, store: sw.GetStore()})
	}
	return storage.NewWatcher(func(stop <-chan struct{}, _ chan error) {
		sw.Run(stop)
	}), nil
}

func (c *resolverMapsClient) createOrUpdateResolverMapCrd(resolverMap *v1.ResolverMap, op crud.Operation) (*v1.ResolverMap, error) {
	resolverMapCrd, err := ConfigObjectToCrd(c.namespace, resolverMap)
	if err != nil {
		return nil, errors.Wrap(err, "converting qloo object to crd")
	}
	resolverMaps := c.crds.QlooV1().ResolverMaps(resolverMapCrd.GetNamespace())
	var returnedCrd *crdv1.ResolverMap
	switch op {
	case crud.OperationCreate:
		returnedCrd, err = resolverMaps.Create(resolverMapCrd.(*crdv1.ResolverMap))
		if err != nil {
			if kuberrs.IsAlreadyExists(err) {
				return nil, storage.NewAlreadyExistsErr(err)
			}
			return nil, errors.Wrap(err, "kubernetes create api request")
		}
	case crud.OperationUpdate:
		// need to make sure we preserve labels
		currentCrd, err := resolverMaps.Get(resolverMapCrd.GetName(), metav1.GetOptions{ResourceVersion: resolverMapCrd.GetResourceVersion()})
		if err != nil {
			return nil, errors.Wrap(err, "kubernetes get api request")
		}
		// copy labels
		resolverMapCrd.SetLabels(currentCrd.Labels)
		returnedCrd, err = resolverMaps.Update(resolverMapCrd.(*crdv1.ResolverMap))
		if err != nil {
			return nil, errors.Wrap(err, "kubernetes update api request")
		}
	}
	var returnedResolverMap v1.ResolverMap
	if err := ConfigObjectFromCrd(
		returnedCrd.ObjectMeta,
		returnedCrd.Spec,
		returnedCrd.Status,
		&returnedResolverMap); err != nil {
		return nil, errors.Wrap(err, "converting returned crd to resolverMap")
	}
	return &returnedResolverMap, nil
}

// implements the kubernetes ResourceEventHandler interface
type resolverMapEventHandler struct {
	handler storage.ResolverMapEventHandler
	store   cache.Store
}

func (eh *resolverMapEventHandler) getUpdatedList() []*v1.ResolverMap {
	updatedList := eh.store.List()
	var updatedResolverMapList []*v1.ResolverMap
	for _, updated := range updatedList {
		resolverMapCrd, ok := updated.(*crdv1.ResolverMap)
		if !ok {
			continue
		}
		var returnedResolverMap v1.ResolverMap
		if err := ConfigObjectFromCrd(
			resolverMapCrd.ObjectMeta,
			resolverMapCrd.Spec,
			resolverMapCrd.Status,
			&returnedResolverMap); err != nil {
			log.Warnf("watch event: %v", errors.Wrap(err, "converting returned crd to resolverMap"))
		}
		updatedResolverMapList = append(updatedResolverMapList, &returnedResolverMap)
	}
	return updatedResolverMapList
}

func convertResolverMap(obj interface{}) (*v1.ResolverMap, bool) {
	resolverMapCrd, ok := obj.(*crdv1.ResolverMap)
	if !ok {
		return nil, ok
	}
	var returnedResolverMap v1.ResolverMap
	if err := ConfigObjectFromCrd(
		resolverMapCrd.ObjectMeta,
		resolverMapCrd.Spec,
		resolverMapCrd.Status,
		&returnedResolverMap); err != nil {
		log.Warnf("watch event: %v", errors.Wrap(err, "converting returned crd to resolverMap"))
		return nil, false
	}
	return &returnedResolverMap, true
}

func (eh *resolverMapEventHandler) OnAdd(obj interface{}) {
	resolverMap, ok := convertResolverMap(obj)
	if !ok {
		return
	}
	eh.handler.OnAdd(eh.getUpdatedList(), resolverMap)
}
func (eh *resolverMapEventHandler) OnUpdate(_, newObj interface{}) {
	newResolverMap, ok := convertResolverMap(newObj)
	if !ok {
		return
	}
	eh.handler.OnUpdate(eh.getUpdatedList(), newResolverMap)
}

func (eh *resolverMapEventHandler) OnDelete(obj interface{}) {
	resolverMap, ok := convertResolverMap(obj)
	if !ok {
		return
	}
	eh.handler.OnDelete(eh.getUpdatedList(), resolverMap)
}
