package reporter_test

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/solo-io/qloo/pkg/storage"
	"github.com/solo-io/qloo/pkg/storage/crd"
	"k8s.io/client-go/tools/clientcmd"

	. "github.com/solo-io/qloo/pkg/reporter"
	"github.com/solo-io/qloo/pkg/api/types/v1"
	gloov1 "github.com/solo-io/gloo/pkg/api/types/v1"
	"github.com/solo-io/gloo/pkg/log"
	. "github.com/solo-io/gloo/test/helpers"
	"github.com/solo-io/qloo/test"
)

var _ = Describe("CrdReporter", func() {
	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}
	var (
		masterUrl, kubeconfigPath string
		namespace                 string
		rptr                      Interface
	)
	BeforeEach(func() {
		namespace = RandString(8)
		err := SetupKubeForTest(namespace)
		Must(err)
		kubeconfigPath = filepath.Join(os.Getenv("HOME"), ".kube", "config")
		masterUrl = ""
	})
	AfterEach(func() {
		TeardownKube(namespace)
	})
	Describe("writereports", func() {
		var (
			glooClient   storage.Interface
			reports      []ConfigObjectReport
			schemas      []*v1.Schema
			resolverMaps []*v1.ResolverMap
		)
		Context("writes status reports for cfg crds with 0 errors", func() {
			BeforeEach(func() {
				reports = nil
				cfg, err := clientcmd.BuildConfigFromFlags(masterUrl, kubeconfigPath)
				Expect(err).NotTo(HaveOccurred())
				glooClient, err = crd.NewStorage(cfg, namespace, time.Second)
				Expect(err).NotTo(HaveOccurred())
				rptr = NewReporter(glooClient)

				testCfg := newTestConfig()
				schemas = testCfg.Schemas
				var storables []gloov1.ConfigObject
				for _, us := range schemas {
					_, err := glooClient.V1().Schemas().Create(us)
					Expect(err).NotTo(HaveOccurred())
					storables = append(storables, us)
				}
				resolverMaps = testCfg.ResolverMaps
				for _, vService := range resolverMaps {
					_, err := glooClient.V1().ResolverMaps().Create(vService)
					Expect(err).NotTo(HaveOccurred())
					storables = append(storables, vService)
				}
				for _, storable := range storables {
					reports = append(reports, ConfigObjectReport{
						CfgObject: storable,
						Err:       nil,
					})
				}
			})

			It("writes an acceptance status for each crd", func() {
				err := rptr.WriteReports(reports)
				Expect(err).NotTo(HaveOccurred())
				updatedSchemas, err := glooClient.V1().Schemas().List()
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedSchemas).To(HaveLen(len(schemas)))
				for _, updatedSchema := range updatedSchemas {
					Expect(updatedSchema.Status.State).To(Equal(gloov1.Status_Accepted))
				}
				updatedvServices, err := glooClient.V1().ResolverMaps().List()
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedvServices).To(HaveLen(len(schemas)))
				for _, updatedvService := range updatedvServices {
					Expect(updatedvService.Status.State).To(Equal(gloov1.Status_Accepted))
				}
			})
		})
		Context("writes status reports for cfg crds with SOME errors", func() {
			BeforeEach(func() {
				reports = nil
				cfg, err := clientcmd.BuildConfigFromFlags(masterUrl, kubeconfigPath)
				Expect(err).NotTo(HaveOccurred())
				glooClient, err = crd.NewStorage(cfg, namespace, time.Second)
				Expect(err).NotTo(HaveOccurred())
				rptr = NewReporter(glooClient)

				testCfg := newTestConfig()
				schemas = testCfg.Schemas
				var storables []gloov1.ConfigObject
				for _, us := range schemas {
					_, err := glooClient.V1().Schemas().Create(us)
					Expect(err).NotTo(HaveOccurred())
					storables = append(storables, us)
				}
				resolverMaps = testCfg.ResolverMaps
				for _, vService := range resolverMaps {
					_, err := glooClient.V1().ResolverMaps().Create(vService)
					Expect(err).NotTo(HaveOccurred())
					storables = append(storables, vService)
				}
				for _, storable := range storables {
					reports = append(reports, ConfigObjectReport{
						CfgObject: storable,
						Err:       errors.New("oh no an error what did u do!"),
					})
				}
			})

			It("writes an rejected status for each crd", func() {
				err := rptr.WriteReports(reports)
				Expect(err).NotTo(HaveOccurred())
				updatedSchemas, err := glooClient.V1().Schemas().List()
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedSchemas).To(HaveLen(len(schemas)))
				for _, updatedSchema := range updatedSchemas {
					Expect(updatedSchema.Status.State).To(Equal(gloov1.Status_Rejected))
				}
				updatedvServices, err := glooClient.V1().ResolverMaps().List()
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedvServices).To(HaveLen(len(schemas)))
				for _, updatedvService := range updatedvServices {
					Expect(updatedvService.Status.State).To(Equal(gloov1.Status_Rejected))
				}
			})
		})
	})
})

func newTestConfig() *v1.Config {
	return &v1.Config{
		Schemas:      []*v1.Schema{test.StarWarsV1Schema()},
		ResolverMaps: []*v1.ResolverMap{test.StarWarsResolverMap()},
	}
}
