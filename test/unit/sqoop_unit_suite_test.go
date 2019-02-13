package unit_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDynamic(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sqoop unit Suite")
}
