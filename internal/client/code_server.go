package client

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/xerrors"
)

func (c *Client) CodeServer(id string) (*CodeServer, error) {
	path := fmt.Sprintf("/api/servers/%v", id)

	var response CodeServer
	err := c.requestBody("GET", path, nil, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

type AccessURLResponse struct {
	URL string `json:"url"`
}

func (c *Client) AccessURL(id string) (string, error) {
	path := fmt.Sprintf("/api/servers/%v/access-url", id)

	var response AccessURLResponse
	err := c.requestBody("GET", path, nil, &response)
	if err != nil {
		return "", err
	}

	return response.URL, nil
}

type CodeServer struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	Name             string    `json:"name"`
	Hostname         string    `json:"hostname"`
	CreatedAt        time.Time `json:"created_at"`
	LastConnectionAt time.Time `json:"last_connection_at"`
}

// RegisterServerRequest is the request body sent in a
// register server request.
type RegisterServerRequest struct {
	Name     string `json:"name"`
	Hostname string `json:"hostname"`
}

func (c *Client) RegisterCodeServer(name string) (*CodeServer, error) {
	const path = "/api/servers"
	hostname, err := os.Hostname()
	if err != nil {
		return nil, xerrors.Errorf("get hostname: %w", err)
	}

	var response CodeServer
	err = c.requestBody("POST", path,
		&RegisterServerRequest{
			Name:     name,
			Hostname: hostname,
		},
		&response,
	)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
