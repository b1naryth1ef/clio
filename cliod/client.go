package cliod

import (
	"code.google.com/p/go.crypto/openpgp"
)

type Client struct {
	Entity *openpgp.Entity
	pw     string
}

func ClientLogin(e *openpgp.Entity, pw string) Client {
	return Client{
		Entity: e,
	}
}
