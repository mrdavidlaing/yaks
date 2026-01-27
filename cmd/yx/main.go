package main

import (
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "yx",
	Short: "A non-linear TODO list for humans and robots",
	Long: `yx is a CLI tool for managing Yak Maps - a TODO list of nested goals.
A Yak Map is basically the same as a Mikado Graph or a Discovery Tree.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
