package acceptance_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("AcceptanceTests", func() {
	var serverSession *gexec.Session

	BeforeEach(func() {
		var err error

		serverCmd := exec.Command(pathToBinary, "some-flag")
		serverSession, err = gexec.Start(serverCmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if serverSession != nil {
			serverSession.Interrupt()
			Eventually(serverSession).Should(gexec.Exit())
		}
	})

	It("exits with status 0", func() {
		Consistently(serverSession).ShouldNot(gexec.Exit())

		// shut down server
		serverSession.Interrupt()
		Eventually(serverSession).Should(gexec.Exit(0))
		serverSession = nil
	})

	It("responds to DNS queries", func() {
		Consistently(serverSession).ShouldNot(gexec.Exit())

		// run the client
		clientCmd := exec.Command("dig", "@127.0.0.1", "-p", "9999", "www.example.com")
		clientOutput, err := clientCmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred())

		// verify client works
		Expect(clientOutput).To(ContainSubstring("93.184.216.34"))
		Expect(clientOutput).To(ContainSubstring(`ANSWER SECTION:
www.example.com.`))

		// shut down server
		serverSession.Interrupt()
		Eventually(serverSession).Should(gexec.Exit(0))
		serverSession = nil
	})

})
