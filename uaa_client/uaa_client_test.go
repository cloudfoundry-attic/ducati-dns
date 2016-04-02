package uaa_client_test

import (
	"errors"

	"github.com/cloudfoundry-incubator/ducati-dns/fakes"
	"github.com/cloudfoundry-incubator/ducati-dns/uaa_client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UAAClient", func() {
	var (
		client            uaa_client.Client
		fakeWarrantClient *fakes.WarrantClient
	)

	BeforeEach(func() {
		fakeWarrantClient = &fakes.WarrantClient{}
		fakeWarrantClient.GetTokenReturns("bear", nil)
		client = uaa_client.Client{
			Service: fakeWarrantClient,
			User:    "yogi",
			Secret:  "picnic",
		}
	})

	Describe("GetToken", func() {
		It("returns the token", func() {
			token, err := client.GetToken()
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeWarrantClient.GetTokenCallCount()).To(Equal(1))
			user, secret := fakeWarrantClient.GetTokenArgsForCall(0)
			Expect(user).To(Equal("yogi"))
			Expect(secret).To(Equal("picnic"))
			Expect(token).To(Equal("bear"))
		})

		Context("when it fails to get the token from the service", func() {
			BeforeEach(func() {
				fakeWarrantClient.GetTokenReturns("", errors.New("get token failed"))
			})
			It("returns an error", func() {
				_, err := client.GetToken()

				Expect(err).To(MatchError(ContainSubstring("get token failed")))
			})
		})
	})
})
