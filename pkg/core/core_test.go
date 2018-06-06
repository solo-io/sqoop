package core_test

import (
	. "github.com/onsi/ginkgo"
	"github.com/solo-io/qloo/pkg/bootstrap"
	glooopts "github.com/solo-io/gloo/pkg/bootstrap"
	"time"
	"github.com/solo-io/gloo/pkg/bootstrap/configstorage"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/qloo/pkg/core"
	"math/rand"
	"fmt"
	"github.com/solo-io/qloo/test"
	gloov1 "github.com/solo-io/gloo/pkg/api/types/v1"
	"github.com/solo-io/gloo/pkg/coreplugins/service"
	"github.com/solo-io/gloo/pkg/plugins/rest"
)

var _ = Describe("Core", func() {
	It("does the happy path", func() {
		rand.Seed(time.Now().Unix())
		port := rand.Int31n(10) + 10000
		opts := bootstrap.Options{
			Options: glooopts.Options{
				ConfigStorageOptions: glooopts.StorageOptions{
					Type:          "file",
					SyncFrequency: time.Millisecond,
				},
				FileOptions: glooopts.FileOptions{
					ConfigDir: glooInstance.ConfigDir(),
				},
			},
			ProxyAddr:          envoyInstance.LocalAddr(),
			BindAddr:           fmt.Sprintf(":%v", port),
			RoleName:           "qloo-test",
			VirtualServiceName: "qloo-test",
		}
		eventLoop, err := Setup(opts)
		Expect(err).NotTo(HaveOccurred())
		stop := make(chan struct{})
		go eventLoop.Run(stop)

		gloo, err := configstorage.Bootstrap(opts.Options)
		Expect(err).NotTo(HaveOccurred())

		_, err = gloo.V1().Upstreams().Create(starWarsUpstream)
		Expect(err).NotTo(HaveOccurred())

		qloo, err := bootstrap.Bootstrap(opts.Options)
		Expect(err).NotTo(HaveOccurred())

		_, err = qloo.V1().Schemas().Create(test.StarWarsV1Schema())
		Expect(err).NotTo(HaveOccurred())

		_, err = qloo.V1().ResolverMaps().Create(test.StarWarsResolverMap())
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() ([]*gloov1.VirtualService, error) {
			return gloo.V1().VirtualServices().List()
		}).Should(Equal("foo"))
	})
})

func ptr(str string) *string {
	return &str
}

var starWarsUpstream = &gloov1.Upstream{
	Name: "starwars-rest-test",
	Type: service.UpstreamTypeService,
	Spec: service.EncodeUpstreamSpec(service.UpstreamSpec{
		Hosts: []service.Host{
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
				Path: "/api/hero",
			}),
		},
		{
			Name: "GetCharacter",
			Spec: rest.EncodeFunctionSpec(rest.Template{
				Body: ptr(""),
				Header: map[string]string{
					"x-id": "{{id}}",
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
