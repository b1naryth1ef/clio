package cliod

import (
	"io"
)

type Client struct{}

func ClientRegister(email string) Client {
	return Client{}
}

func ClientLogin(keyfile io.Reader, pw string) Client {
	return Client{}
}
