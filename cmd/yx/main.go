package main

import (
	"os"
	"strings"

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
	SilenceErrors: true,
	SilenceUsage:  true,
}

func main() {
	yaksPath := os.Getenv("YAKS_PATH")
	if yaksPath == "" {
		yaksPath = ".yaks"
	}

	store := yak.NewStore(yaksPath)
	store.Migrate()

	rootCmd.AddCommand(cmd.NewAddCmd(store))
	rootCmd.AddCommand(cmd.NewListCmd(store))
	rootCmd.AddCommand(cmd.NewRmCmd(store))
	rootCmd.AddCommand(cmd.NewContextCmd(store))
	rootCmd.AddCommand(cmd.NewDoneCmd(store))
	rootCmd.AddCommand(cmd.NewMoveCmd(store))
	rootCmd.AddCommand(cmd.NewMvCmd(store))
	rootCmd.AddCommand(cmd.NewPruneCmd(store))
	rootCmd.AddCommand(cmd.NewSyncCmd(store))
	rootCmd.AddCommand(cmd.NewCompletionsCmd(store))

	if err := rootCmd.Execute(); err != nil {
		// Check if it's an unknown command error
		if strings.Contains(err.Error(), "unknown command") {
			rootCmd.Help()
			os.Exit(0)
		}
		os.Exit(1)
	}
}
