package main

import (
	"os"

	"github.com/Trisia/tlcpchan-cli/commands"
)

var version = "1.0.0"

func main() {
	if err := commands.Execute(version); err != nil {
		os.Exit(1)
	}
}
