package gloo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gorilla/mux"
	"github.com/solo-io/sqoop/pkg/api/types/v1"
	. "github.com/solo-io/sqoop/pkg/resolvers/gloo"
	"github.com/solo-io/sqoop/test"
)

var _ = Describe("GlooResolvers", func() {
	var (
		mockProxyAddr   string
		server          *httptest.Server
		response        = []byte(`{"have":"a","nice":"day","okay":"?"}`)
		resolverFactory *ResolverFactory
		requestBody     *bytes.Buffer
	)
	BeforeEach(func() {
		requestBody = &bytes.Buffer{}
		m := mux.NewRouter()
		m.HandleFunc("/mytype.myfield", func(w http.ResponseWriter, r *http.Request) {
			io.Copy(requestBody, r.Body)
			w.Write(response)
		})
		server = httptest.NewServer(m)
		mockProxyAddr = strings.TrimPrefix(server.URL, "http://")

		resolverFactory = NewResolverFactory(mockProxyAddr)
	})
	AfterEach(func() {
		server.Close()
	})
	Context("happy path with req+response template and params", func() {
		typeName := "mytype"
		fieldName := "myfield"
		gResolver := &v1.GlooResolver{
			RequestTemplate:  `REQUEST: best scene: {{ marshal (index .Args "best_scene") }} friendIds: {{ marshal (index .Parent "CharacterFields") }}`,
			ResponseTemplate: `RESPONSE: {{ marshal (index .Result "nice") }}`,
		}
		Context("it returns a resolver which ", func() {
			It("renders the template as the request body", func() {
				rawResolver, err := resolverFactory.CreateResolver(typeName, fieldName, gResolver)
				Expect(err).NotTo(HaveOccurred())
				_, err = rawResolver(test.LukeSkywalkerParams)
				Expect(err).NotTo(HaveOccurred())
				str := requestBody.String()
				Expect(str).To(Equal(`REQUEST: best scene: "cloud city" friendIds: ` +
					`{"AppearsIn":["NEWHOPE","EMPIRE","JEDI"],"FriendIds":["1002","1003","2000","2001"],` +
					`"ID":"1000","Name":"Luke Skywalker","TypeName":"Human"}`))
			})
			It("renders the result template on the json response body", func() {
				rawResolver, err := resolverFactory.CreateResolver(typeName, fieldName, gResolver)
				Expect(err).NotTo(HaveOccurred())
				b, err := rawResolver(test.LukeSkywalkerParams)
				Expect(err).NotTo(HaveOccurred())
				Expect(b).To(Equal([]byte(`RESPONSE: "day"`)))
			})
		})
	})
})
