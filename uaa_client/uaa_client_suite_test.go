package uaa_client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestUAAClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UAAClient Suite")
}
