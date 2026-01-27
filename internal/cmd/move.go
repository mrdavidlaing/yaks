package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mattwynne/yaks/internal/git"
	"github.com/mattwynne/yaks/internal/yak"
	"github.com/spf13/cobra"
)

func NewMoveCmd(store *yak.Store) *cobra.Command {
	return &cobra.Command{
		Use:   "move <old> <new...>",
		Short: "Rename or relocate a yak",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("move requires at least 2 arguments: old name and new name")
			}
			oldName := args[0]
			newName := strings.Join(args[1:], " ")
			return moveYak(store, oldName, newName)
		},
	}
}

func NewMvCmd(store *yak.Store) *cobra.Command {
	return &cobra.Command{
		Use:   "mv <old> <new...>",
		Short: "Rename or relocate a yak (alias for move)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("mv requires at least 2 arguments: old name and new name")
			}
			oldName := args[0]
			newName := strings.Join(args[1:], " ")
			return moveYak(store, oldName, newName)
		},
	}
}

func moveYak(store *yak.Store, oldName, newName string) error {
	resolvedOld, err := yak.FindYak(store, oldName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}

	if err := yak.ValidateName(newName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}

	if err := store.EnsureParents(newName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}

	if err := store.Move(resolvedOld, newName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}
	git.LogCommand(store.BasePath, "move "+resolvedOld+" "+newName)

	return nil
}
