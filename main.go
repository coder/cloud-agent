package main

import (
	"go.coder.com/cli"
	"go.coder.com/cloud-agent/internal/cmd"
)

// Compile-time variables
var (
	DefaultCloudURL = "http://localhost:8080"
)

func main() {
	cli.RunRoot(cmd.Make())
}
