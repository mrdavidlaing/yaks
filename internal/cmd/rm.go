package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mattwynne/yaks/internal/git"
	"github.com/mattwynne/yaks/internal/yak"
	"github.com/spf13/cobra"
)

func NewRmCmd(store *yak.Store) *cobra.Command {
	return &cobra.Command{
		Use:   "rm [name...]",
		Short: "Remove a yak",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("yak name required")
			}
			name := strings.Join(args, " ")

			resolvedName, err := yak.FindYak(store, name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				return err
			}

		if err := store.Delete(resolvedName); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return err
		}
		git.LogCommand(store.BasePath, "rm "+resolvedName)

		return nil
		},
	}
}
