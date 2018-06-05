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
	"io/ioutil"
)

var _ = Describe("GlooResolvers", func() {
	var (
		mockProxyAddr string
		server        *httptest.Server
		received      chan *http.Request
		response      = []byte(`{"have":"a","nice":"day","okay":"?"`)
	)
	BeforeEach(func() {
		received = make(chan *http.Request)
		m := mux.NewRouter()
		m.HandleFunc("/mytype.myfield", func(w http.ResponseWriter, r *http.Request) {
			go func() { received <- r }()
			w.Write(response)
		})
		server = httptest.NewServer(m)
		mockProxyAddr = strings.TrimPrefix(server.URL, "http://")
	})
	AfterEach(func() {
		server.Close()
	})
	proxyAddr := "doesnt-exist"
	resolverFactory := NewResolverFactory(proxyAddr)
	Context("happy path with req+response template and params", func() {
		path := ResolverPath{
			TypeName:  "mytype",
			FieldName: "myfield",
		}
		gResolver := &v1.GlooResolver{
			RequestTemplate:  "Params: {{ marshal . }}",
			ResponseTemplate: "{{  }}",
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
			Expect(b).To(Equal("foo"))
			var r *http.Request
			Eventually(func() *http.Request {
				select {
				case r = <-received:
					return r
				default:
					return nil
				}
			})
			Expect(r).To(Equal("foo"))
			b, err = ioutil.ReadAll(r.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(b).To(Equal("foo"))
		})
	})
})
