package resolver_test

import (
	"errors"
	"time"

	"github.com/cloudfoundry-incubator/ducati-dns/fakes"
	"github.com/cloudfoundry-incubator/ducati-dns/resolver"
	"github.com/miekg/dns"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ForwardingResolver", func() {
	var (
		forwardingResolver *resolver.ForwardingResolver
		responseWriter     *fakes.ResponseWriter
		request            *dns.Msg
	)

	BeforeEach(func() {
		request = &dns.Msg{}
		request.SetQuestion(dns.Fqdn("cloudfoundry.org"), dns.TypeA)
		forwardingResolver = &resolver.ForwardingResolver{}
		responseWriter = &fakes.ResponseWriter{}
	})

	Context("when the resolver is not configured with dns servers", func() {
		It("responds with the appropriate error", func() {
			expectedResponse := &dns.Msg{}
			expectedResponse.SetReply(request)
			expectedResponse.SetRcode(request, dns.RcodeNameError)

			forwardingResolver.ServeDNS(responseWriter, request)
			Expect(responseWriter.WriteMsgCallCount()).To(Equal(1))

			response := responseWriter.WriteMsgArgsForCall(0)
			Expect(response).To(Equal(expectedResponse))
		})
	})

	Context("when the resolver is configures with a dns server", func() {
		var fakeExchanger *fakes.Exchanger

		BeforeEach(func() {
			fakeExchanger = &fakes.Exchanger{}
			forwardingResolver = &resolver.ForwardingResolver{
				Servers:   []string{"1.2.3.4:53"},
				Exchanger: fakeExchanger,
			}
		})

		It("forwards the request to the server", func() {
			expectedResponse := &dns.Msg{}
			expectedResponse.SetReply(request)
			fakeExchanger.ExchangeReturns(expectedResponse, 99*time.Second, nil)

			forwardingResolver.ServeDNS(responseWriter, request)

			Expect(fakeExchanger.ExchangeCallCount()).To(Equal(1))

			msg, address := fakeExchanger.ExchangeArgsForCall(0)
			Expect(msg).To(Equal(request))
			Expect(address).To(Equal("1.2.3.4:53"))

			Expect(responseWriter.WriteMsgCallCount()).To(Equal(1))

			response := responseWriter.WriteMsgArgsForCall(0)
			Expect(response).To(Equal(expectedResponse))
		})
	})

	Context("when at least one of several forwarded requests returns a response", func() {
		var fakeExchanger *fakes.Exchanger

		BeforeEach(func() {
			fakeExchanger = &fakes.Exchanger{}
			forwardingResolver = &resolver.ForwardingResolver{
				Servers:   []string{"1.2.3.4:53", "2.3.4.5:53"},
				Exchanger: fakeExchanger,
			}
		})

		It("writes the response", func() {
			expectedResponse := &dns.Msg{}
			expectedResponse.SetReply(request)
			fakeExchanger.ExchangeStub = func(m *dns.Msg, a string) (*dns.Msg, time.Duration, error) {
				if a == "2.3.4.5:53" {
					return expectedResponse, 99 * time.Second, nil
				} else {
					return &dns.Msg{}, 99 * time.Second, errors.New("server fail")
				}
			}

			forwardingResolver.ServeDNS(responseWriter, request)

			Eventually(responseWriter.WriteMsgCallCount()).Should(Equal(1))

			response := responseWriter.WriteMsgArgsForCall(0)
			Eventually(response).Should(Equal(expectedResponse))
		})
	})
})
