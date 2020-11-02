package client

import (
	"bytes"
	"encoding/json"
	"net/http"

	"go.coder.com/cloud-agent/internal/version"
	"golang.org/x/xerrors"
)

const (
	sessionHeader = "Session-Token"
)

func (c *Client) request(method, path string, body interface{}) (*http.Response, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, xerrors.Errorf("marshal body: %w", err)
	}

	req, err := http.NewRequest(method, c.BaseURL.String()+path, bytes.NewReader(b))
	if err != nil {
		return nil, xerrors.Errorf("new request: %w", err)
	}

	req.Header.Set(sessionHeader, c.Token)
	req.Header.Set("User-Agent", userAgent())

	return http.DefaultClient.Do(req)
}

func (c *Client) requestBody(method, path string, request, response interface{}) error {
	resp, err := c.request(method, path, request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return bodyError(resp)
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return xerrors.Errorf("unmarshal response: %w", err)
	}
	return nil
}

type apiError struct {
	Err struct {
		Msg string `json:"msg"`
	} `json:"error"`
}

func bodyError(resp *http.Response) error {
	var apiErr apiError

	err := json.NewDecoder(resp.Body).Decode(&apiErr)
	if err != nil {
		return xerrors.Errorf("decode err: %w", err)
	}

	return xerrors.New(apiErr.Err.Msg)
}

func userAgent() string {
	return "CoderCloud/" + version.Version
}
