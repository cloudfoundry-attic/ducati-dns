package cc_client_test

import (
	"errors"

	"github.com/cloudfoundry-incubator/ducati-dns/cc_client"
	"github.com/cloudfoundry-incubator/ducati-dns/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/rainmaker"
)

var _ = Describe("CCClient", func() {
	var (
		fakeOrgService   *fakes.OrgService
		fakeSpaceService *fakes.SpaceService
		fakeAppService   *fakes.AppService
		client           cc_client.Client
	)

	BeforeEach(func() {
		fakeOrgService = &fakes.OrgService{}
		fakeSpaceService = &fakes.SpaceService{}
		fakeAppService = &fakes.AppService{}
		client = cc_client.Client{
			Org:      fakeOrgService,
			Space:    fakeSpaceService,
			App:      fakeAppService,
			UAAToken: "my-token",
		}
		orgList := rainmaker.OrganizationsList{
			Organizations: []rainmaker.Organization{
				{
					Name: "some-org",
					GUID: "org-guid",
				},
			},
		}
		spaceList := rainmaker.SpacesList{
			Spaces: []rainmaker.Space{
				{
					Name:             "some-space",
					GUID:             "space-guid",
					OrganizationGUID: "org-guid",
				},
			},
		}
		appList := rainmaker.ApplicationsList{
			Applications: []rainmaker.Application{
				{
					Name:      "some-app",
					GUID:      "my-container-id",
					SpaceGUID: "space-guid",
				},
			},
		}
		fakeOrgService.ListReturns(orgList, nil)
		fakeSpaceService.ListReturns(spaceList, nil)
		fakeAppService.ListReturns(appList, nil)
	})

	Describe("GetAppGuid", func() {
		It("Uses the rainmaker client to get the app guid for a given app name, space and org", func() {
			appGuid, err := client.GetAppGuid("some-app", "some-space", "some-org")
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeOrgService.ListCallCount()).To(Equal(1))
			Expect(fakeOrgService.ListArgsForCall(0)).To(Equal("my-token"))
			Expect(fakeSpaceService.ListCallCount()).To(Equal(1))
			Expect(fakeSpaceService.ListArgsForCall(0)).To(Equal("my-token"))
			Expect(fakeAppService.ListCallCount()).To(Equal(1))
			Expect(fakeAppService.ListArgsForCall(0)).To(Equal("my-token"))
			Expect(appGuid).To(Equal("my-container-id"))
		})
	})

	Context("when rainmaker client errors", func() {
		Context("when looking up organizations", func() {
			BeforeEach(func() {
				fakeOrgService.ListReturns(rainmaker.OrganizationsList{}, errors.New("org list failure"))
			})

			It("returns an error", func() {
				_, err := client.GetAppGuid("some-app", "some-space", "some-org")

				Expect(err).To(MatchError(ContainSubstring("org list failure")))
			})
		})

		Context("when looking up spaces", func() {
			BeforeEach(func() {
				fakeSpaceService.ListReturns(rainmaker.SpacesList{}, errors.New("space list failure"))
			})

			It("returns an error", func() {
				_, err := client.GetAppGuid("some-app", "some-space", "some-org")

				Expect(err).To(MatchError(ContainSubstring("space list failure")))
			})
		})

		Context("when looking up apps", func() {
			BeforeEach(func() {
				fakeAppService.ListReturns(rainmaker.ApplicationsList{}, errors.New("app list failure"))
			})

			It("returns an error", func() {
				_, err := client.GetAppGuid("some-app", "some-space", "some-org")

				Expect(err).To(MatchError(ContainSubstring("app list failure")))
			})
		})
	})

	Context("when org guid can not be found", func() {
		It("returns a not found error", func() {
			_, err := client.GetAppGuid("some-app", "some-space", "some-org-i-dont-want")
			Expect(err).To(MatchError(cc_client.DomainNotFoundError))
		})
	})
	Context("when space guid can not be found", func() {
		It("returns an empty string for the app guid", func() {
			_, err := client.GetAppGuid("some-app", "some-space-i-dont-want", "some-org")
			Expect(err).To(MatchError(cc_client.DomainNotFoundError))
		})
	})
	Context("when app guid can not be found", func() {
		It("returns an empty string for the app guid", func() {
			_, err := client.GetAppGuid("some-app-i-dont-have", "some-space", "some-org")
			Expect(err).To(MatchError(cc_client.DomainNotFoundError))
		})
	})
})
