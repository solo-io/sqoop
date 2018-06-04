package storage

import "github.com/solo-io/qloo/pkg/api/types/v1"

type Watcher struct {
	runFunc func(stop <-chan struct{}, errs chan error)
}

func NewWatcher(runFunc func(stop <-chan struct{}, errs chan error)) *Watcher {
	return &Watcher{runFunc: runFunc}
}

func (w *Watcher) Run(stop <-chan struct{}, errs chan error) {
	w.runFunc(stop, errs)
}

type SchemaEventHandler interface {
	OnAdd(updatedList []*v1.Schema, obj *v1.Schema)
	OnUpdate(updatedList []*v1.Schema, newObj *v1.Schema)
	OnDelete(updatedList []*v1.Schema, obj *v1.Schema)
}

type ResolverMapEventHandler interface {
	OnAdd(updatedList []*v1.ResolverMap, obj *v1.ResolverMap)
	OnUpdate(updatedList []*v1.ResolverMap, newObj *v1.ResolverMap)
	OnDelete(updatedList []*v1.ResolverMap, obj *v1.ResolverMap)
}

// SchemaEventHandlerFuncs is an adaptor to let you easily specify as many or
// as few of the notification functions as you want while still implementing
// SchemaEventHandler.
type SchemaEventHandlerFuncs struct {
	AddFunc    func(updatedList []*v1.Schema, obj *v1.Schema)
	UpdateFunc func(updatedList []*v1.Schema, newObj *v1.Schema)
	DeleteFunc func(updatedList []*v1.Schema, obj *v1.Schema)
}

// OnAdd calls AddFunc if it's not nil.
func (r SchemaEventHandlerFuncs) OnAdd(updatedList []*v1.Schema, obj *v1.Schema) {
	if r.AddFunc != nil {
		r.AddFunc(updatedList, obj)
	}
}

// OnUpdate calls UpdateFunc if it's not nil.
func (r SchemaEventHandlerFuncs) OnUpdate(updatedList []*v1.Schema, newObj *v1.Schema) {
	if r.UpdateFunc != nil {
		r.UpdateFunc(updatedList, newObj)
	}
}

// OnDelete calls DeleteFunc if it's not nil.
func (r SchemaEventHandlerFuncs) OnDelete(updatedList []*v1.Schema, obj *v1.Schema) {
	if r.DeleteFunc != nil {
		r.DeleteFunc(updatedList, obj)
	}
}

// ResolverMapEventHandlerFuncs is an adaptor to let you easily specify as many or
// as few of the notification functions as you want while still implementing
// ResolverMapEventHandler.
type ResolverMapEventHandlerFuncs struct {
	AddFunc    func(updatedList []*v1.ResolverMap, obj *v1.ResolverMap)
	UpdateFunc func(updatedList []*v1.ResolverMap, newObj *v1.ResolverMap)
	DeleteFunc func(updatedList []*v1.ResolverMap, obj *v1.ResolverMap)
}

// OnAdd calls AddFunc if it's not nil.
func (r ResolverMapEventHandlerFuncs) OnAdd(updatedList []*v1.ResolverMap, obj *v1.ResolverMap) {
	if r.AddFunc != nil {
		r.AddFunc(updatedList, obj)
	}
}

// OnUpdate calls UpdateFunc if it's not nil.
func (r ResolverMapEventHandlerFuncs) OnUpdate(updatedList []*v1.ResolverMap, newObj *v1.ResolverMap) {
	if r.UpdateFunc != nil {
		r.UpdateFunc(updatedList, newObj)
	}
}

// OnDelete calls DeleteFunc if it's not nil.
func (r ResolverMapEventHandlerFuncs) OnDelete(updatedList []*v1.ResolverMap, obj *v1.ResolverMap) {
	if r.DeleteFunc != nil {
		r.DeleteFunc(updatedList, obj)
	}
}
