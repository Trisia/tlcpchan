package commands

import (
	"fmt"
	"os"
	"text/tabwriter"
)

var cliVersion string

func versionCmd(args []string) error {
	if isJSONOutput() {
		return printJSON(map[string]string{
			"cli_version": cliVersion,
		})
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "CLI版本:\t%s\n", cliVersion)

	serverVer, err := cli.GetVersion()
	if err != nil {
		fmt.Fprintf(w, "服务端版本:\t无法连接 (%v)\n", err)
	} else {
		fmt.Fprintf(w, "服务端版本:\t%s\n", serverVer.Version)
		fmt.Fprintf(w, "服务端Go版本:\t%s\n", serverVer.GoVersion)
	}

	w.Flush()
	return nil
}
