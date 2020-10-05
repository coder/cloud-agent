package cmd

import (
	"os"
	"regexp"

	"github.com/spf13/pflag"
	"golang.org/x/xerrors"

	"go.coder.com/cli"
	"go.coder.com/cloud-agent/internal/client"
	"go.coder.com/cloud-agent/internal/config"
	"go.coder.com/flog"
)

var DefaultCloudURL string = "https://cloud.coder.com"

var codeServerNameRx = regexp.MustCompile("^[a-z][a-z0-9_]{0,50}$")

type linkCmd struct {
	cloudURL string
}

func (c *linkCmd) Spec() cli.CommandSpec {
	return cli.CommandSpec{
		Name:  "link",
		Usage: "NAME [FLAGS]",
		Desc:  "Link a server with Coder Cloud",
	}
}

func (c *linkCmd) RegisterFlags(fl *pflag.FlagSet) {
	// TODO: this should be updated whenever we figure out the domain we're
	// using.
	fl.StringVar(&c.cloudURL, "cloud-url", DefaultCloudURL, "The Coder Cloud URL to connect to.")
}

func (c *linkCmd) Run(fl *pflag.FlagSet) {
	name := fl.Arg(0)
	if name == "" {
		flog.Fatal("Must provide a name")
	}

	if !codeServerNameRx.MatchString(name) {
		flog.Fatal("Name must conform to regex %s", codeServerNameRx.String())
	}

	_, err := config.ServerID.Read()
	if err == nil {
		flog.Info("Server already registered!")
		return
	}

	cli, err := client.FromEnv()
	if xerrors.Is(err, os.ErrNotExist) {
		cli, err = loginClient(c.cloudURL, name)
	}
	if err != nil {
		flog.Fatal("Failed to login: %v", err)
	}

	cs, err := cli.RegisterCodeServer(name)
	if err != nil {
		flog.Fatal("Failed to register server: %v", err)
	}

	err = config.ServerID.Write(cs.ID)
	if err != nil {
		flog.Fatal("Failed to store server id: %v", err)
	}
	flog.Success("Successfully registered server!")
}

func loginClient(url, serverName string) (*client.Client, error) {
	token, err := client.Login(url, serverName)
	if err != nil {
		return nil, xerrors.Errorf("unable to login: %w", err)
	}

	err = config.SessionToken.Write(token)
	if err != nil {
		return nil, xerrors.Errorf("write session token to file: %w", err)
	}

	err = config.URL.Write(url)
	if err != nil {
		return nil, xerrors.Errorf("write coder-cloud url to file: %w", err)
	}

	return client.FromEnv()
}
