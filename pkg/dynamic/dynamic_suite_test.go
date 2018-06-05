package dynamic_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDynamic(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dynamic Suite")
}
