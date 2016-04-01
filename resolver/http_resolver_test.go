package resolver_test

import (
	"errors"
	"math/rand"
	"net"

	"github.com/cloudfoundry-incubator/ducati-daemon/client"
	"github.com/cloudfoundry-incubator/ducati-daemon/models"
	"github.com/cloudfoundry-incubator/ducati-dns/cc_client"
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
		fakeCCClient     *fakes.CCClient
	)

	BeforeEach(func() {
		request = &dns.Msg{
			MsgHdr: dns.MsgHdr{
				Id: uint16(rand.Int()),
			},
		}
		request.SetQuestion(dns.Fqdn("some-app.some-space.some-org.potato"), dns.TypeA)
		fakeLogger = lagertest.NewTestLogger("test")
		fakeDaemonClient = &fakes.DucatiDaemonClient{}
		fakeDaemonClient.GetContainerReturns(models.Container{
			IP: "10.11.12.13",
		}, nil)
		fakeCCClient = &fakes.CCClient{}
		fakeCCClient.GetAppGuidReturns("my-container-id", nil)
		httpResolver = &resolver.HTTPResolver{
			Suffix:       "potato",
			DaemonClient: fakeDaemonClient,
			CCClient:     fakeCCClient,
			TTL:          42,
			Logger:       fakeLogger,
		}
		responseWriter = &fakes.ResponseWriter{}
	})

	It("resolves DNS queries by using the ducati daemon client and ccclient", func() {
		httpResolver.ServeDNS(responseWriter, request)

		Expect(fakeCCClient.GetAppGuidCallCount()).To(Equal(1))
		appInCall, spaceInCall, orgInCall := fakeCCClient.GetAppGuidArgsForCall(0)
		Expect(appInCall).To(Equal("some-app"))
		Expect(spaceInCall).To(Equal("some-space"))
		Expect(orgInCall).To(Equal("some-org"))

		Expect(fakeDaemonClient.GetContainerCallCount()).To(Equal(1))
		Expect(fakeDaemonClient.GetContainerArgsForCall(0)).To(Equal("my-container-id"))

		Expect(responseWriter.WriteMsgCallCount()).To(Equal(1))

		expectedResp := &dns.Msg{}
		expectedResp.SetReply(request)
		rr_header := dns.RR_Header{
			Name:   dns.Fqdn("some-app.some-space.some-org.potato"),
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    42,
		}
		a := &dns.A{rr_header, net.ParseIP("10.11.12.13")}
		expectedResp.Answer = []dns.RR{a}
		Expect(responseWriter.WriteMsgArgsForCall(0)).To(Equal(expectedResp))
	})

	Context("when the cc client errors", func() {
		BeforeEach(func() {
			fakeCCClient.GetAppGuidReturns("", errors.New("pineapple"))
		})
		It("should reply with SERVFAIL", func() {
			httpResolver.ServeDNS(responseWriter, request)

			Expect(responseWriter.WriteMsgCallCount()).To(Equal(1))
			Expect(responseWriter.WriteMsgArgsForCall(0).Id).To(Equal(request.Id))
			Expect(responseWriter.WriteMsgArgsForCall(0).Rcode).To(Equal(dns.RcodeServerFailure))
		})
		It("logs the error", func() {
			httpResolver.ServeDNS(responseWriter, request)

			Expect(fakeLogger).To(gbytes.Say("cloud-controller-client-error.*pineapple"))
		})
	})

	Context("when the cc client errors with DomainNotFoundError", func() {
		BeforeEach(func() {
			fakeCCClient.GetAppGuidReturns("", cc_client.DomainNotFoundError)
		})

		It("should reply with NXDOMAIN", func() {
			httpResolver.ServeDNS(responseWriter, request)

			Expect(responseWriter.WriteMsgCallCount()).To(Equal(1))
			Expect(responseWriter.WriteMsgArgsForCall(0).Id).To(Equal(request.Id))
			Expect(responseWriter.WriteMsgArgsForCall(0).Rcode).To(Equal(dns.RcodeNameError))
		})
		It("should log the error", func() {
			httpResolver.ServeDNS(responseWriter, request)

			Expect(fakeLogger).To(gbytes.Say("domain-not-found.*some-app.some-space.some-org.potato."))
		})
	})

	Context("when the domain is too long", func() {
		BeforeEach(func() {
			request.SetQuestion(dns.Fqdn("not.the.right.number.of.things.potato"), dns.TypeA)
		})

		It("should reply with NXDOMAIN", func() {
			httpResolver.ServeDNS(responseWriter, request)

			Expect(responseWriter.WriteMsgCallCount()).To(Equal(1))
			Expect(responseWriter.WriteMsgArgsForCall(0).Id).To(Equal(request.Id))
			Expect(responseWriter.WriteMsgArgsForCall(0).Rcode).To(Equal(dns.RcodeNameError))
		})
		It("should log the error", func() {
			httpResolver.ServeDNS(responseWriter, request)

			Expect(fakeLogger).To(gbytes.Say("invalid-domain.*not.the.right.number.of.things.potato."))
		})
	})

	Context("when the domain is too short", func() {
		BeforeEach(func() {
			request.SetQuestion(dns.Fqdn("short.potato"), dns.TypeA)
		})

		It("should reply with NXDOMAIN", func() {
			httpResolver.ServeDNS(responseWriter, request)

			Expect(responseWriter.WriteMsgCallCount()).To(Equal(1))
			Expect(responseWriter.WriteMsgArgsForCall(0).Id).To(Equal(request.Id))
			Expect(responseWriter.WriteMsgArgsForCall(0).Rcode).To(Equal(dns.RcodeNameError))
		})
		It("should log the error", func() {
			httpResolver.ServeDNS(responseWriter, request)

			Expect(fakeLogger).To(gbytes.Say("invalid-domain.*short.potato."))
		})
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

	Context("when getting the container from the ducati daemon errors", func() {
		Context("when the error is a client.RecordNotFound error", func() {
			BeforeEach(func() {
				fakeDaemonClient.GetContainerReturns(models.Container{}, client.RecordNotFoundError)
			})
			It("should reply with NXDOMAIN", func() {
				httpResolver.ServeDNS(responseWriter, request)

				Expect(responseWriter.WriteMsgCallCount()).To(Equal(1))
				Expect(responseWriter.WriteMsgArgsForCall(0).Id).To(Equal(request.Id))
				Expect(responseWriter.WriteMsgArgsForCall(0).Rcode).To(Equal(dns.RcodeNameError))
			})
			It("should log the error", func() {
				httpResolver.ServeDNS(responseWriter, request)

				Expect(fakeLogger).To(gbytes.Say("record-not-found.*some-app.some-space.some-org.potato."))
			})
		})
		Context("when the error is something else", func() {
			BeforeEach(func() {
				fakeDaemonClient.GetContainerReturns(models.Container{}, errors.New("some server failure"))
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
