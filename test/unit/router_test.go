package unit_test

//import (
//	. "github.com/onsi/ginkgo"
//	. "github.com/onsi/gomega"
//
//	"bytes"
//	"io/ioutil"
//	"net/http"
//	"net/http/httptest"
//
//	"github.com/solo-io/sqoop/test/testdata"
//)
//
//var _ = Describe("Router", func() {
//	var (
//		router *Router
//		server *httptest.Server
//	)
//	BeforeEach(func() {
//		router = NewRouter()
//		server = httptest.NewServer(router)
//	})
//	AfterEach(func() {
//		server.Close()
//	})
//	It("serves and updates routes dynamically from graphql endpoints", func() {
//		testEndpoints := []*Endpoint{
//			{
//				SchemaName: "StarWars1",
//				RootPath:   "/root1",
//				QueryPath:  "/query2",
//				ExecSchema: testdata.StarWarsExecutableSchema("no-address-defined"),
//			},
//			{
//				SchemaName: "StarWars2",
//				RootPath:   "/root2",
//				QueryPath:  "/query2",
//				ExecSchema: testdata.StarWarsExecutableSchema("no-address-defined"),
//			},
//		}
//		router.UpdateEndpoints(testEndpoints...)
//		for _, ep := range testEndpoints {
//			res, err := http.Get(server.URL + ep.RootPath)
//			Expect(err).NotTo(HaveOccurred())
//			data, err := ioutil.ReadAll(res.Body)
//			Expect(err).NotTo(HaveOccurred())
//			Expect(string(data)).To(ContainSubstring(ep.SchemaName))
//			Expect(string(data)).To(ContainSubstring(ep.QueryPath))
//			res, err = http.Post(server.URL+ep.QueryPath, "", bytes.NewBuffer(queryString))
//			Expect(err).NotTo(HaveOccurred())
//			data, err = ioutil.ReadAll(res.Body)
//			Expect(err).NotTo(HaveOccurred())
//			Expect(string(data)).To(ContainSubstring(`{"data":{"hero":null},"errors":` +
//				`[{"message":"executing resolver for field \"hero\": failed executing resolver for Query.hero: ` +
//				`performing http post: Post http://no-address-defined/Query.hero: dial tcp: lookup no-address-defined on`))
//		}
//	})
//})
//
//var queryString = []byte(`{"query": "{hero{name}}"}`)
