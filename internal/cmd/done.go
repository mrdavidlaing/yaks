package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mattwynne/yaks/internal/yak"
	"github.com/spf13/cobra"
)

func NewDoneCmd(store *yak.Store) *cobra.Command {
	var undo bool
	var recursive bool

	cmd := &cobra.Command{
		Use:   "done [name...]",
		Short: "Mark a yak as done",
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

			if undo {
				if err := store.SetState(resolvedName, yak.StateTodo); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					return err
				}
				return nil
			}

			if recursive {
				if err := store.MarkDoneRecursively(resolvedName); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					return err
				}
				return nil
			}

			hasIncomplete, err := store.HasIncompleteChildren(resolvedName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				return err
			}
			if hasIncomplete {
				fmt.Fprintf(os.Stderr, "Error: cannot mark '%s' as done - it has incomplete children\n", resolvedName)
				return fmt.Errorf("cannot mark '%s' as done - it has incomplete children", resolvedName)
			}

			if err := store.SetState(resolvedName, yak.StateDone); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				return err
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&undo, "undo", false, "Mark yak as todo instead of done")
	cmd.Flags().BoolVar(&recursive, "recursive", false, "Mark yak and all children as done")

	return cmd
}
