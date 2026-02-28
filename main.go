// Package main is the entry point for the etoro-cli application.
// etoro-cli is a production-quality CLI for eToro trading automation,
// designed to be agent-friendly with full JSON output support.
package main

import (
	"github.com/lijinlar/etoro-cli/cmd"
)

func main() {
	cmd.Execute()
}
