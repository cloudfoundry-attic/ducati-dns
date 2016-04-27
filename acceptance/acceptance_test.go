package acceptance_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("AcceptanceTests", func() {
	var serverSession *gexec.Session
	var listenPort string
	var mockDucatiAPIServer *httptest.Server
	var happyPathArgs []string

	BeforeEach(func() {
		listenPort = strconv.Itoa(11999 + GinkgoParallelNode())
		mockDucatiAPIServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/containers/my-app-guid" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"container_id": "my-app-guid", "IP": "10.11.12.13"}`))
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		happyPathArgs = []string{
			"--listenAddress=127.0.0.1:" + listenPort,
			"--server=8.8.8.8:53",
			"--ducatiSuffix=potato",
			"--ducatiAPI=" + mockDucatiAPIServer.URL,
		}
	})

	Context("when only one server is specified", func() {
		BeforeEach(func() {
			var err error

			serverCmd := exec.Command(pathToBinary, happyPathArgs...)
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
			clientCmd := exec.Command("dig", "@127.0.0.1", "-p", listenPort, "www.example.com")
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

		Context("when a ducati api server is specified", func() {
			Context("when the query ends with the configured suffix", func() {
				It("it will resolve requests using the api server", func() {
					Consistently(serverSession).ShouldNot(gexec.Exit())

					// run the client
					clientCmd := exec.Command("dig", "@127.0.0.1", "-p", listenPort, "my-app-guid.potato")
					clientSession, err := gexec.Start(clientCmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())

					Eventually(clientSession).Should(gexec.Exit(0))

					// verify client works
					Expect(clientSession.Out).To(gbytes.Say("ANSWER SECTION:\nmy-app-guid.potato"))
					Expect(clientSession.Out).To(gbytes.Say("10.11.12.13"))
				})
			})
		})
	})

	var removeArrayElement = func(src []string, elementToRemove string) []string {
		reduced := []string{}
		for _, element := range src {
			if !strings.HasPrefix(element, elementToRemove) {
				reduced = append(reduced, element)
			}
		}

		return reduced
	}

	Describe("missing flags", func() {
		DescribeTable("when flags are ommitted", func(missing string) {
			var err error

			serverCmd := exec.Command(pathToBinary, happyPathArgs...)
			serverCmd.Args = removeArrayElement(serverCmd.Args, "--"+missing)

			serverSession, err = gexec.Start(serverCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(serverSession).Should(gexec.Exit(1))
			Expect(serverSession.Err.Contents()).To(ContainSubstring(fmt.Sprintf("missing required arg: %s", missing)))

			if serverSession != nil {
				serverSession.Interrupt()
				Eventually(serverSession).Should(gexec.Exit())
			}
		},
			Entry("server", "server"),
			Entry("ducatiSuffix", "ducatiSuffix"),
			Entry("ducatiAPI", "ducatiAPI"),
		)

	})

	Context("when the server for forwarding is unavailable", func() {
		BeforeEach(func() {
			var err error

			serverCmd := exec.Command(pathToBinary, happyPathArgs...)
			serverCmd.Args = removeArrayElement(serverCmd.Args, "--server")
			serverCmd.Args = append(serverCmd.Args, "--server=127.0.0.1:98765")
			serverSession, err = gexec.Start(serverCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			if serverSession != nil {
				serverSession.Interrupt()
				Eventually(serverSession).Should(gexec.Exit())
			}
		})

		It("logs the resulting errors but keeps running", func() {
			Consistently(serverSession).ShouldNot(gexec.Exit())

			// run the client
			clientCmd := exec.Command("dig", "@127.0.0.1", "-p", listenPort, "www.example.com")
			clientSession, err := gexec.Start(clientCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(clientSession).Should(gexec.Exit(0)) // dig is ok with a SERVFAIL

			// verify client works
			Expect(clientSession.Out).To(gbytes.Say("SERVFAIL"))

			// server is still up
			Consistently(serverSession).ShouldNot(gexec.Exit())
			Expect(serverSession.Out.Contents()).To(ContainSubstring("Serve DNS Exchange"))
		})
	})
})
