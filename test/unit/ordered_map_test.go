package unit_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/solo-io/sqoop/pkg/engine/dynamic"
	"github.com/vektah/gqlgen/neelance/schema"
)

var _ = Describe("Values", func() {
	Describe("OrderedMap", func() {
		var om *dynamic.OrderedMap
		BeforeEach(func() {
			om = dynamic.NewOrderedMap()
		})
		It("gets by key", func() {
			val := om.Get("foo")
			Expect(val).To(BeNil())
			om.Keys = append(om.Keys, "foo")
			expected := &dynamic.String{Scalar: &schema.Scalar{Name: "String"}, Data: "bar"}
			om.Values = append(om.Values, expected)
			val = om.Get("foo")
			Expect(val).To(Equal(expected))
		})
		It("deletes by key", func() {
			om.Keys = append(om.Keys, "foo")
			expected := &dynamic.String{Scalar: &schema.Scalar{Name: "String"}, Data: "bar"}
			om.Values = append(om.Values, expected)
			val := om.Get("foo")
			Expect(val).To(Equal(expected))
			om.Delete("foo")
			val = om.Get("foo")
			Expect(val).To(BeNil())
		})
		It("inserts key/val", func() {
			val := om.Get("foo")
			Expect(val).To(BeNil())
			expected := &dynamic.String{Scalar: &schema.Scalar{Name: "String"}, Data: "bar"}
			om.Set("foo", expected)
			val = om.Get("foo")
			Expect(val).To(Equal(expected))
		})
	})
})
