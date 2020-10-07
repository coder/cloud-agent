package main

import (
	"log"

	"go.coder.com/cli"
	"go.coder.com/cloud-agent/internal/cmd"
)

func main() {
	log.SetFlags(0)
	cli.RunRoot(cmd.Make())
}
