package unit_test


import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gorilla/mux"
	. "github.com/solo-io/sqoop/pkg/engine/exec"
	"github.com/solo-io/sqoop/pkg/engine/resolvers"
	"github.com/solo-io/sqoop/test/testdata"
)

var _ = Describe("ExecutableResolverMap", func() {
	var (
		proxyAddr      string
		server         *httptest.Server
		response, _    = json.Marshal(testdata.LukeSkywalker)
		requestBody    *bytes.Buffer
		createResolver func(typeName, fieldName string) (RawResolver, error)
	)
	BeforeEach(func() {
		requestBody = &bytes.Buffer{}
		m := mux.NewRouter()
		m.HandleFunc("/Query.hero", func(w http.ResponseWriter, r *http.Request) {
			io.Copy(requestBody, r.Body)
			w.Write(response)
		})
		m.HandleFunc("/default.starwars-resolvers.Query.hero", func(w http.ResponseWriter, r *http.Request) {
			io.Copy(requestBody, r.Body)
			w.Write(response)
		})
		server = httptest.NewServer(m)
		proxyAddr = strings.TrimPrefix(server.URL, "http://")

		resolverFactory := resolvers.NewResolverFactory(proxyAddr, testdata.StarWarsResolverMap())
		createResolver = resolverFactory.CreateResolver
	})
	AfterEach(func() {
		server.Close()
	})
	It("does the happy path", func() {
		execResolve, err := NewExecutableResolvers(testdata.StarWarsSchema, createResolver)
		Expect(err).NotTo(HaveOccurred())
		res, err := execResolve.Resolve(testdata.StarWarsSchema.Types["Query"], "hero", Params{})
		Expect(err).NotTo(HaveOccurred())
		data, ok := res.GoValue().(map[string]interface{})
		Expect(ok).To(BeTrue())
		m, _ := toMap(testdata.LukeSkywalker)
		// all the keys match
		for k, v := range m {
			Expect(data).To(HaveKey(k))
			Expect(data[k]).To(Equal(v))
			delete(data, k)
		}
		// extra fields are nil
		for k, v := range data {
			Expect(v).To(BeNil())
			Expect(m[k]).To(BeNil())
		}
	})
})

func toMap(v interface{}) (map[string]interface{}, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	err = json.Unmarshal(b, &m)
	return m, err
}
