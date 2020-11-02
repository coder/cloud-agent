package client

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/browser"
	"golang.org/x/xerrors"
	"nhooyr.io/websocket"

	"go.coder.com/cloud-agent/internal/version"
	"go.coder.com/cloud-agent/pkg/agentlogin"
	"go.coder.com/flog"
)

func init() {
	browser.Stderr = ioutil.Discard
	browser.Stdout = ioutil.Discard
}

// Login performs the login flow for an agent. It returns the resulting
// session token to use for authenticated routes.
func Login(addr, serverName string) (string, error) {
	ctx := context.Background()

	u, err := url.Parse(addr)
	if err != nil {
		return "", err
	}

	query := url.Values{}
	query.Add(agentlogin.ServerNameQueryParam, serverName)

	loginURL := &url.URL{
		Scheme:   u.Scheme,
		Host:     u.Host,
		Path:     "/login",
		RawQuery: query.Encode(),
	}

	conn, resp, err := websocket.Dial(ctx, loginURL.String(), &websocket.DialOptions{
		HTTPHeader: http.Header{
			agentVersionHeader: []string{version.Version},
		},
	})
	if resp.StatusCode != http.StatusSwitchingProtocols {
		return "", bodyError(resp)
	}
	if err != nil {
		return "", err
	}
	defer conn.Close(websocket.StatusInternalError, "")

	client := &agentlogin.Client{
		Ctx:  ctx,
		Conn: conn,
	}

	url, err := client.ReadAuthURL()
	if err != nil {
		return "", xerrors.Errorf("read auth url: %w", err)
	}

	err = browser.OpenURL(url)
	if err != nil {
		flog.Info("visit %s to login", url)
	}

	token, err := client.ReadSessionToken()
	if err != nil {
		return "", xerrors.Errorf("read session token: %w", err)
	}

	return token, nil
}
