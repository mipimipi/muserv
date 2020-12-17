package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/mipimipi/muserv/src/internal/config"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Verify muserv configuration",
	Long:  "Check the muserv configuration file for completeness and consistency",
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.Test(); err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
