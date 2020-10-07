package cmd

import (
	"github.com/spf13/pflag"
	"go.coder.com/cli"
)

func Make() cli.Command {
	return &rootCmd{}
}

var _ interface {
	cli.Command
	cli.ParentCommand
} = &rootCmd{}

type rootCmd struct {
}

func (c *rootCmd) Spec() cli.CommandSpec {
	return cli.CommandSpec{
		Name:  "agent",
		Usage: "[GLOBAL FLAGS] COMMAND [COMMAND FLAGS] [ARGS...]",
		Desc:  `Run the Coder Cloud Agent.`,
	}
}

func (c *rootCmd) Subcommands() []cli.Command {
	return []cli.Command{
		&bindCmd{},
		&versionCmd{},
	}
}

func (c *rootCmd) Run(fl *pflag.FlagSet) {
	fl.Usage()
}
