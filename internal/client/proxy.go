package client

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/xerrors"
	"nhooyr.io/websocket"
)

func (c *Client) ProxyAgent(ctx context.Context, id string) (*websocket.Conn, error) {
	ws, resp, err := websocket.Dial(ctx, //nolint:bodyclose
		fmt.Sprintf("%v/proxy/ide/%v/server",
			c.BaseURL.String(),
			id,
		),
		&websocket.DialOptions{
			HTTPHeader: http.Header{
				"User-Agent":  []string{userAgent()},
				sessionHeader: []string{c.Token},
			},
		})
	if resp != nil && resp.StatusCode != http.StatusSwitchingProtocols {
		return nil, bodyError(resp)
	}
	if err != nil {
		return nil, xerrors.Errorf("dial cproxy: %w", err)
	}

	return ws, nil
}
