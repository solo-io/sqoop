package schema_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/sqoop/cli/pkg/helpers"
	v1 "github.com/solo-io/sqoop/pkg/api/v1"
	"github.com/solo-io/sqoop/pkg/defaults"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/sqoop/cli/pkg/testutils"
)

const (
	schemaName = "one"

	update = "update"
	create = "create"
)

var _ = Describe("Schema", func() {

	BeforeEach(func() {
		helpers.UseMemoryClients()
	})

	getSchema := func(name string) *v1.Schema {
		schema, err := helpers.MustSchemaClient().Read(defaults.GlooSystem, name, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())
		return schema
	}

	baseCommand := func(cmdType string) {
		schemaFile := testutils.MustWriteTestFile(exampleSchema)
		defer os.Remove(schemaFile)
		err := testutils.Sqoopctl(fmt.Sprintf("schema %s %s -f %s", cmdType, schemaName, schemaFile))
		Expect(err).NotTo(HaveOccurred())
		schemaOne := getSchema(schemaName)
		Expect(schemaOne.InlineSchema).To(Equal(exampleSchema))
		Expect(schemaOne.Metadata.Name).To(Equal(schemaName))
	}

	Context("create", func() {
		It("can create a schema when a file is properly supplied", func() {
			baseCommand("create")
		})

		It("fails when no schema is provided", func() {
			err := testutils.Sqoopctl(fmt.Sprintf("schema create %s", schemaName))
			Expect(err).To(HaveOccurred())
		})

		It("fails when no name is provided", func() {
			schemaFile := testutils.MustWriteTestFile(exampleSchema)
			defer os.Remove(schemaFile)
			err := testutils.Sqoopctl(fmt.Sprintf("schema create -f %s", schemaFile))
			Expect(err).To(HaveOccurred())
		})
	})

	Context("update", func() {
		It("can update a schema when a file is properly supplied, and schema exists", func() {
			baseCommand(create)
			baseCommand(update)
		})

		It("fails when schema doesn't exist is provided", func() {
			err := testutils.Sqoopctl(fmt.Sprintf("schema update %s -f %s", "incorrect_name", "hello"))
			Expect(err).To(HaveOccurred())
		})

		It("fails when no schema is provided", func() {
			err := testutils.Sqoopctl(fmt.Sprintf("schema update %s", schemaName))
			Expect(err).To(HaveOccurred())
		})

		It("fails when no name is provided", func() {
			schemaFile := testutils.MustWriteTestFile(exampleSchema)
			defer os.Remove(schemaFile)
			err := testutils.Sqoopctl(fmt.Sprintf("schema update -f %s", schemaFile))
			Expect(err).To(HaveOccurred())
		})
	})

	Context("delete", func() {
		It("can delete a schema when schema exists", func() {
			baseCommand(create)
			err := testutils.Sqoopctl(fmt.Sprintf("schema delete %s", schemaName))
			Expect(err).NotTo(HaveOccurred())
		})

		It("fails when schema doesn't exist is provided", func() {
			err := testutils.Sqoopctl(fmt.Sprintf("schema delete %s", schemaName))
			Expect(err).To(HaveOccurred())
		})

		It("fails when no name is provided", func() {
			err := testutils.Sqoopctl(fmt.Sprintf("schema delete"))
			Expect(err).To(HaveOccurred())
		})
	})

})

const exampleSchema = `
# The query type, represents all of the entry points into our object graph
type Query {
    pets: [Pet]
    pet(id: Int!): Pet
}

type Mutation {
    addPet(pet: InputPet!): Pet
}

type Pet{
    id: ID!
    name: String!
}

input InputPet{
    id: ID!
    name: String!
    tag: String
}
`
