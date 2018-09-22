package crd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	gloov1 "github.com/solo-io/gloo/pkg/api/types/v1"
	. "github.com/solo-io/sqoop/pkg/storage/crd"
	crdv1 "github.com/solo-io/sqoop/pkg/storage/crd/solo.io/v1"
)

var _ = Describe("Conversion", func() {
	Describe("SchemaToCrd", func() {
		It("Converts a gloo schema to crd", func() {
			schema := NewTestSchema1()
			annotations := map[string]string{"foo": "bar"}
			schema.Metadata = &gloov1.Metadata{
				Annotations: annotations,
			}
			upCrd, err := ConfigObjectToCrd("foo", schema)
			Expect(err).NotTo(HaveOccurred())
			Expect(upCrd.GetName()).To(Equal(schema.Name))
			Expect(upCrd.GetNamespace()).To(Equal("foo"))
			Expect(upCrd.GetAnnotations()).To(Equal(annotations))
			spec := *upCrd.(*crdv1.Schema).Spec
			// removed parts
			Expect(spec["name"]).To(BeNil())
			Expect(spec["metadata"]).To(BeNil())
			Expect(spec["status"]).To(BeNil())
			Expect(spec["annotations"]).To(BeNil())

			// shifted parts
			Expect(spec["inline_schema"]).To(Equal(schema.InlineSchema))
		})
	})
	Describe("ResolverMapToCrd", func() {
		It("Converts a gloo resolverMap to crd", func() {
			resolverMap := NewTestResolverMap("foo")
			annotations := map[string]string{"foo": "bar"}
			resolverMap.Metadata = &gloov1.Metadata{
				Annotations: annotations,
			}
			resolverMapCrd, err := ConfigObjectToCrd("foo", resolverMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(resolverMapCrd.GetName()).To(Equal(resolverMap.Name))
			Expect(resolverMapCrd.GetNamespace()).To(Equal("foo"))
			Expect(resolverMapCrd.GetAnnotations()).To(Equal(annotations))
			spec := *resolverMapCrd.(*crdv1.ResolverMap).Spec
			// removed parts
			Expect(spec["name"]).To(BeNil())
			Expect(spec["metadata"]).To(BeNil())
			Expect(spec["status"]).To(BeNil())
			Expect(spec["annotations"]).To(BeNil())

		})
	})
})
