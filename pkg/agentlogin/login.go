package agentlogin

import (
	"context"
	"io"

	"cdr.dev/slog"
	"golang.org/x/xerrors"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// ServerNameQueryParam is the query parameter indicating the
// name of the server that login is being performed for. It is mainly
// used as a mechanism for redirecting the user to the server IDE once
// authentication is successful.
const ServerNameQueryParam = "server_name"

// Server implements the server-side for the agent login flow.
type Server struct {
	Ctx  context.Context
	Conn *websocket.Conn
	Log  slog.Logger
}

const (
	msgTypeAuthURL = "auth_url"
	msgTypeError   = "error"
	msgTypeToken   = "token"
)

type loginMsg struct {
	Type string `json:"type"`
	Msg  string `json:"msg"`
}

// WriteAuthURL writes the Auth Code URL to the websocket.
func (s *Server) WriteAuthURL(url string) bool {
	return write(s.Ctx, s.Log, s.Conn, loginMsg{
		Type: msgTypeAuthURL,
		Msg:  url,
	})
}

// WriteError writes an error that occurred during the login
// process to the client.
func (s *Server) WriteError(err string) bool {
	return write(s.Ctx, s.Log, s.Conn, loginMsg{
		Type: msgTypeError,
		Msg:  err,
	})
}

// WriteSessionToken writes the session token to the webscoket.
func (s *Server) WriteSessionToken(token string) bool {
	return write(s.Ctx, s.Log, s.Conn, loginMsg{
		Type: msgTypeToken,
		Msg:  token,
	})
}

// Client implements the client-side of the agent login
// flow.
type Client struct {
	Ctx  context.Context
	Conn *websocket.Conn
}

// ReadAuthURL reads the auth code URL endpoint from the websocket.
func (c *Client) ReadAuthURL() (string, error) {
	return readLoginMsg(c.Ctx, c.Conn, msgTypeAuthURL)
}

// ReadSessionToken reads the session token that is created from a successful
// login from the websocket.
func (c *Client) ReadSessionToken() (string, error) {
	return readLoginMsg(c.Ctx, c.Conn, msgTypeToken)
}

func readLoginMsg(ctx context.Context, c *websocket.Conn, msgType string) (string, error) {
	var msg loginMsg

	err := wsjson.Read(ctx, c, &msg)
	if err != nil {
		return "", xerrors.Errorf("read msg: %w", err)
	}
	if msg.Type == msgTypeError {
		return "", xerrors.New(msg.Msg)
	}
	if msg.Type != msgType {
		return "", xerrors.Errorf("unexpected message type %v", msg.Type)
	}

	return msg.Msg, nil
}

// Write writes the provided message to the connection, logging and returning false if an error occurs.
func write(ctx context.Context, log slog.Logger, c *websocket.Conn, msg interface{}) bool {
	err := wsjson.Write(ctx, c, msg)
	if err != nil {
		logLevel(log, err)(ctx, "write websocket message",
			slog.F("msg", msg),
			slog.Error(err),
		)
	}

	return err == nil
}

func logLevel(log slog.Logger, err error) func(context.Context, string, ...slog.Field) {
	if xerrors.Is(err, io.EOF) || xerrors.Is(err, context.Canceled) {
		return log.Warn
	}
	var closeErr websocket.CloseError
	if xerrors.As(err, &closeErr) {
		switch closeErr.Code {
		case
			websocket.StatusNoStatusRcvd,
			websocket.StatusGoingAway,
			websocket.StatusNormalClosure:
			return log.Warn
		}
	}
	return log.Error
}
