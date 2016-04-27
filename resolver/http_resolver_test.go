package resolver_test

import (
	"errors"
	"math/rand"
	"net"

	"github.com/cloudfoundry-incubator/ducati-daemon/models"
	"github.com/cloudfoundry-incubator/ducati-dns/fakes"
	"github.com/cloudfoundry-incubator/ducati-dns/resolver"
	"github.com/miekg/dns"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("HTTPResolver", func() {
	var (
		httpResolver     *resolver.HTTPResolver
		responseWriter   *fakes.ResponseWriter
		request          *dns.Msg
		fakeLogger       *lagertest.TestLogger
		fakeDaemonClient *fakes.DucatiDaemonClient
	)

	BeforeEach(func() {
		request = &dns.Msg{
			MsgHdr: dns.MsgHdr{
				Id: uint16(rand.Int()),
			},
		}
		request.SetQuestion(dns.Fqdn("some-app-guid.potato"), dns.TypeA)
		fakeLogger = lagertest.NewTestLogger("test")
		fakeDaemonClient = &fakes.DucatiDaemonClient{}
		fakeDaemonClient.ListContainersReturns([]models.Container{
			{IP: "10.11.12.13", App: "some-app-guid"},
		}, nil)
		httpResolver = &resolver.HTTPResolver{
			Suffix:       "potato",
			DaemonClient: fakeDaemonClient,
			TTL:          42,
			Logger:       fakeLogger,
		}
		responseWriter = &fakes.ResponseWriter{}
	})

	It("resolves DNS queries by using the ducati daemon client", func() {
		httpResolver.ServeDNS(responseWriter, request)

		Expect(fakeDaemonClient.ListContainersCallCount()).To(Equal(1))

		Expect(responseWriter.WriteMsgCallCount()).To(Equal(1))

		expectedResp := &dns.Msg{}
		expectedResp.SetReply(request)
		rr_header := dns.RR_Header{
			Name:   dns.Fqdn("some-app-guid.potato"),
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    42,
		}
		a := &dns.A{rr_header, net.ParseIP("10.11.12.13")}
		expectedResp.Answer = []dns.RR{a}
		Expect(responseWriter.WriteMsgArgsForCall(0)).To(Equal(expectedResp))
	})

	Context("when the requestedName does not end in the suffix", func() {
		BeforeEach(func() {
			request.SetQuestion(dns.Fqdn("something.else.entirely"), dns.TypeA)
		})

		It("should reply with NXDOMAIN", func() {
			httpResolver.ServeDNS(responseWriter, request)

			Expect(responseWriter.WriteMsgCallCount()).To(Equal(1))
			Expect(responseWriter.WriteMsgArgsForCall(0).Id).To(Equal(request.Id))
			Expect(responseWriter.WriteMsgArgsForCall(0).Rcode).To(Equal(dns.RcodeNameError))
		})

		It("should log the error", func() {
			httpResolver.ServeDNS(responseWriter, request)

			Expect(fakeLogger).To(gbytes.Say("unknown-name.*something.else.entirely."))
		})
	})

	Context("when there are no containers at all", func() {
		BeforeEach(func() {
			fakeDaemonClient.ListContainersReturns([]models.Container{}, nil)
		})

		It("should reply with NXDOMAIN", func() {
			httpResolver.ServeDNS(responseWriter, request)

			Expect(responseWriter.WriteMsgCallCount()).To(Equal(1))
			Expect(responseWriter.WriteMsgArgsForCall(0).Id).To(Equal(request.Id))
			Expect(responseWriter.WriteMsgArgsForCall(0).Rcode).To(Equal(dns.RcodeNameError))
		})

		It("should log the error", func() {
			httpResolver.ServeDNS(responseWriter, request)

			Expect(fakeLogger).To(gbytes.Say("record-not-found.*some-app-guid.potato."))
		})
	})

	Context("when getting the container from the ducati daemon errors", func() {
		Context("when the error is something else", func() {
			BeforeEach(func() {
				fakeDaemonClient.ListContainersReturns([]models.Container{}, errors.New("some server failure"))
			})

			It("should reply with SERVFAIL", func() {
				httpResolver.ServeDNS(responseWriter, request)

				Expect(responseWriter.WriteMsgCallCount()).To(Equal(1))
				Expect(responseWriter.WriteMsgArgsForCall(0).Id).To(Equal(request.Id))
				Expect(responseWriter.WriteMsgArgsForCall(0).Rcode).To(Equal(dns.RcodeServerFailure))
			})

			It("should log the error", func() {
				httpResolver.ServeDNS(responseWriter, request)

				Expect(fakeLogger).To(gbytes.Say("ducati-client-error"))
			})
		})
	})
})
