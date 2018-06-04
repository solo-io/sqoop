package crd_test

import (
	"os"
	"path/filepath"

	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo/pkg/log"
	. "github.com/solo-io/qloo/pkg/storage/crd"
	crdv1 "github.com/solo-io/qloo/pkg/storage/crd/solo.io/v1"
	. "github.com/solo-io/gloo/test/helpers"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"github.com/solo-io/qloo/pkg/api/types/v1"
	gloov1 "github.com/solo-io/gloo/pkg/api/types/v1"
)

var _ = Describe("CrdStorageClient", func() {
	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}
	var (
		masterUrl, kubeconfigPath string
		namespace                 string
		syncFreq                  = time.Minute
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
	Describe("New", func() {
		It("creates a new client without error", func() {
			cfg, err := clientcmd.BuildConfigFromFlags(masterUrl, kubeconfigPath)
			Expect(err).NotTo(HaveOccurred())
			_, err = NewStorage(cfg, namespace, syncFreq)
			Expect(err).NotTo(HaveOccurred())
		})
	})
	Describe("Register", func() {
		It("registers the crds", func() {
			cfg, err := clientcmd.BuildConfigFromFlags(masterUrl, kubeconfigPath)
			Expect(err).NotTo(HaveOccurred())
			client, err := NewStorage(cfg, namespace, syncFreq)
			Expect(err).NotTo(HaveOccurred())
			err = client.V1().Register()
			Expect(err).NotTo(HaveOccurred())
			apiextClient, err := apiexts.NewForConfig(cfg)
			Expect(err).NotTo(HaveOccurred())
			crds, err := apiextClient.ApiextensionsV1beta1().CustomResourceDefinitions().List(metav1.ListOptions{})
			Expect(err).NotTo(HaveOccurred())
			for _, crdSchema := range crdv1.KnownCRDs {
				var foundCrd *v1beta1.CustomResourceDefinition
				for _, crd := range crds.Items {
					if crd.Spec.Names.Kind == crdSchema.Kind {
						foundCrd = &crd
						break
					}
				}
				// if crd wasnt found, err
				Expect(foundCrd).NotTo(BeNil())

				Expect(foundCrd.Spec.Version).To(Equal(crdSchema.Version))
				Expect(foundCrd.Spec.Group).To(Equal(crdSchema.Group))
			}
		})
	})
	Describe("schemas", func() {
		Describe("Create", func() {
			It("creates a crd from the item", func() {
				cfg, err := clientcmd.BuildConfigFromFlags(masterUrl, kubeconfigPath)
				Expect(err).NotTo(HaveOccurred())
				client, err := NewStorage(cfg, namespace, syncFreq)
				Expect(err).NotTo(HaveOccurred())
				err = client.V1().Register()
				Expect(err).NotTo(HaveOccurred())
				schema := NewTestSchema1()
				createdSchema, err := client.V1().Schemas().Create(schema)
				Expect(err).NotTo(HaveOccurred())
				schema.Metadata = createdSchema.GetMetadata()
				Expect(schema).To(Equal(createdSchema))
			})
		})
		Describe("Get", func() {
			It("gets a crd from the name", func() {
				cfg, err := clientcmd.BuildConfigFromFlags(masterUrl, kubeconfigPath)
				Expect(err).NotTo(HaveOccurred())
				client, err := NewStorage(cfg, namespace, syncFreq)
				Expect(err).NotTo(HaveOccurred())
				err = client.V1().Register()
				Expect(err).NotTo(HaveOccurred())
				schema := NewTestSchema1()
				_, err = client.V1().Schemas().Create(schema)
				Expect(err).NotTo(HaveOccurred())
				created, err := client.V1().Schemas().Get(schema.Name)
				Expect(err).NotTo(HaveOccurred())
				schema.Metadata = created.Metadata
				Expect(created).To(Equal(schema))
			})
		})
		Describe("Update", func() {
			It("updates a crd from the item", func() {
				cfg, err := clientcmd.BuildConfigFromFlags(masterUrl, kubeconfigPath)
				Expect(err).NotTo(HaveOccurred())
				client, err := NewStorage(cfg, namespace, syncFreq)
				Expect(err).NotTo(HaveOccurred())
				err = client.V1().Register()
				Expect(err).NotTo(HaveOccurred())
				schema := NewTestSchema1()
				created, err := client.V1().Schemas().Create(schema)
				Expect(err).NotTo(HaveOccurred())
				schema.Metadata = created.GetMetadata()
				schema.ResolverMap = "something-else"
				schema.Metadata.Annotations["just_for_this_test"] = "bar"
				updated, err := client.V1().Schemas().Update(schema)
				Expect(err).NotTo(HaveOccurred())
				Expect(updated.Metadata.Annotations).To(HaveKey("just_for_this_test"))
				schema.Metadata = updated.GetMetadata()
				Expect(updated).To(Equal(schema))
			})
		})
		Describe("Delete", func() {
			It("deletes a crd from the name", func() {
				cfg, err := clientcmd.BuildConfigFromFlags(masterUrl, kubeconfigPath)
				Expect(err).NotTo(HaveOccurred())
				client, err := NewStorage(cfg, namespace, syncFreq)
				Expect(err).NotTo(HaveOccurred())
				err = client.V1().Register()
				Expect(err).NotTo(HaveOccurred())
				schema := NewTestSchema1()
				_, err = client.V1().Schemas().Create(schema)
				Expect(err).NotTo(HaveOccurred())
				err = client.V1().Schemas().Delete(schema.Name)
				Expect(err).NotTo(HaveOccurred())
				_, err = client.V1().Schemas().Get(schema.Name)
				Expect(err).To(HaveOccurred())
			})
		})
	})
	Describe("resolverMaps", func() {
		Describe("Create", func() {
			It("creates a crd from the item", func() {
				cfg, err := clientcmd.BuildConfigFromFlags(masterUrl, kubeconfigPath)
				Expect(err).NotTo(HaveOccurred())
				client, err := NewStorage(cfg, namespace, syncFreq)
				Expect(err).NotTo(HaveOccurred())
				err = client.V1().Register()
				Expect(err).NotTo(HaveOccurred())
				resolverMap := NewTestResolverMap("something")
				createdSchema, err := client.V1().ResolverMaps().Create(resolverMap)
				Expect(err).NotTo(HaveOccurred())
				resolverMap.Metadata = createdSchema.GetMetadata()
				Expect(resolverMap).To(Equal(createdSchema))
			})
		})
		Describe("Get", func() {
			It("gets a crd from the name", func() {
				cfg, err := clientcmd.BuildConfigFromFlags(masterUrl, kubeconfigPath)
				Expect(err).NotTo(HaveOccurred())
				client, err := NewStorage(cfg, namespace, syncFreq)
				Expect(err).NotTo(HaveOccurred())
				err = client.V1().Register()
				Expect(err).NotTo(HaveOccurred())
				resolverMap := NewTestResolverMap("something")
				_, err = client.V1().ResolverMaps().Create(resolverMap)
				Expect(err).NotTo(HaveOccurred())
				created, err := client.V1().ResolverMaps().Get(resolverMap.Name)
				Expect(err).NotTo(HaveOccurred())
				resolverMap.Metadata = created.Metadata
				Expect(created).To(Equal(resolverMap))
			})
		})
		Describe("Update", func() {
			It("updates a crd from the item", func() {
				cfg, err := clientcmd.BuildConfigFromFlags(masterUrl, kubeconfigPath)
				Expect(err).NotTo(HaveOccurred())
				client, err := NewStorage(cfg, namespace, syncFreq)
				Expect(err).NotTo(HaveOccurred())
				err = client.V1().Register()
				Expect(err).NotTo(HaveOccurred())
				resolverMap := NewTestResolverMap("something")
				created, err := client.V1().ResolverMaps().Create(resolverMap)
				Expect(err).NotTo(HaveOccurred())
				// need to set resource ver
				resolverMap.Metadata = created.GetMetadata()
				resolverMap.Metadata.Annotations["just_for_this_test"] = "bar"
				updated, err := client.V1().ResolverMaps().Update(resolverMap)
				Expect(err).NotTo(HaveOccurred())
				Expect(updated.Metadata.Annotations).To(HaveKey("just_for_this_test"))
				resolverMap.Metadata = updated.GetMetadata()
				Expect(updated).To(Equal(resolverMap))
			})
		})
		Describe("Delete", func() {
			It("deletes a crd from the name", func() {
				cfg, err := clientcmd.BuildConfigFromFlags(masterUrl, kubeconfigPath)
				Expect(err).NotTo(HaveOccurred())
				client, err := NewStorage(cfg, namespace, syncFreq)
				Expect(err).NotTo(HaveOccurred())
				err = client.V1().Register()
				Expect(err).NotTo(HaveOccurred())
				resolverMap := NewTestResolverMap("something")
				_, err = client.V1().ResolverMaps().Create(resolverMap)
				Expect(err).NotTo(HaveOccurred())
				err = client.V1().ResolverMaps().Delete(resolverMap.Name)
				Expect(err).NotTo(HaveOccurred())
				_, err = client.V1().ResolverMaps().Get(resolverMap.Name)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})

func NewTestSchema1() *v1.Schema{ 
	return &v1.Schema{
		Name: "schema1",
		ResolverMap: "resolvers",
		InlineSchema: "SOMETHING",
		Metadata: &gloov1.Metadata{
			Annotations: map[string]string{
				"foo": "bar",
			},
		},
	}
}
func NewTestResolverMap(name string) *v1.ResolverMap{
	return &v1.ResolverMap{
		Name: name,
		Metadata: &gloov1.Metadata{
			Annotations: map[string]string{
				"foo": "bar",
			},
		},
	}
}