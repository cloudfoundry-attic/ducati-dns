package resolver_test

import (
	"github.com/cloudfoundry-incubator/ducati-dns/fakes"
	"github.com/cloudfoundry-incubator/ducati-dns/resolver"
	"github.com/miekg/dns"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("Muxer", func() {
	var (
		muxer *resolver.Muxer

		responseWriter *fakes.ResponseWriter
		request        *dns.Msg
		fakeLogger     *lagertest.TestLogger

		suffixPresentHandler *fakes.Handler
		defaultHandler       *fakes.Handler
	)

	BeforeEach(func() {
		request = &dns.Msg{}
		fakeLogger = lagertest.NewTestLogger("test")

		suffixPresentHandler = &fakes.Handler{}
		defaultHandler = &fakes.Handler{}
		muxer = &resolver.Muxer{
			Logger:               fakeLogger,
			Suffix:               "potato",
			SuffixPresentHandler: suffixPresentHandler,
			DefaultHandler:       defaultHandler,
		}
		responseWriter = &fakes.ResponseWriter{}
	})

	It("logs the request", func() {
		request.SetQuestion(dns.Fqdn("something.potato"), dns.TypeA)
		muxer.ServeDNS(responseWriter, request)

		Expect(fakeLogger).To(gbytes.Say("test.serve-dns"))
	})

	Context("when the suffix is present", func() {
		Context("when the request question ends with the suffix", func() {
			BeforeEach(func() {
				request.SetQuestion(dns.Fqdn("something.potato"), dns.TypeA)
			})

			It("forwards the request to suffix present handler", func() {
				muxer.ServeDNS(responseWriter, request)

				Expect(suffixPresentHandler.ServeDNSCallCount()).To(Equal(1))
				w, r := suffixPresentHandler.ServeDNSArgsForCall(0)
				Expect(w).To(Equal(responseWriter))
				Expect(r).To(Equal(request))
			})

			It("does not use the default handler", func() {
				muxer.ServeDNS(responseWriter, request)

				Expect(defaultHandler.ServeDNSCallCount()).To(Equal(0))
			})
		})

		Context("when the request question does not end with the suffix", func() {
			BeforeEach(func() {
				request.SetQuestion(dns.Fqdn("potato.else"), dns.TypeA)
			})

			It("forwards the request to default handler", func() {
				muxer.ServeDNS(responseWriter, request)

				Expect(defaultHandler.ServeDNSCallCount()).To(Equal(1))
				w, r := defaultHandler.ServeDNSArgsForCall(0)
				Expect(w).To(Equal(responseWriter))
				Expect(r).To(Equal(request))
			})

			It("does not use the suffix present handler", func() {
				muxer.ServeDNS(responseWriter, request)

				Expect(suffixPresentHandler.ServeDNSCallCount()).To(Equal(0))
			})
		})
	})

	Context("when the suffix is not set", func() {
		BeforeEach(func() {
			muxer.Suffix = ""
			request.SetQuestion(dns.Fqdn("something.potato"), dns.TypeA)
		})

		It("forwards the request to default handler", func() {
			muxer.ServeDNS(responseWriter, request)

			Expect(defaultHandler.ServeDNSCallCount()).To(Equal(1))
			w, r := defaultHandler.ServeDNSArgsForCall(0)
			Expect(w).To(Equal(responseWriter))
			Expect(r).To(Equal(request))
		})

		It("does not use the suffix present handler", func() {
			muxer.ServeDNS(responseWriter, request)

			Expect(suffixPresentHandler.ServeDNSCallCount()).To(Equal(0))
		})
	})
})
