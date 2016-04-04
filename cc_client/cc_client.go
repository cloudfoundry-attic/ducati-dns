package cc_client

import (
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

//go:generate counterfeiter -o ../fakes/uaa_client_service.go --fake-name UAAClientService . uaaClientService
type uaaClientService interface {
	GetToken() (string, error)
}

type Client struct {
	Org   orgService
	Space spaceService
	App   appService
	UAA   uaaClientService
}

type NotFoundError struct {
	Msg string
}

func (n *NotFoundError) Error() string {
	if n != nil && n.Msg != "" {
		return n.Msg
	} else {
		return "not found"
	}
}

var OrgNotFoundError = &NotFoundError{"org not found"}
var SpaceNotFoundError = &NotFoundError{"space not found"}
var AppNotFoundError = &NotFoundError{"app not found"}

func (c *Client) GetAppGuid(appName string, spaceName string, orgName string) (string, error) {
	var orgGuid string
	var spaceGuid string
	var appGuid string

	token, err := c.UAA.GetToken()
	if err != nil {
		return "", fmt.Errorf("uaa client: %s", err)
	}

	orgList, err := c.Org.List(token)
	if err != nil {
		return "", fmt.Errorf("cc client: %s", err)
	}

	for _, org := range orgList.Organizations {
		if org.Name == orgName {
			orgGuid = org.GUID
		}
	}

	if orgGuid == "" {
		return "", OrgNotFoundError
	}

	spaceList, err := c.Space.List(token)
	if err != nil {
		return "", fmt.Errorf("cc client: %s", err)
	}

	for _, space := range spaceList.Spaces {
		if space.OrganizationGUID == orgGuid && space.Name == spaceName {
			spaceGuid = space.GUID
		}
	}

	if spaceGuid == "" {
		return "", SpaceNotFoundError
	}

	appList, err := c.App.List(token)
	if err != nil {
		return "", fmt.Errorf("cc client: %s", err)
	}

	for _, app := range appList.Applications {
		if app.SpaceGUID == spaceGuid && app.Name == appName {
			appGuid = app.GUID
		}
	}

	if appGuid == "" {
		return "", AppNotFoundError
	}

	return appGuid, nil
}
