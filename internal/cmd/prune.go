package cmd

import (
	"fmt"

	"github.com/mattwynne/yaks/internal/git"
	"github.com/mattwynne/yaks/internal/yak"
	"github.com/spf13/cobra"
)

func NewPruneCmd(store *yak.Store) *cobra.Command {
	return &cobra.Command{
		Use:   "prune",
		Short: "Remove all done yaks",
		RunE: func(cmd *cobra.Command, args []string) error {
			yakNames, err := store.List()
			if err != nil {
				return err
			}

			for _, name := range yakNames {
				y, err := store.Get(name)
				if err != nil {
					continue
				}
			if y.State == yak.StateDone {
				if err := store.Delete(name); err != nil {
					fmt.Fprintf(cmd.OutOrStderr(), "Error deleting %s: %v\n", name, err)
				} else {
					git.LogCommand(store.BasePath, "rm "+name)
				}
			}
			}

			return nil
		},
	}
}
