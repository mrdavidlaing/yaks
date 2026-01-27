package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mattwynne/yaks/internal/git"
	"github.com/mattwynne/yaks/internal/yak"
	"github.com/spf13/cobra"
)

func NewAddCmd(store *yak.Store) *cobra.Command {
	return &cobra.Command{
		Use:   "add [name...]",
		Short: "Add a new yak",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return addInteractive(store)
			}
			name := strings.Join(args, " ")
			return addSingle(store, name)
		},
	}
}

func addInteractive(store *yak.Store) error {
	fmt.Println("Enter yaks (empty line to finish):")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		if err := store.Create(line); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return err
		}
		git.LogCommand(store.BasePath, "add "+line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}

	return nil
}

func addSingle(store *yak.Store, name string) error {
	if err := store.Create(name); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}
	git.LogCommand(store.BasePath, "add "+name)
	return nil
}
