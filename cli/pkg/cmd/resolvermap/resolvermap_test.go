package resolvermap_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	glooHelpers "github.com/solo-io/gloo/projects/gloo/cli/pkg/helpers"
	glooV1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/sqoop/cli/pkg/helpers"
	"github.com/solo-io/sqoop/pkg/api/v1"
	"github.com/solo-io/sqoop/pkg/defaults"
	"sigs.k8s.io/yaml"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/sqoop/cli/pkg/testutils"
)

var _ = Describe("Schema", func() {

	registerResolverMap := func() {
		client := helpers.MustResolverMapClient()
		var rm v1.ResolverMap
		err := yaml.Unmarshal([]byte(resolverMap), &rm)
		Expect(err).NotTo(HaveOccurred())
		rm.Types = map[string]*v1.TypeResolver{
			"Query": &v1.TypeResolver{
				Fields: map[string]*v1.FieldResolver{
					"pet":  nil,
					"pets": nil,
				},
			},
			"Mutation": &v1.TypeResolver{
				Fields: map[string]*v1.FieldResolver{
					"addPet": nil,
				},
			},
			"Pet": &v1.TypeResolver{
				Fields: map[string]*v1.FieldResolver{
					"id":   nil,
					"name": nil,
				},
			},
		}
		_, err = client.Write(&rm, clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
	}

	registerUpstream := func() {
		client := glooHelpers.MustUpstreamClient()
		var us glooV1.Upstream
		err := yaml.Unmarshal([]byte(petstoreUpstream), &us)
		Expect(err).NotTo(HaveOccurred())
		_, err = client.Write(&us, clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
	}

	BeforeEach(func() {
		// create proper resolver map because event loop is not running
		helpers.UseMemoryClients()
		glooHelpers.UseMemoryClients()
		registerResolverMap()
	})

	getResolverMap := func(name string) *v1.ResolverMap {
		rm, err := helpers.MustResolverMapClient().Read(defaults.SqoopSystem, name, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())
		return rm
	}

	Context("register", func() {
		It("should register properly", func() {
			registerUpstream()
			err := testutils.Sqoopctl(fmt.Sprintf("resolvermap register -u %s -g %s -s %s Query pets", upstreamName, upstreamFunctionName, schemaName))
			Expect(err).NotTo(HaveOccurred())
			getResolverMap(schemaName)
		})

		It("should error when schema not provided", func() {
			err := testutils.Sqoopctl(fmt.Sprintf("resolvermap register -u %s -g %s Query pets", upstreamName, upstreamFunctionName))
			Expect(err).To(HaveOccurred())
		})

		It("should error when upstream not provided", func() {
			err := testutils.Sqoopctl(fmt.Sprintf("resolvermap register -g %s -s %s Query pets", upstreamFunctionName, schemaName))
			Expect(err).To(HaveOccurred())
		})

		It("should error when function not provided", func() {
			err := testutils.Sqoopctl(fmt.Sprintf("resolvermap register -u %s -s %s Query pets", upstreamName, schemaName))
			Expect(err).To(HaveOccurred())
		})

		It("should error when # of args incorrect", func() {
			err := testutils.Sqoopctl(fmt.Sprintf("resolvermap register -u %s -g %s -s %s Query", upstreamName, upstreamFunctionName, schemaName))
			Expect(err).To(HaveOccurred())
		})

		It("should error when type and field don't match resolver map", func() {
			err := testutils.Sqoopctl(fmt.Sprintf("resolvermap register -u %s -g %s -s %s Query test", upstreamName, upstreamFunctionName, schemaName))
			Expect(err).To(HaveOccurred())
		})
	})

})

const schemaName = "one"
const resolverMap = `
apiVersion: sqoop.solo.io/v1
kind: ResolverMap
metadata:
  annotations:
    created_for: one
  name: one
  namespace: gloo-system
spec:
  types:
    Mutation:
      fields:
        addPet: {}
    Pet:
      fields:
        id: {}
        name: {}
    Query:
      fields:
        pet: {}
        pets: {}
`
const upstreamName = "gloo-system-petstore-8080"
const upstreamFunctionName = "findPets"
const petstoreUpstream = `
apiVersion: gloo.solo.io/v1
kind: Upstream
metadata:
  labels:
    service: petstore
  name: gloo-system-petstore-8080
  namespace: gloo-system
spec:
  discoveryMetadata: {}
  upstreamSpec:
    kube:
      selector:
        app: petstore
      serviceName: petstore
      serviceNamespace: gloo-system
      servicePort: 8080
      serviceSpec:
        rest:
          swaggerInfo:
            url: http://petstore.gloo-system.svc.cluster.local:8080/swagger.json
          transformations:
            addPet:
              body:
                text: '{"id": {{ default(id, "") }},"name": "{{ default(name, "")}}","tag":
                  "{{ default(tag, "")}}"}'
              headers:
                :method:
                  text: POST
                :path:
                  text: /api/pets
                content-type:
                  text: application/json
            deletePet:
              headers:
                :method:
                  text: DELETE
                :path:
                  text: /api/pets/{{ default(id, "") }}
                content-type:
                  text: application/json
            findPetById:
              body: {}
              headers:
                :method:
                  text: GET
                :path:
                  text: /api/pets/{{ default(id, "") }}
                content-length:
                  text: "0"
                content-type: {}
                transfer-encoding: {}
            findPets:
              body: {}
              headers:
                :method:
                  text: GET
                :path:
                  text: /api/pets?tags={{default(tags, "")}}&limit={{default(limit,
                    "")}}
                content-length:
                  text: "0"
                content-type: {}
                transfer-encoding: {}
`
