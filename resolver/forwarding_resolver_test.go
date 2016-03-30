package resolver_test

import (
	"errors"
	"time"

	"github.com/cloudfoundry-incubator/ducati-dns/fakes"
	"github.com/cloudfoundry-incubator/ducati-dns/resolver"
	"github.com/miekg/dns"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("ForwardingResolver", func() {
	var (
		forwardingResolver *resolver.ForwardingResolver
		responseWriter     *fakes.ResponseWriter
		request            *dns.Msg
		fakeExchanger      *fakes.Exchanger
		fakeLogger         *lagertest.TestLogger
	)

	BeforeEach(func() {
		request = &dns.Msg{}
		request.SetQuestion(dns.Fqdn("cloudfoundry.org"), dns.TypeA)
		fakeExchanger = &fakes.Exchanger{}
		fakeExchanger.ExchangeStub = func(request *dns.Msg, server string) (*dns.Msg, time.Duration, error) {
			resp := &dns.Msg{}
			resp.SetReply(request)
			return resp, 99 * time.Second, nil
		}
		fakeLogger = lagertest.NewTestLogger("test")
		forwardingResolver = &resolver.ForwardingResolver{
			Server:    "1.2.3.4:53",
			Exchanger: fakeExchanger,
			Logger:    fakeLogger,
		}
		responseWriter = &fakes.ResponseWriter{}
	})

	It("forwards the request to the server", func() {
		forwardingResolver.ServeDNS(responseWriter, request)

		Expect(fakeExchanger.ExchangeCallCount()).To(Equal(1))

		msg, address := fakeExchanger.ExchangeArgsForCall(0)
		Expect(msg).To(Equal(request))
		Expect(address).To(Equal("1.2.3.4:53"))

		Expect(responseWriter.WriteMsgCallCount()).To(Equal(1))

		expectedResp := &dns.Msg{}
		expectedResp.SetReply(request)
		response := responseWriter.WriteMsgArgsForCall(0)
		Expect(response).To(Equal(expectedResp))
	})

	Context("when the exchanger returns an error", func() {
		BeforeEach(func() {
			fakeExchanger.ExchangeReturns(nil, 0, errors.New("potato"))
		})
		It("logs the error", func() {
			forwardingResolver.ServeDNS(responseWriter, request)
			Expect(fakeLogger).To(gbytes.Say("test.*potato"))
		})
		It("responds with SERVFAIL", func() {
			forwardingResolver.ServeDNS(responseWriter, request)
			Expect(responseWriter.WriteMsgCallCount()).To(Equal(1))
			Expect(responseWriter.WriteMsgArgsForCall(0).MsgHdr.Rcode).To(Equal(dns.RcodeServerFailure))
		})
	})
})
