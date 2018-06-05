package gloo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/solo-io/qloo/pkg/resolvers/gloo"
	"github.com/solo-io/qloo/pkg/api/types/v1"
	"net/http/httptest"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
	"github.com/solo-io/qloo/test"
	"io"
	"bytes"
)

var _ = Describe("GlooResolvers", func() {
	var (
		mockProxyAddr   string
		server          *httptest.Server
		response        = []byte(`{"have":"a","nice":"day","okay":"?"}`)
		resolverFactory *ResolverFactory
		buf             *bytes.Buffer
	)
	BeforeEach(func() {
		buf = &bytes.Buffer{}
		m := mux.NewRouter()
		m.HandleFunc("/mytype.myfield", func(w http.ResponseWriter, r *http.Request) {
			io.Copy(buf, r.Body)
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
		path := ResolverPath{
			TypeName:  "mytype",
			FieldName: "myfield",
		}
		gResolver := &v1.GlooResolver{
			RequestTemplate:  "{{ marshal . }}",
			ResponseTemplate: "{{ marshal . }}",
		}
		It("creates the resolver without error", func() {
			_, err := resolverFactory.CreateResolver(path, gResolver)
			Expect(err).NotTo(HaveOccurred())
		})
		It("calls the resolver without error", func() {
			rawResolver, err := resolverFactory.CreateResolver(path, gResolver)
			Expect(err).NotTo(HaveOccurred())
			b, err := rawResolver(test.LukeSkywalkerParams)
			Expect(err).NotTo(HaveOccurred())
			Expect(b).To(Equal([]byte(`{"Result":{"have":"a","nice":"day","okay":"?"}}`)))
			Expect(buf.String()).To(Equal(`{"Args":{"acting":5,"best_scene":"cloud city"},` +
				`"Parent":{"CharacterFields":{"AppearsIn":["NEWHOPE","EMPIRE","JEDI"],` +
				`"FriendIds":["1002","1003","2000","2001"],"ID":"1000","Name":"Luke Skywalker",` +
				`"TypeName":"Human"},"Mass":77,"StarshipIds":["3001","3003"],"appearsIn":null,` +
				`"friends":null,"friendsConnection":null,"height":null,"id":null,"mass":null,` +
				`"name":null,"starships":null}}`))
		})
	})
})
