package uaa_client

import "fmt"

//go:generate counterfeiter -o ../fakes/warrant_client.go --fake-name WarrantClient . warrantClient
type warrantClient interface {
	GetToken(user string, secret string) (string, error)
}

type Client struct {
	Service warrantClient
	User    string
	Secret  string
}

func (c *Client) GetToken() (string, error) {
	token, err := c.Service.GetToken(c.User, c.Secret)
	if err != nil {
		return "", fmt.Errorf("get token:", err)
	}

	return token, nil
}
