package client

import (
	"context"
	"net/url"
	"time"

	"golang.org/x/xerrors"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type pingMsg struct {
	LatencyMS int64  `json:"latency_ms"`
	Tolerable bool   `json:"tolerable"`
	Error     string `json:"error"`
}

// Ping determines the websocket latency of the agent connection
// to the server. A value of true is returned if the latency is
// tolerable.
func Ping(baseURL string) (time.Duration, bool, error) {
	var ctx = context.Background()

	uri, err := url.Parse(baseURL)
	if err != nil {
		return 0, false, xerrors.Errorf("parse url: %w", err)
	}
	uri.Path = "/latency"

	conn, _, err := websocket.Dial(context.Background(), uri.String(), nil)
	if err != nil {
		return 0, false, xerrors.Errorf("dial server: %w", err)
	}

	var msg pingMsg
	err = wsjson.Read(ctx, conn, &msg)
	if err != nil {
		return 0, false, xerrors.Errorf("read msg: %w", err)
	}
	if msg.Error != "" {
		return 0, false, xerrors.New(msg.Error)
	}

	return time.Duration(msg.LatencyMS) * time.Millisecond, msg.Tolerable, nil
}
