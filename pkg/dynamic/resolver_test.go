package dynamic_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/solo-io/qloo/pkg/dynamic"
	"github.com/solo-io/qloo/test"
)

var _ = Describe("Resolver", func() {
	It("gets the correct key from the schema", func() {
		r := NewResolverMap(test.StarWarsSchema, inputResolvers)
		queryObject := test.StarWarsSchema.Types["Query"]
		Expect(r.Types).To(HaveKey(queryObject))
		Expect(r.Types[queryObject]).NotTo(BeNil())
		Expect(r.Types[queryObject].Fields).To(HaveKey("hero"))
		Expect(r.Resolve(queryObject, "hero", nil)).To(Equal(map[string]string{"name": "Luke"}))
		Expect(r.Resolve(nil, "nonexistent", nil)).To(BeNil())
		Expect(r.Resolve(queryObject, "nonexistent", nil)).To(BeNil())
	})
})

var inputResolvers = map[string]ResolverFunc{
	"Query.hero": func(args map[string]interface{}) (interface{}, error) {
		return map[string]string{"name": "Luke"}, nil
	},
}
