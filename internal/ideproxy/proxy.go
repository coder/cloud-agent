package ideproxy

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	"cdr.dev/slog"
	"github.com/hashicorp/yamux"
	"golang.org/x/xerrors"
	"nhooyr.io/websocket"
)

const sessionHeader = "Session-Token"

// Agent is the agent running on a user's personal machine.
type Agent struct {
	Log                slog.Logger
	CodeServerID       string
	SessionToken       string
	CodeServerAddr     string
	CodeServerPassword string
	CloudProxyURL      string
}

// Proxy proxies a Coder Cloud connection to a local code server instance.
func (a *Agent) Proxy(ctx context.Context) error {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return xerrors.Errorf("listen on local port: %w", err)
	}
	defer l.Close()

	go func() {
		err := http.Serve(l,
			codeServerReverseProxy(a.CodeServerAddr, a.CodeServerPassword))
		a.Log.Warn(ctx, "code-server proxy exited", slog.Error(err))
	}()

	ws, err := dialProxy(ctx, a.CloudProxyURL, a.CodeServerID, a.SessionToken)
	if err != nil && !xerrors.Is(err, io.EOF) {
		return xerrors.Errorf("dial proxy: %w", err)
	}

	conn := websocket.NetConn(ctx, ws, websocket.MessageBinary)

	err = proxyCodeServer(ctx, a.Log, conn, l.Addr().String())
	if err != nil && !xerrors.Is(err, io.EOF) {
		return xerrors.Errorf("proxy code-server: %w", err)
	}
	return nil
}

// proxyCodeServer proxies a Coder Cloud connection to the local code-server.
func proxyCodeServer(ctx context.Context, log slog.Logger, proxyConn net.Conn, addr string) error {
	stream, err := yamux.Server(proxyConn, nil)
	if err != nil {
		return xerrors.Errorf("multiplex stream: %w", err)
	}

	for {
		conn, err := stream.Accept()
		if err != nil {
			return xerrors.Errorf("accept stream: %w", err)
		}

		go func() {
			csConn, err := net.Dial("tcp", addr)
			if err != nil {
				log.Error(ctx, "dial code-server", slog.Error(err))
				return
			}
			// Bicopy closes the streams.
			bicopy(ctx, csConn, conn)
		}()
	}
}

func codeServerReverseProxy(addr, password string) http.Handler {
	rp := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   addr,
	})

	dir := rp.Director
	rp.Director = func(r *http.Request) {
		if password != "" {
			r.AddCookie(&http.Cookie{
				Name:  "key",
				Value: fmt.Sprintf("%x", sha256.Sum256([]byte(password))),
			})
		}
		dir(r)
	}

	return rp
}

// dialProxy dials the Coder Cloud proxy.
func dialProxy(ctx context.Context, addr, serverID, token string) (*websocket.Conn, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, xerrors.Errorf("parse coder cloud url: %w", err)
	}
	ws, _, err := websocket.Dial(ctx, //nolint:bodyclose
		fmt.Sprintf("%v://%v/proxy/ide/%v/server",
			u.Scheme,
			u.Host,
			serverID),
		&websocket.DialOptions{
			HTTPHeader: http.Header{
				sessionHeader: []string{token},
			},
		})
	if err != nil {
		return nil, xerrors.Errorf("dial cproxy: %w", err)
	}
	return ws, nil
}

// bicopy copies all of the data between the two connections
// and will close them after one or both of them are done writing.
// If the context is cancelled, both of the connections will be
// closed.
//
// NOTE: This function will block until the copying is done or the
// context is canceled.
func bicopy(ctx context.Context, c1, c2 io.ReadWriteCloser) {
	defer c1.Close()
	defer c2.Close()

	ctx, cancel := context.WithCancel(ctx)

	copy := func(dst io.WriteCloser, src io.Reader) {
		defer cancel()
		_, _ = io.Copy(dst, src)
	}

	go copy(c1, c2)
	go copy(c2, c1)

	<-ctx.Done()
}
