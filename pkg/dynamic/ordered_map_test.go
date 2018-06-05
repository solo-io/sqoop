package dynamic_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/solo-io/qloo/pkg/dynamic"
	"github.com/vektah/gqlgen/neelance/schema"
)

var _ = Describe("Values", func() {
	Describe("OrderedMap", func() {
		var om *OrderedMap
		BeforeEach(func() {
			om = NewOrderedMap()
		})
		It("gets by key", func() {
			val := om.Get("foo")
			Expect(val).To(BeNil())
			om.Keys = append(om.Keys, "foo")
			expected := &String{Scalar: &schema.Scalar{Name: "String"}, Data: "bar" }
			om.Values = append(om.Values, expected)
			val = om.Get("foo")
			Expect(val).To(Equal(expected))
		})
		It("deletes by key", func() {
			om.Keys = append(om.Keys, "foo")
			expected := &String{Scalar: &schema.Scalar{Name: "String"}, Data: "bar" }
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
			expected := &String{Scalar: &schema.Scalar{Name: "String"}, Data: "bar" }
			om.Set("foo", expected)
			val = om.Get("foo")
			Expect(val).To(Equal(expected))
		})
	})
})
