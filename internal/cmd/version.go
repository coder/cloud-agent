package cmd

import (
	"fmt"

	"github.com/spf13/pflag"

	"go.coder.com/cli"
)

var Version string

type versionCmd struct{}

func (c *versionCmd) Spec() cli.CommandSpec {
	return cli.CommandSpec{
		Name:  "version",
		Usage: "",
		Desc:  "Print the agent version.",
	}
}

func (c *versionCmd) Run(fl *pflag.FlagSet) {
	fmt.Println(Version)
}
