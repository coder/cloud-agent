package client

import (
	"net/url"
)

type Client struct {
	Token   string
	BaseURL *url.URL
}
