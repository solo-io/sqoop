package file_test

import (
	"io/ioutil"
	"os"

	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo/pkg/log"
	. "github.com/solo-io/gloo/pkg/storage/file"
	. "github.com/solo-io/gloo/test/helpers"
)

var _ = Describe("CrdStorageClient", func() {
	var (
		dir    string
		err    error
		resync = time.Second
	)
	BeforeEach(func() {
		dir, err = ioutil.TempDir("", "filecachetest")
		Must(err)
	})
	AfterEach(func() {
		log.Debugf("removing " + dir)
		os.RemoveAll(dir)
	})
	Describe("New", func() {
		It("creates a new client without error", func() {
			_, err = NewStorage(dir, resync)
			Expect(err).NotTo(HaveOccurred())
		})
	})
	Describe("Create", func() {
		It("creates a file from the item", func() {
			client, err := NewStorage(dir, resync)
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
	Describe("Create2Update", func() {
		It("creates and updates", func() {
			client, err := NewStorage(dir, resync)
			Expect(err).NotTo(HaveOccurred())
			err = client.V1().Register()
			Expect(err).NotTo(HaveOccurred())
			schema := NewTestSchema1()
			schema, err = client.V1().Schemas().Create(schema)
			schema2 := NewTestSchema2()
			schema2, err = client.V1().Schemas().Create(schema2)
			Expect(err).NotTo(HaveOccurred())

			_, err = client.V1().Schemas().Update(schema2)
			Expect(err).NotTo(HaveOccurred())

			created1, err := client.V1().Schemas().Get(schema.Name)
			Expect(err).NotTo(HaveOccurred())
			schema.Metadata = created1.Metadata
			Expect(created1).To(Equal(schema))

			created2, err := client.V1().Schemas().Get(schema2.Name)
			Expect(err).NotTo(HaveOccurred())
			schema2.Metadata = created2.Metadata
			Expect(created2).To(Equal(schema2))

		})
	})
	Describe("Create2Update resolverMap", func() {
		It("creates and updates", func() {
			client, err := NewStorage(dir, resync)
			Expect(err).NotTo(HaveOccurred())
			err = client.V1().Register()
			Expect(err).NotTo(HaveOccurred())
			resolverMap := NewTestResolverMap("v1")
			resolverMap, err = client.V1().ResolverMaps().Create(resolverMap)
			resolverMap2 := NewTestResolverMap("v2")
			resolverMap2, err = client.V1().ResolverMaps().Create(resolverMap2)
			Expect(err).NotTo(HaveOccurred())

			_, err = client.V1().ResolverMaps().Update(resolverMap)
			Expect(err).NotTo(HaveOccurred())

			created1, err := client.V1().ResolverMaps().Get(resolverMap.Name)
			Expect(err).NotTo(HaveOccurred())
			resolverMap.Metadata = created1.Metadata
			Expect(created1).To(Equal(resolverMap))

			created2, err := client.V1().ResolverMaps().Get(resolverMap2.Name)
			Expect(err).NotTo(HaveOccurred())
			resolverMap2.Metadata = created2.Metadata
			Expect(created2).To(Equal(resolverMap2))
		})
	})
	Describe("Create2Update role", func() {
		It("creates and updates", func() {
			client, err := NewStorage(dir, resync)
			Expect(err).NotTo(HaveOccurred())
			err = client.V1().Register()
			Expect(err).NotTo(HaveOccurred())
			role := NewTestRole("v1")
			role, err = client.V1().Roles().Create(role)
			role2 := NewTestRole("v2")
			role2, err = client.V1().Roles().Create(role2)
			Expect(err).NotTo(HaveOccurred())

			_, err = client.V1().Roles().Update(role)
			Expect(err).NotTo(HaveOccurred())

			created1, err := client.V1().Roles().Get(role.Name)
			Expect(err).NotTo(HaveOccurred())
			role.Metadata = created1.Metadata
			Expect(created1).To(Equal(role))

			created2, err := client.V1().Roles().Get(role2.Name)
			Expect(err).NotTo(HaveOccurred())
			role2.Metadata = created2.Metadata
			Expect(created2).To(Equal(role2))
		})
	})
	Describe("Get", func() {
		It("gets a file from the name", func() {
			client, err := NewStorage(dir, resync)
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
		It("updates a file from the item", func() {
			client, err := NewStorage(dir, resync)
			Expect(err).NotTo(HaveOccurred())
			err = client.V1().Register()
			Expect(err).NotTo(HaveOccurred())
			schema := NewTestSchema1()
			created, err := client.V1().Schemas().Create(schema)
			Expect(err).NotTo(HaveOccurred())
			schema.Type = "something-else"
			_, err = client.V1().Schemas().Update(schema)
			// need to set resource ver
			Expect(err).To(HaveOccurred())
			schema.Metadata = created.GetMetadata()
			updated, err := client.V1().Schemas().Update(schema)
			Expect(err).NotTo(HaveOccurred())
			schema.Metadata = updated.GetMetadata()
			Expect(updated).To(Equal(schema))
		})
	})
	Describe("Delete", func() {
		It("deletes a file from the name", func() {
			client, err := NewStorage(dir, resync)
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
