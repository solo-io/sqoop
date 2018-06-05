package dynamic_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/solo-io/qloo/pkg/dynamic"
	"github.com/solo-io/qloo/test"
	"github.com/vektah/gqlgen/example/starwars"
	"context"
)

var _ = Describe("Resolver", func() {
	r := NewExecutableResolvers(test.StarWarsSchema, starWarsResolvers)
	queryObject := test.StarWarsSchema.Types["Query"]
	Context("Resolve", func() {
		It("calls the provided resolver for the field on the type", func() {
			Expect(r.Types).To(HaveKey(queryObject))
			Expect(r.Types[queryObject]).NotTo(BeNil())
			Expect(r.Types[queryObject].Fields).To(HaveKey("hero"))
			p := Params{Args: map[string]interface{}{
				"id": "1000",
			}}
			res, err := r.Resolve(queryObject, "hero", p)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeAssignableToTypeOf(&starwars.Human{}))
			Expect(res.(*starwars.Human).Name).To(Equal("Luke Skywalker"))
			Expect(res.(*starwars.Human).ID).To(Equal("1000"))
			Expect(res.(*starwars.Human).FriendIds).To(Equal([]string{"1002", "1003", "2000", "2001"}))
		})
		It("returns error for nonexistent type", func() {
			result, err := r.Resolve(test.StarWarsSchema.Types["__Type"], "nonexistent", Params{})
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("type __Type unknown"))
			Expect(result).To(BeNil())
		})
		It("returns error for nonexistent field on type", func() {
			result, err := r.Resolve(queryObject, "nonexistent", Params{})
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("type Query does not contain field nonexistent"))
			Expect(result).To(BeNil())
		})
	})
})

var baseResolvers = starwars.NewResolver()

var starWarsResolvers = map[string]ResolverFunc{
	"Query.hero": func(params Params) (interface{}, error) {
		return baseResolvers.Query_character(context.TODO(), params.Arg("id").(string))
	},
}
