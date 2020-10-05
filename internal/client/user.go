package client

import (
	"time"
)

// User describe a Coder Cloud user.
type User struct {
	ID        string    `json:"id" `
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func (c *Client) Me() (*User, error) {
	const path = "/api/users/me"

	var response User
	err := c.requestBody("GET", path, nil, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil

}
