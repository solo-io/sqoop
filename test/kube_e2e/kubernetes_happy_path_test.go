package kube_e2e

import (
	"time"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("Happy Path running on Kubernetes", func() {
	Context("creating kube upstream and a vService with a single route to it", func() {
		It("should configure envoy with a 200 OK route (backed by helloservice)", func() {
			curlEventuallyShouldRespond(curlOpts{
				port: 9090,
				method: "POST",
				path: "/starwars/query",
				body: `{"query": "{hero{name}}"}`,
			}, `{"data":{"hero":{"name":"R2-D2"}}}`, time.Second*50)
		})
	})
})
