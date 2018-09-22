package configwatcher

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/solo-io/gloo/pkg/log"
	. "github.com/solo-io/gloo/test/helpers"
	"github.com/solo-io/sqoop/pkg/storage/file"
	"github.com/solo-io/sqoop/test"
)

var _ = Describe("FileConfigWatcher", func() {
	var (
		dir string
		err error
	)
	BeforeEach(func() {
		dir, err = ioutil.TempDir("", "filecachetest")
		Must(err)
	})
	AfterEach(func() {
		log.Debugf("removing " + dir)
		os.RemoveAll(dir)
	})
	Describe("controller", func() {
		It("watches gloo files", func() {
			storageClient, err := file.NewStorage(dir, time.Millisecond)
			Must(err)
			watcher, err := NewConfigWatcher(storageClient)
			Must(err)
			go func() { watcher.Run(make(chan struct{})) }()

			time.Sleep(time.Second)

			schema := test.StarWarsV1Schema()

			created, err := storageClient.V1().Schemas().Create(schema)
			Expect(err).NotTo(HaveOccurred())

			select {
			case <-time.After(time.Second * 5):
				Expect(fmt.Errorf("expected to have received resource event before 5s")).NotTo(HaveOccurred())
			case cfg := <-watcher.Config():
				Expect(len(cfg.Schemas)).To(Equal(1))
				Expect(cfg.Schemas[0].InlineSchema).To(Equal(schema.InlineSchema))
			case err := <-watcher.Error():
				Expect(err).NotTo(HaveOccurred())
			}

			// update with no op, should not give us a config
			_, err = storageClient.V1().Schemas().Update(created)
			Expect(err).NotTo(HaveOccurred())

			select {
			case <-time.After(time.Millisecond * 50):
			case c := <-watcher.Config():
				Expect(c).To(BeNil())
				Expect(fmt.Errorf("should not have recieved duplicate config")).NotTo(HaveOccurred())
			case err := <-watcher.Error():
				Expect(err).NotTo(HaveOccurred())
			}
		})
	})
})
