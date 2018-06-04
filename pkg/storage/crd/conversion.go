package crd

import (
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	gloov1 "github.com/solo-io/gloo/pkg/api/types/v1"
	"github.com/solo-io/gloo/pkg/protoutil"
	crdv1 "github.com/solo-io/qloo/pkg/storage/crd/solo.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/solo-io/qloo/pkg/api/types/v1"
)

func ConfigObjectToCrd(namespace string, item gloov1.ConfigObject) (metav1.Object, error) {
	name := item.GetName()
	var (
		status *gloov1.Status
		ok     bool
	)
	if item.GetStatus() != nil {
		status, ok = proto.Clone(item.GetStatus()).(*gloov1.Status)
		if !ok {
			return nil, errors.New("internal error: output of proto.Clone was not expected type")
		}
	}
	var (
		resourceVersion string
		annotations     map[string]string
	)
	if item.GetMetadata() != nil {
		resourceVersion = item.GetMetadata().ResourceVersion
		if item.GetMetadata().Namespace != "" {
			namespace = item.GetMetadata().Namespace
		}
		annotations = item.GetMetadata().Annotations
	}

	// TODO: merge this with ConfigObjectToCRD in Gloo
	// clone and remove fields
	var clone gloov1.ConfigObject
	switch item.(type) {
	case *v1.Schema:
		clone, ok = proto.Clone(item).(*v1.Schema)
		if !ok {
			return nil, errors.New("internal error: output of proto.Clone was not expected type")
		}
	case *v1.ResolverMap:
		// clone and remove fields
		clone, ok = proto.Clone(item).(*v1.ResolverMap)
		if !ok {
			return nil, errors.New("internal error: output of proto.Clone was not expected type")
		}
	default:
		panic(errors.Errorf("unknown type: %v", item))
	}
	clone.SetMetadata(nil)
	clone.SetName("")
	clone.SetStatus(nil)

	spec, err := protoutil.MarshalMap(clone)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert proto config object to map[string]interface{}")
	}
	copySpec := crdv1.Spec(spec)

	meta := metav1.ObjectMeta{
		Name:            name,
		Namespace:       namespace,
		ResourceVersion: resourceVersion,
		Annotations:     annotations,
	}

	var crdObject metav1.Object

	switch item.(type) {
	case *v1.Schema:
		crdObject = &crdv1.Schema{
			ObjectMeta: meta,
			Status:     status,
			Spec:       &copySpec,
		}
	case *v1.ResolverMap:
		crdObject = &crdv1.ResolverMap{
			ObjectMeta: meta,
			Status:     status,
			Spec:       &copySpec,
		}
	default:
		panic(errors.Errorf("unknown type: %v", item))
	}

	return crdObject, nil
}

func ConfigObjectFromCrd(objectMeta metav1.ObjectMeta,
	spec *crdv1.Spec,
	status *gloov1.Status,
	item gloov1.ConfigObject) error {
	if spec != nil {
		err := protoutil.UnmarshalMap(*spec, item)
		if err != nil {
			return errors.Wrap(err, "failed to convert crd spec to config object")
		}
	}
	// add removed fields to the internal object
	item.SetName(objectMeta.Name)
	item.SetMetadata(&gloov1.Metadata{
		ResourceVersion: objectMeta.ResourceVersion,
		Namespace:       objectMeta.Namespace,
		Annotations:     objectMeta.Annotations,
	})
	item.SetStatus(status)
	return nil
}
