package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/mipimipi/muserv/src/internal/server"
)

// runCmd represents the start command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run muserv service",
	Long:  "Run the muserv service",
	Run: func(cmd *cobra.Command, args []string) {
		if err := server.Run(Version); err != nil {
			fmt.Printf("muserv cannot be run: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
