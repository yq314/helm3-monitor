package main

import (
	"github.com/yq314/helm3-monitor/cmd"
	"os"
)

func main() {
	monitorCmd := cmd.NewRootCmd(os.Stdout)

	if err := monitorCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

