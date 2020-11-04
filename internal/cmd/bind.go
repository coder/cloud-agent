package cmd

import (
	"context"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"cdr.dev/slog/sloggers/sloghuman"
	"github.com/spf13/pflag"
	"golang.org/x/xerrors"

	"go.coder.com/cli"
	"go.coder.com/cloud-agent/internal/client"
	"go.coder.com/cloud-agent/internal/config"
	"go.coder.com/cloud-agent/internal/ideproxy"
	"go.coder.com/flog"
)

var (
	DefaultCloudURL = "https://cloud.coder.com"
)

var codeServerNameRx = regexp.MustCompile("^[a-z][a-z0-9_]{0,50}$")

type bindCmd struct {
	cloudURL       string
	codeServerAddr string
}

func (c *bindCmd) Spec() cli.CommandSpec {
	return cli.CommandSpec{
		Name:  "bind",
		Usage: "[NAME]",
		Desc:  "Bind a server to Coder Cloud. A name will be generated from the hostname if one is not provided.",
	}
}

func (c *bindCmd) RegisterFlags(fl *pflag.FlagSet) {
	fl.StringVar(&c.cloudURL, "cloud-url", DefaultCloudURL, "The Coder Cloud URL to connect to.")
	fl.StringVar(&c.codeServerAddr,
		"code-server-addr",
		"localhost:8080",
		"The address of the code-server instance to proxy.",
	)
}

func (c *bindCmd) Run(fl *pflag.FlagSet) {
	var (
		err error
		ctx = context.Background()
	)

	name := fl.Arg(0)
	if name == "" {
		// Generate a name based on the hostname if one is not provided.
		name, err = genServerName()
		if err != nil {
			flog.Fatal("Failed to generate server name: %v", err.Error())
		}
	}

	if !codeServerNameRx.MatchString(name) {
		flog.Fatal("Name must conform to regex %s", codeServerNameRx.String())
	}

	cloudURL, err := url.Parse(c.cloudURL)
	if err != nil {
		flog.Fatal("Invalid Cloud URL: %v", err.Error())
	}

	token, err := config.SessionToken.Read()
	if xerrors.Is(err, os.ErrNotExist) {
		checkLatency(c.cloudURL)
		token, err = login(cloudURL.String(), name)
	}
	if err != nil {
		flog.Fatal("Failed to login: %v", err)
	}

	cli := client.Client{
		Token:   token,
		BaseURL: cloudURL,
	}

	// Register the server with Coder Cloud. This is an idempotent
	// operation.
	cs, err := cli.RegisterCodeServer(name)
	if err != nil {
		flog.Fatal("Failed to register server: %v", err)
	}

	// Get the Access URL for the user.
	url, err := cli.AccessURL(cs.ID)
	if err != nil {
		flog.Fatal("Failed to query server: %v", err)
	}

	agent := &ideproxy.Agent{
		Log:            sloghuman.Make(os.Stderr),
		CodeServerID:   cs.ID,
		SessionToken:   token,
		CloudProxyURL:  c.cloudURL,
		CodeServerAddr: c.codeServerAddr,
	}

	proxy := func() {
		err = agent.Proxy(ctx)
		if err != nil {
			flog.Error("Connection to Coder-Cloud disrupted, re-establishing connection: %v", err.Error())
		}
	}

	flog.Info("Proxying code-server to Coder Cloud, you can access your IDE at %v", url)

	proxy()

	// Avoid a super tight loop.
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		proxy()
	}
}

func login(url, serverName string) (string, error) {
	token, err := client.Login(url, serverName)
	if err != nil {
		return "", xerrors.Errorf("unable to login: %w", err)
	}

	err = config.SessionToken.Write(token)
	if err != nil {
		return "", xerrors.Errorf("write session token to file: %w", err)
	}

	return token, nil
}

func genServerName() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		xerrors.Errorf("get hostname: %w", err)
	}

	hostname = strings.ToLower(hostname)

	// Only use the first token.
	hostname = strings.Split(hostname, ".")[0]
	// '-' are not allowed, convert them to '_'.
	return strings.Replace(hostname, "-", "_", -1), nil
}

func checkLatency(cloudURL string) {
	latency, tolerable, err := client.Ping(cloudURL)
	if err != nil {
		flog.Fatal("ping server: %s", err.Error())
	}

	if !tolerable {
		flog.Fatal("Unfortunately we cannot ensure a good user experience with your connection latency (%s). Efforts are underway to accommodate users in most areas.", latency)
	}

	flog.Info("Detected an acceptable latency of %s", latency)
}
