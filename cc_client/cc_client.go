package cc_client

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf-experimental/rainmaker"
)

//go:generate counterfeiter -o ../fakes/org_service.go --fake-name OrgService . orgService
type orgService interface {
	List(token string) (rainmaker.OrganizationsList, error)
}

//go:generate counterfeiter -o ../fakes/space_service.go --fake-name SpaceService . spaceService
type spaceService interface {
	List(token string) (rainmaker.SpacesList, error)
}

//go:generate counterfeiter -o ../fakes/app_service.go --fake-name AppService . appService
type appService interface {
	List(token string) (rainmaker.ApplicationsList, error)
}

type Client struct {
	Org      orgService
	Space    spaceService
	App      appService
	UAAToken string
}

var DomainNotFoundError error = errors.New("domain not found")

func (c *Client) GetAppGuid(appName string, spaceName string, orgName string) (string, error) {
	var orgGuid string
	var spaceGuid string
	var appGuid string

	orgList, err := c.Org.List(c.UAAToken)
	if err != nil {
		return "", fmt.Errorf("cc client: %s", err)
	}

	for _, org := range orgList.Organizations {
		if org.Name == orgName {
			orgGuid = org.GUID
		}
	}

	if orgGuid == "" {
		return "", DomainNotFoundError
	}

	spaceList, err := c.Space.List(c.UAAToken)
	if err != nil {
		return "", fmt.Errorf("cc client: %s", err)
	}

	for _, space := range spaceList.Spaces {
		if space.OrganizationGUID == orgGuid && space.Name == spaceName {
			spaceGuid = space.GUID
		}
	}

	if spaceGuid == "" {
		return "", DomainNotFoundError
	}

	appList, err := c.App.List(c.UAAToken)
	if err != nil {
		return "", fmt.Errorf("cc client: %s", err)
	}

	for _, app := range appList.Applications {
		if app.SpaceGUID == spaceGuid && app.Name == appName {
			appGuid = app.GUID
		}
	}

	if appGuid == "" {
		return "", DomainNotFoundError
	}

	return appGuid, nil
}
