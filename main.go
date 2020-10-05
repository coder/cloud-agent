package main

import (
	"go.coder.com/cli"
	"go.coder.com/cloud-agent/internal/cmd"
)

func main() {
	cli.RunRoot(cmd.Make())
}
