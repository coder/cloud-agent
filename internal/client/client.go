package client

import (
	"net/url"

	"go.coder.com/cloud-agent/internal/config"
)

type Client struct {
	token   string
	baseURL *url.URL
}

func FromEnv() (*Client, error) {
	token, err := config.SessionToken.Read()
	if err != nil {
		return nil, err
	}

	u, err := config.URL.Read()
	if err != nil {
		return nil, err
	}

	parsed, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	return &Client{
		token:   token,
		baseURL: parsed,
	}, nil
}
