package runner_test

import (
	"errors"

	"github.com/cloudfoundry-incubator/ducati-dns/fakes"
	"github.com/cloudfoundry-incubator/ducati-dns/runner"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Runner", func() {
	var (
		r         *runner.Runner
		dnsServer *fakes.DNSServer
		process   ifrit.Process
	)

	BeforeEach(func() {
		dnsServer = &fakes.DNSServer{}
		r = &runner.Runner{
			DNSServer: dnsServer,
		}
	})

	AfterEach(func() {
		ginkgomon.Kill(process)
	})

	It("calls ListenAndServe", func() {
		process = ifrit.Background(r)
		Eventually(process.Ready()).Should(BeClosed())

		Expect(dnsServer.ListenAndServeCallCount()).To(Equal(1))
	})

	It("shuts down when signaled", func() {
		done := make(chan struct{}, 1)
		dnsServer.ListenAndServeStub = func() error {
			<-done
			return nil
		}
		dnsServer.ShutdownStub = func() error {
			close(done)
			return nil
		}

		process = ifrit.Background(r)
		Eventually(process.Ready()).Should(BeClosed())

		ginkgomon.Interrupt(process)
		Eventually(process.Wait()).Should(Receive(BeNil()))

		Expect(dnsServer.ShutdownCallCount()).To(Equal(1))
	})

	Context("when ListenAndServeFails", func() {
		BeforeEach(func() {
			dnsServer.ListenAndServeReturns(errors.New("welp"))
		})

		It("propagates a meaningful error", func() {
			process = ifrit.Background(r)
			Eventually(process.Wait()).Should(Receive(MatchError("listen and serve: welp")))
		})
	})
})
