package consul_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"

	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/hashicorp/consul/api"
	"github.com/solo-io/gloo/test/helpers"
	"github.com/solo-io/sqoop/pkg/api/types/v1"
	"github.com/solo-io/sqoop/pkg/storage"
	. "github.com/solo-io/sqoop/pkg/storage/consul"
)

var _ = Describe("ConsulStorageClient", func() {
	var rootPath string
	var consul *api.Client
	BeforeEach(func() {
		rootPath = helpers.RandString(4)
		c, err := api.NewClient(api.DefaultConfig())
		Expect(err).NotTo(HaveOccurred())
		consul = c
	})
	AfterEach(func() {
		consul.KV().DeleteTree(rootPath, nil)
	})
	Describe("Schemas", func() {
		Describe("create", func() {
			It("creates the schema as a consul key", func() {
				client, err := NewStorage(api.DefaultConfig(), rootPath, time.Second)
				Expect(err).NotTo(HaveOccurred())
				input := &v1.Schema{
					Name:         "myschema",
					InlineSchema: "foo",
					ResolverMap:  "myresolvers",
				}
				schema, err := client.V1().Schemas().Create(input)
				Expect(err).NotTo(HaveOccurred())
				Expect(schema).NotTo(Equal(input))
				p, _, err := consul.KV().Get(rootPath+"/schemas/"+input.Name, nil)
				Expect(err).NotTo(HaveOccurred())
				var unmarshalledSchema v1.Schema
				err = proto.Unmarshal(p.Value, &unmarshalledSchema)
				Expect(err).NotTo(HaveOccurred())
				Expect(&unmarshalledSchema).To(Equal(input))
				resourceVersion := fmt.Sprintf("%v", p.CreateIndex)
				Expect(schema.Metadata.ResourceVersion).To(Equal(resourceVersion))
				input.Metadata = schema.Metadata
				Expect(schema).To(Equal(input))
			})
			It("errors when creating the same schema twice", func() {
				client, err := NewStorage(api.DefaultConfig(), rootPath, time.Second)
				Expect(err).NotTo(HaveOccurred())
				input := &v1.Schema{
					Name:         "myschema",
					InlineSchema: "foo",
					ResolverMap:  "myresolvers",
				}
				_, err = client.V1().Schemas().Create(input)
				Expect(err).NotTo(HaveOccurred())
				_, err = client.V1().Schemas().Create(input)
				Expect(err).To(HaveOccurred())
			})
			Describe("update", func() {
				It("fails if the schema doesn't exist", func() {
					client, err := NewStorage(api.DefaultConfig(), rootPath, time.Second)
					Expect(err).NotTo(HaveOccurred())
					input := &v1.Schema{
						Name:         "myschema",
						InlineSchema: "foo",
						ResolverMap:  "myresolvers",
					}
					schema, err := client.V1().Schemas().Update(input)
					Expect(err).To(HaveOccurred())
					Expect(schema).To(BeNil())
				})
				It("fails if the resourceversion is not up to date", func() {
					client, err := NewStorage(api.DefaultConfig(), rootPath, time.Second)
					Expect(err).NotTo(HaveOccurred())
					input := &v1.Schema{
						Name:         "myschema",
						InlineSchema: "foo",
						ResolverMap:  "myresolvers",
					}
					_, err = client.V1().Schemas().Create(input)
					Expect(err).NotTo(HaveOccurred())
					v, err := client.V1().Schemas().Update(input)
					Expect(v).To(BeNil())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("resource version"))
				})
				It("updates the schema", func() {
					client, err := NewStorage(api.DefaultConfig(), rootPath, time.Second)
					Expect(err).NotTo(HaveOccurred())
					input := &v1.Schema{
						Name:         "myschema",
						InlineSchema: "foo",
						ResolverMap:  "myresolvers",
					}
					schema, err := client.V1().Schemas().Create(input)
					Expect(err).NotTo(HaveOccurred())
					changed := proto.Clone(input).(*v1.Schema)
					changed.InlineSchema = "bar"
					// match resource version
					changed.Metadata = schema.Metadata
					out, err := client.V1().Schemas().Update(changed)
					Expect(err).NotTo(HaveOccurred())
					Expect(out.InlineSchema).To(Equal(changed.InlineSchema))
				})
				Describe("get", func() {
					It("fails if the schema doesn't exist", func() {
						client, err := NewStorage(api.DefaultConfig(), rootPath, time.Second)
						Expect(err).NotTo(HaveOccurred())
						schema, err := client.V1().Schemas().Get("foo")
						Expect(err).To(HaveOccurred())
						Expect(schema).To(BeNil())
					})
					It("returns the schema", func() {
						client, err := NewStorage(api.DefaultConfig(), rootPath, time.Second)
						Expect(err).NotTo(HaveOccurred())
						input := &v1.Schema{
							Name:         "myschema",
							InlineSchema: "foo",
							ResolverMap:  "myresolvers",
						}
						schema, err := client.V1().Schemas().Create(input)
						Expect(err).NotTo(HaveOccurred())
						out, err := client.V1().Schemas().Get(input.Name)
						Expect(err).NotTo(HaveOccurred())
						Expect(out).To(Equal(schema))
						input.Metadata = out.Metadata
						Expect(out).To(Equal(input))
					})
				})
				Describe("list", func() {
					It("returns all existing schemas", func() {
						client, err := NewStorage(api.DefaultConfig(), rootPath, time.Second)
						Expect(err).NotTo(HaveOccurred())
						input1 := &v1.Schema{
							Name: "myschema1",
						}
						input2 := &v1.Schema{
							Name: "myschema2",
						}
						input3 := &v1.Schema{
							Name: "myschema3",
						}
						schema1, err := client.V1().Schemas().Create(input1)
						Expect(err).NotTo(HaveOccurred())
						schema2, err := client.V1().Schemas().Create(input2)
						Expect(err).NotTo(HaveOccurred())
						schema3, err := client.V1().Schemas().Create(input3)
						Expect(err).NotTo(HaveOccurred())
						out, err := client.V1().Schemas().List()
						Expect(err).NotTo(HaveOccurred())
						Expect(out).To(ContainElement(schema1))
						Expect(out).To(ContainElement(schema2))
						Expect(out).To(ContainElement(schema3))
					})
				})
				Describe("watch", func() {
					It("watches", func() {
						client, err := NewStorage(api.DefaultConfig(), rootPath, time.Second)
						Expect(err).NotTo(HaveOccurred())
						lists := make(chan []*v1.Schema, 3)
						stop := make(chan struct{})
						defer close(stop)
						errs := make(chan error)
						w, err := client.V1().Schemas().Watch(&storage.SchemaEventHandlerFuncs{
							UpdateFunc: func(updatedList []*v1.Schema, _ *v1.Schema) {
								lists <- updatedList
							},
						})
						Expect(err).NotTo(HaveOccurred())
						go func() {
							w.Run(stop, errs)
						}()
						input1 := &v1.Schema{
							Name: "myschema1",
						}
						input2 := &v1.Schema{
							Name: "myschema2",
						}
						input3 := &v1.Schema{
							Name: "myschema3",
						}
						schema1, err := client.V1().Schemas().Create(input1)
						Expect(err).NotTo(HaveOccurred())
						schema2, err := client.V1().Schemas().Create(input2)
						Expect(err).NotTo(HaveOccurred())
						schema3, err := client.V1().Schemas().Create(input3)
						Expect(err).NotTo(HaveOccurred())

						var list []*v1.Schema
						Eventually(func() []*v1.Schema {
							select {
							default:
								return nil
							case l := <-lists:
								list = l
								return l
							}
						}).Should(HaveLen(3))
						Expect(list).To(HaveLen(3))
						Expect(list).To(ContainElement(schema1))
						Expect(list).To(ContainElement(schema2))
						Expect(list).To(ContainElement(schema3))
					})
				})
			})
		})
	})
	Describe("ResolverMaps", func() {
		Describe("create", func() {
			It("creates the resolverMap as a consul key", func() {
				client, err := NewStorage(api.DefaultConfig(), rootPath, time.Second)
				Expect(err).NotTo(HaveOccurred())
				input := &v1.ResolverMap{
					Name: "myresolverMap",
				}
				resolverMap, err := client.V1().ResolverMaps().Create(input)
				Expect(err).NotTo(HaveOccurred())
				Expect(resolverMap).NotTo(Equal(input))
				p, _, err := consul.KV().Get(rootPath+"/resolverMaps/"+input.Name, nil)
				Expect(err).NotTo(HaveOccurred())
				var unmarshalledResolverMap v1.ResolverMap
				err = proto.Unmarshal(p.Value, &unmarshalledResolverMap)
				Expect(err).NotTo(HaveOccurred())
				Expect(&unmarshalledResolverMap).To(Equal(input))
				resourceVersion := fmt.Sprintf("%v", p.CreateIndex)
				Expect(resolverMap.Metadata.ResourceVersion).To(Equal(resourceVersion))
				input.Metadata = resolverMap.Metadata
				Expect(resolverMap).To(Equal(input))
			})
			It("errors when creating the same resolverMap twice", func() {
				client, err := NewStorage(api.DefaultConfig(), rootPath, time.Second)
				Expect(err).NotTo(HaveOccurred())
				input := &v1.ResolverMap{
					Name: "myresolverMap",
				}
				_, err = client.V1().ResolverMaps().Create(input)
				Expect(err).NotTo(HaveOccurred())
				_, err = client.V1().ResolverMaps().Create(input)
				Expect(err).To(HaveOccurred())
			})
			Describe("update", func() {
				It("fails if the resolverMap doesn't exist", func() {
					client, err := NewStorage(api.DefaultConfig(), rootPath, time.Second)
					Expect(err).NotTo(HaveOccurred())
					input := &v1.ResolverMap{
						Name: "myresolverMap",
					}
					resolverMap, err := client.V1().ResolverMaps().Update(input)
					Expect(err).To(HaveOccurred())
					Expect(resolverMap).To(BeNil())
				})
				It("fails if the resourceversion is not up to date", func() {
					client, err := NewStorage(api.DefaultConfig(), rootPath, time.Second)
					Expect(err).NotTo(HaveOccurred())
					input := &v1.ResolverMap{
						Name: "myresolverMap",
					}
					_, err = client.V1().ResolverMaps().Create(input)
					Expect(err).NotTo(HaveOccurred())
					v, err := client.V1().ResolverMaps().Update(input)
					Expect(v).To(BeNil())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("resource version"))
				})
				Describe("get", func() {
					It("fails if the resolverMap doesn't exist", func() {
						client, err := NewStorage(api.DefaultConfig(), rootPath, time.Second)
						Expect(err).NotTo(HaveOccurred())
						resolverMap, err := client.V1().ResolverMaps().Get("foo")
						Expect(err).To(HaveOccurred())
						Expect(resolverMap).To(BeNil())
					})
					It("returns the resolverMap", func() {
						client, err := NewStorage(api.DefaultConfig(), rootPath, time.Second)
						Expect(err).NotTo(HaveOccurred())
						input := &v1.ResolverMap{
							Name: "myresolverMap",
						}
						resolverMap, err := client.V1().ResolverMaps().Create(input)
						Expect(err).NotTo(HaveOccurred())
						out, err := client.V1().ResolverMaps().Get(input.Name)
						Expect(err).NotTo(HaveOccurred())
						Expect(out).To(Equal(resolverMap))
						input.Metadata = out.Metadata
						Expect(out).To(Equal(input))
					})
				})
				Describe("list", func() {
					It("returns all existing resolverMaps", func() {
						client, err := NewStorage(api.DefaultConfig(), rootPath, time.Second)
						Expect(err).NotTo(HaveOccurred())
						input1 := &v1.ResolverMap{
							Name: "myresolverMap1",
						}
						input2 := &v1.ResolverMap{
							Name: "myresolverMap2",
						}
						input3 := &v1.ResolverMap{
							Name: "myresolverMap3",
						}
						resolverMap1, err := client.V1().ResolverMaps().Create(input1)
						Expect(err).NotTo(HaveOccurred())
						time.Sleep(time.Second)
						resolverMap2, err := client.V1().ResolverMaps().Create(input2)
						Expect(err).NotTo(HaveOccurred())
						time.Sleep(time.Second)
						resolverMap3, err := client.V1().ResolverMaps().Create(input3)
						Expect(err).NotTo(HaveOccurred())
						out, err := client.V1().ResolverMaps().List()
						Expect(err).NotTo(HaveOccurred())
						Expect(out).To(ContainElement(resolverMap1))
						Expect(out).To(ContainElement(resolverMap2))
						Expect(out).To(ContainElement(resolverMap3))
					})
				})
				Describe("watch", func() {
					It("watches", func() {
						client, err := NewStorage(api.DefaultConfig(), rootPath, time.Second)
						Expect(err).NotTo(HaveOccurred())
						lists := make(chan []*v1.ResolverMap, 3)
						stop := make(chan struct{})
						defer close(stop)
						errs := make(chan error)
						w, err := client.V1().ResolverMaps().Watch(&storage.ResolverMapEventHandlerFuncs{
							UpdateFunc: func(updatedList []*v1.ResolverMap, _ *v1.ResolverMap) {
								lists <- updatedList
							},
						})
						Expect(err).NotTo(HaveOccurred())
						go func() {
							w.Run(stop, errs)
						}()
						input1 := &v1.ResolverMap{
							Name: "myresolverMap1",
						}
						input2 := &v1.ResolverMap{
							Name: "myresolverMap2",
						}
						input3 := &v1.ResolverMap{
							Name: "myresolverMap3",
						}
						resolverMap1, err := client.V1().ResolverMaps().Create(input1)
						Expect(err).NotTo(HaveOccurred())
						resolverMap2, err := client.V1().ResolverMaps().Create(input2)
						Expect(err).NotTo(HaveOccurred())
						resolverMap3, err := client.V1().ResolverMaps().Create(input3)
						Expect(err).NotTo(HaveOccurred())

						var list []*v1.ResolverMap
						Eventually(func() []*v1.ResolverMap {
							select {
							default:
								return nil
							case l := <-lists:
								list = l
								return l
							}
						}).Should(HaveLen(3))
						Expect(list).To(HaveLen(3))
						Expect(list).To(ContainElement(resolverMap1))
						Expect(list).To(ContainElement(resolverMap2))
						Expect(list).To(ContainElement(resolverMap3))
					})
				})
			})
		})
	})
})
