package cc_client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCCClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CCClient Suite")
}
