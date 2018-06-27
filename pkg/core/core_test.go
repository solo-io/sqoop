package core_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	gloov1 "github.com/solo-io/gloo/pkg/api/types/v1"
	glooopts "github.com/solo-io/gloo/pkg/bootstrap"
	"github.com/solo-io/gloo/pkg/bootstrap/configstorage"
	"github.com/solo-io/gloo/pkg/coreplugins/static"
	"github.com/solo-io/gloo/pkg/plugins/rest"
	"github.com/solo-io/qloo/pkg/bootstrap"
	. "github.com/solo-io/qloo/pkg/core"
	"github.com/solo-io/qloo/test"
)

var qlooPort int

var _ = Describe("Core", func() {
	It("does the happy path", func() {
		rand.Seed(time.Now().Unix())
		qlooPort = 9090
		opts := bootstrap.Options{
			Options: glooopts.Options{
				ConfigStorageOptions: glooopts.StorageOptions{
					Type:          "file",
					SyncFrequency: time.Millisecond,
				},
				FileOptions: glooopts.FileOptions{
					ConfigDir: glooInstance.ConfigDir() + "/_gloo_config",
				},
			},
			ProxyAddr:          envoyInstance.LocalAddr() + ":8080",
			BindAddr:           fmt.Sprintf(":%v", qlooPort),
			RoleName:           "qloo-test",
			VirtualServiceName: "qloo-test",
		}
		eventLoop, err := Setup(opts)
		Expect(err).NotTo(HaveOccurred())
		stop := make(chan struct{})
		go eventLoop.Run(stop)

		err = envoyInstance.RunWithId(opts.RoleName + "~e2e-test")
		Expect(err).NotTo(HaveOccurred())

		err = glooInstance.Run()
		Expect(err).NotTo(HaveOccurred())

		gloo, err := configstorage.Bootstrap(opts.Options)
		Expect(err).NotTo(HaveOccurred())

		_, err = gloo.V1().Upstreams().Create(starWarsUpstream())
		Expect(err).NotTo(HaveOccurred())

		qloo, err := bootstrap.Bootstrap(opts.Options)
		Expect(err).NotTo(HaveOccurred())

		_, err = qloo.V1().Schemas().Create(test.StarWarsV1Schema())
		Expect(err).NotTo(HaveOccurred())

		_, err = qloo.V1().ResolverMaps().Create(test.StarWarsResolverMap())
		Expect(err).NotTo(HaveOccurred())

		// it should create the virtual service
		var virtualServices []*gloov1.VirtualService
		Eventually(func() ([]*gloov1.VirtualService, error) {
			virtualServices, err = gloo.V1().VirtualServices().List()
			return virtualServices, err
		}, time.Second*2).Should(HaveLen(1))
		Expect(virtualServices[0].Name).To(Equal(opts.VirtualServiceName))

		// it should create the role
		var roles []*gloov1.Role
		Eventually(func() ([]*gloov1.Role, error) {
			roles, err = gloo.V1().Roles().List()
			return roles, err
		}, time.Second*2).Should(HaveLen(1))
		Expect(roles[0].Name).To(Equal(opts.VirtualServiceName))
		Expect(roles[0].Listeners).To(HaveLen(1))
		Expect(roles[0].Listeners[0].BindPort).To(Equal(8080))
		Expect(roles[0].Listeners[0].VirtualServices).To(HaveLen(1))
		Expect(roles[0].Listeners[0].VirtualServices[0]).To(Equal(opts.VirtualServiceName))

		eventuallyQueryShouldRespond(`{"query": "{hero{name}}"}`,
			`{"data":{"hero":{"name":"R2-D2"}}}`)

		eventuallyQueryShouldRespond(`{"query": "{human(id: 1001){name friends{name}}}"}`,
			`{"data":{"human":{"name":"Darth Vader","friends":[{"name":"Wilhuff Tarkin"}]}}}`)

	})
})

func eventuallyQueryShouldRespond(queryString, expectedString string) {
	Eventually(func() (string, error) {
		res, err := http.Post(fmt.Sprintf("http://localhost:%v/starwars-schema/query", qlooPort),
			"",
			bytes.NewBuffer([]byte(queryString)))
		if err != nil {
			return "", err
		}
		if res.StatusCode != 200 {
			return "", errors.Errorf("bad status code %v", res.StatusCode)
		}
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}, time.Second*45).Should(ContainSubstring(expectedString))
}

func ptr(str string) *string {
	return &str
}

func starWarsUpstream() *gloov1.Upstream {
	return &gloov1.Upstream{
		Name: "starwars-rest",
		Type: static.UpstreamTypeService,
		Spec: static.EncodeUpstreamSpec(static.UpstreamSpec{
			Hosts: []static.Host{
				{
					Addr: "localhost",
					Port: starWarsPort,
				},
			},
		}),
		ServiceInfo: &gloov1.ServiceInfo{
			Type: rest.ServiceTypeREST,
		},
		Functions: []*gloov1.Function{
			{
				Name: "GetHero",
				Spec: rest.EncodeFunctionSpec(rest.Template{
					Header: map[string]string{":method": "GET"},
					Path:   "/api/hero",
				}),
			},
			{
				Name: "GetCharacter",
				Spec: rest.EncodeFunctionSpec(rest.Template{
					Body: ptr(""),
					Header: map[string]string{
						"x-id":    "{{id}}",
						":method": "GET",
					},
					Path: "/api/character",
				}),
			},
			{
				Name: "GetCharacters",
				Spec: rest.EncodeFunctionSpec(rest.Template{
					Header: map[string]string{
						":method": "POST",
					},
					Path: "/api/characters",
				}),
			},
		},
	}
}
