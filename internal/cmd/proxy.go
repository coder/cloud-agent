package cmd

import (
	"context"
	"os"
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

var codeserverPasswordEnv = "CODESERVER_PASSWORD"

type proxyCmd struct {
	codeServerAddr string
}

func (c *proxyCmd) Spec() cli.CommandSpec {
	return cli.CommandSpec{
		Name:  "proxy",
		Usage: "[FLAGS]",
		Desc:  "Proxy code-server to Coder-Cloud",
	}
}

func (c *proxyCmd) RegisterFlags(fl *pflag.FlagSet) {
	fl.StringVar(&c.codeServerAddr,
		"code-server-addr",
		"localhost:8080",
		"The address of the code-server instance to proxy.",
	)
}

func (c *proxyCmd) Run(fl *pflag.FlagSet) {
	var (
		ctx = context.Background()
	)

	cli, err := client.FromEnv()
	if err != nil {
		flog.Fatal("Failed to get client, have you logged in?")
	}

	csid, err := config.ServerID.Read()
	if err != nil {
		flog.Fatal("Failed to get server id: %v", err)
	}

	url, err := cli.AccessURL(csid)
	if err != nil {
		flog.Fatal("Failed to query server: %v", err)
	}

	conf, err := config.ReadFiles(config.ServerID, config.SessionToken, config.URL)
	if xerrors.Is(err, os.ErrNotExist) {
		flog.Fatal("Failed to read configuration files, have you logged in?")
	}
	if err != nil {
		flog.Fatal("Failed to read configuration file: %v", err.Error())
	}

	password := os.Getenv(codeserverPasswordEnv)

	agent := &ideproxy.Agent{
		Log:                sloghuman.Make(os.Stderr),
		CodeServerID:       conf[config.ServerID],
		SessionToken:       conf[config.SessionToken],
		CloudProxyURL:      conf[config.URL],
		CodeServerAddr:     c.codeServerAddr,
		CodeServerPassword: password,
	}

	proxy := func() {
		err = agent.Proxy(ctx)
		if err != nil {
			flog.Error("Connection to Coder-Cloud distrupted, re-establishing connection: %v", err.Error())
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
