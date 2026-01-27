package main

import (
	"os"

	"github.com/mattwynne/yaks/internal/cmd"
	"github.com/mattwynne/yaks/internal/yak"
	"github.com/spf13/cobra"
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
	yaksPath := os.Getenv("YAKS_PATH")
	if yaksPath == "" {
		yaksPath = ".yaks"
	}

	store := yak.NewStore(yaksPath)
	store.Migrate()

	rootCmd.AddCommand(cmd.NewAddCmd(store))
	rootCmd.AddCommand(cmd.NewRmCmd(store))
	rootCmd.AddCommand(cmd.NewContextCmd(store))

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
