package template_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/solo-io/sqoop/pkg/api/types/v1"
	. "github.com/solo-io/sqoop/pkg/resolvers/template"
	"github.com/solo-io/sqoop/test"
)

var _ = Describe("TemplateResolvers", func() {
	Context("happy path with simple template and params", func() {
		tResolver := &v1.TemplateResolver{
			InlineTemplate: "{{ marshal . }}",
		}
		It("returns a resolver which renders the template", func() {
			rawResolver, err := NewTemplateResolver(tResolver)
			Expect(err).NotTo(HaveOccurred())
			b, err := rawResolver(test.LukeSkywalkerParams)
			Expect(err).NotTo(HaveOccurred())
			Expect(b).To(Equal([]byte(`{"Args":{"acting":5,"best_scene":"cloud city"},` +
				`"Parent":{"CharacterFields":{"AppearsIn":["NEWHOPE","EMPIRE","JEDI"],` +
				`"FriendIds":["1002","1003","2000","2001"],"ID":"1000","Name":"Luke Skywalker","TypeName":"Human"},` +
				`"Mass":77,"StarshipIds":["3001","3003"],"appearsIn":null,"friends":null,"friendsConnection":null,` +
				`"height":null,"id":null,"mass":null,"name":null,"starships":null}}`)))
		})
	})
})
