package cliod

import (
	"code.google.com/p/go.crypto/openpgp"
)

type Client struct {
	Ident *openpgp.Entity
	NC    *NetClient
	CI    *ClientInterface
}

type ClientInterface interface {
	QueryPassword() string
}

type DefaultClientInterface struct{}

func (DCI *DefaultClientInterface) QueryPassword() string {
	return ""
}

type PasswordPolicy interface {
	GetPassword(source string) string
	Init(cl *Client)
}

// This is a defulat password policy which requires authentication on login only
type InitialPasswordPolicy struct {
	cached string
	client *Client
}

func (IPP *InitialPasswordPolicy) Init(cl *Client) {
	IPP.client = cl
}

func (IPP *InitialPasswordPolicy) GetPassword(source string) string {
	// If this is an initial login, grab the password
	if source == "login" {
		IPP.cached = IPP.client.CI.QueryPassword()
	}

	return IPP.cached
}
