package cmd

import (
	"fmt"
	"os"

	"github.com/mattwynne/yaks/internal/git"
	"github.com/mattwynne/yaks/internal/yak"
	"github.com/spf13/cobra"
)

func NewSyncCmd(store *yak.Store) *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Synchronize yaks with remote repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := git.Sync(store.BasePath); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				return err
			}
			return nil
		},
	}
}
