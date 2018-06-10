package core_test

import (
	"testing"

	"github.com/solo-io/gloo/test/helpers/local"

	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo/pkg/log"
	"github.com/solo-io/gloo/test/helpers"
	"github.com/solo-io/qloo/examples/starwars/server"
)

var (
	envoyFactory *localhelpers.EnvoyFactory
	glooFactory  *localhelpers.GlooFactory
	starWarsRest *http.Server
	starWarsPort uint32
)

func TestLocalE2e(t *testing.T) {
	helpers.RegisterCommonFailHandlers()
	log.DefaultOut = GinkgoWriter
	RunSpecs(t, "LocalE2e Suite")
}

var _ = BeforeSuite(func() {
	var err error
	envoyFactory, err = localhelpers.NewEnvoyFactory()
	Expect(err).NotTo(HaveOccurred())
	glooFactory, err = localhelpers.NewGlooFactory()
	Expect(err).NotTo(HaveOccurred())
	starWarsPort = 1234
	starWarsRest = &http.Server{
		Addr:    fmt.Sprintf("localhost:%v", starWarsPort),
		Handler: server.New(),
	}
})

var _ = AfterSuite(func() {
	envoyFactory.Clean()
	glooFactory.Clean()
})

var (
	envoyInstance *localhelpers.EnvoyInstance
	glooInstance  *localhelpers.GlooInstance
)

var _ = BeforeEach(func() {
	var err error
	envoyInstance, err = envoyFactory.NewEnvoyInstance()
	Expect(err).NotTo(HaveOccurred())
	glooInstance, err = glooFactory.NewGlooInstance()
	Expect(err).NotTo(HaveOccurred())
	go func() {
		err := starWarsRest.ListenAndServe()
		if err != nil {
			log.Printf("starwars server error: %v", err.Error())
		}
	}()
})

var _ = AfterEach(func() {
	if envoyInstance != nil {
		envoyInstance.Clean()
	}
	if glooInstance != nil {
		glooInstance.Clean()
	}
	starWarsRest.Close()
})
