package dynamic_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/solo-io/qloo/pkg/dynamic"
	"github.com/solo-io/qloo/test"
)

var _ = Describe("Resolver", func() {
	r := NewResolverMap(test.StarWarsSchema, inputResolvers)
	queryObject := test.StarWarsSchema.Types["Query"]
	Context("Resolve", func() {
		It("calls the provided resolver for the field on the type", func() {
			Expect(r.Types).To(HaveKey(queryObject))
			Expect(r.Types[queryObject]).NotTo(BeNil())
			Expect(r.Types[queryObject].Fields).To(HaveKey("hero"))
			Expect(r.Resolve(queryObject, "hero", nil)).To(Equal(map[string]string{"name": "Luke"}))
		})
		It("returns error for nonexistent type", func() {
			result, err := r.Resolve(test.StarWarsSchema.Types["__Type"], "nonexistent", nil)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("type __Type unknown"))
			Expect(result).To(BeNil())
		})
		It("returns error for nonexistent field on type", func() {
			result, err := r.Resolve(queryObject, "nonexistent", nil)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("type Query does not contain field nonexistent"))
			Expect(result).To(BeNil())
		})
	})
})

var inputResolvers = map[string]ResolverFunc{
	"Query.hero": func(params *Params) (interface{}, error) {
		return map[string]string{"name": "Luke"}, nil
	},
}
