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
	var mockCCAPIServer *httptest.Server
	var mockUAAServer *httptest.Server
	var ccAPIReceivedTokenCount int
	var uaaRequestCount int
	var recordedBasicAuthUser, recordedBasicAuthPassword string
	var happyPathArgs []string

	BeforeEach(func() {
		listenPort = strconv.Itoa(11999 + GinkgoParallelNode())
		mockDucatiAPIServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/containers/my-container-id" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"container_id": "my-container-id", "IP": "10.11.12.13"}`))
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		mockCCAPIServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header["Authorization"][0] == "Bearer my-token" {
				ccAPIReceivedTokenCount++
			}
			switch r.URL.Path {
			case "/v2/organizations":
				w.Write([]byte(`{"resources":[{"metadata":{"guid":"my-org-guid"},"entity":{"name":"my-org"}}]}`))
				return
			case "/v2/spaces":
				w.Write([]byte(`{"resources":[{"metadata":{"guid":"my-space-guid"},"entity":{"name":"my-space", "organization_guid":"my-org-guid"}}]}`))
				return
			case "/v2/apps":
				w.Write([]byte(`{"resources":[{"metadata":{"guid":"my-container-id"},"entity":{"name":"my-app", "space_guid":"my-space-guid"}}]}`))
				return
			}
		}))
		mockUAAServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.Header["Authorization"][0], "Basic ") {
				uaaRequestCount++
				recordedBasicAuthUser, recordedBasicAuthPassword, _ = r.BasicAuth()
			}

			switch r.URL.Path {
			case "/oauth/token":
				w.Write([]byte(`{"access_token":"my-token"}`))
				return
			}
		}))
		happyPathArgs = []string{
			"--listenAddress=127.0.0.1:" + listenPort,
			"--server=8.8.8.8:53",
			"--ducatiSuffix=potato",
			"--ducatiAPI=" + mockDucatiAPIServer.URL,
			"--ccAPI=" + mockCCAPIServer.URL,
			"--uaaBaseURL=" + mockUAAServer.URL,
			"--uaaClientName=some-client",
			"--uaaClientSecret=a-secret",
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
					clientCmd := exec.Command("dig", "@127.0.0.1", "-p", listenPort, "my-app.my-space.my-org.potato")
					clientSession, err := gexec.Start(clientCmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())

					Eventually(clientSession).Should(gexec.Exit(0))

					// verify client works
					Expect(uaaRequestCount).To(Equal(1))
					Expect(ccAPIReceivedTokenCount).To(Equal(3))
					Expect(clientSession.Out).To(gbytes.Say("ANSWER SECTION:\nmy-app.my-space.my-org.potato"))
					Expect(clientSession.Out).To(gbytes.Say("10.11.12.13"))

					Expect(recordedBasicAuthPassword).To(Equal("a-secret"))
					Expect(recordedBasicAuthUser).To(Equal("some-client"))
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
			Entry("uaaClientName", "uaaClientName"),
			Entry("uaaClientSecret", "uaaClientSecret"),
			Entry("uaaBaseURL", "uaaBaseURL"),
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
