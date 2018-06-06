package gloo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/qloo/pkg/gloo"
	"github.com/solo-io/gloo/pkg/storage/file"
	"io/ioutil"
	"os"
	"time"
	"github.com/solo-io/gloo/pkg/storage"
)

var _ = Describe("GlooOperator", func() {
	var (
		tmpDir string
		gloo storage.Interface
		vServiceName, roleName = "test-virtualservice", "test-role"
		operator *GlooOperator
	)
	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		gloo, err = file.NewStorage(tmpDir, time.Millisecond)
		Expect(err).NotTo(HaveOccurred())
		operator = NewGlooOperator(gloo, vServiceName, roleName)
	})
	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})
	It("creates the virtualservice with all the required routes", func() {
		
	})
})
