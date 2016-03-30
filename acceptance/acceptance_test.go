package acceptance_test

import (
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("AcceptanceTests", func() {
	var serverSession *gexec.Session

	Context("when only one server is specified", func() {
		BeforeEach(func() {
			var err error

			serverCmd := exec.Command(pathToBinary, "--server", "8.8.8.8:53")
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
			clientSession, err := gexec.Start(clientCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(clientSession).Should(gexec.Exit(0))

			// verify client works
			Expect(clientSession.Out).To(gbytes.Say("ANSWER SECTION:\nwww.example.com."))
			Expect(clientSession.Out).To(gbytes.Say("93.184.216.34"))

			// shut down server
			serverSession.Interrupt()
			Eventually(serverSession).Should(gexec.Exit(0))
			serverSession = nil
		})
	})

	Context("when multiple servers are specified", func() {
		BeforeEach(func() {
			var err error

			serverCmd := exec.Command(pathToBinary, "--server", "1.2.3.4:53", "--server", "8.8.8.8:53")
			serverSession, err = gexec.Start(serverCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			if serverSession != nil {
				serverSession.Interrupt()
				Eventually(serverSession).Should(gexec.Exit())
			}
		})
		It("will try multiple dns servers", func() {
			Consistently(serverSession).ShouldNot(gexec.Exit())

			// run the client
			clientCmd := exec.Command("dig", "@127.0.0.1", "-p", "9999", "www.example.com")
			clientSession, err := gexec.Start(clientCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(clientSession, 4*time.Second, 50*time.Millisecond).Should(gexec.Exit(0))

			// verify client works
			Expect(clientSession.Out).To(gbytes.Say("ANSWER SECTION:\nwww.example.com."))
			Expect(clientSession.Out).To(gbytes.Say("93.184.216.34"))

			// shut down server
			serverSession.Interrupt()
			Eventually(serverSession).Should(gexec.Exit(0))
			serverSession = nil
		})
	})
})
