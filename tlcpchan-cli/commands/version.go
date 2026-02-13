package commands

import (
	"fmt"
)

var cliVersion string

func versionCmd(args []string) error {
	fmt.Printf("tlcpchan-cli 版本: %s\n", cliVersion)
	return nil
}
