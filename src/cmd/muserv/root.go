package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var preamble = `muserv ` + Version + `
Copyright (C) 2020 Michael Picht <https://gitlab.com/mipimipi/muserv>

muserv is a UPnP and DLNA compatible music server.

Web site: https://gitlab.com/mipimipi/muserv/

muserv comes with ABSOLUTELY NO WARRANTY. This is free software, and you
are welcome to redistribute it under certain conditions.  See the GNU
General Public Licence for details.`

var rootCmd = &cobra.Command{
	Use:     "muserv",
	Short:   "muserv music server",
	Long:    preamble,
	Version: Version,
}

func execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
