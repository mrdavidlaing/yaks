package cmd

import (
	"fmt"
	"sort"

	"github.com/mattwynne/yaks/internal/yak"
	"github.com/spf13/cobra"
)

func NewCompletionsCmd(store *yak.Store) *cobra.Command {
	return &cobra.Command{
		Use:                "completions [command] [flags]",
		Short:              "Generate completion suggestions",
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			command := ""
			if len(args) > 0 {
				command = args[0]
			}

			flag := ""
			if len(args) > 1 {
				flag = args[1]
			}

			yaks, err := store.List()
			if err != nil {
				return err
			}

			sort.Strings(yaks)

			for _, name := range yaks {
				if shouldInclude(store, name, command, flag) {
					fmt.Println(name)
				}
			}

			return nil
		},
	}
}

func shouldInclude(store *yak.Store, name, command, flag string) bool {
	if command == "done" {
		y, err := store.Get(name)
		if err != nil {
			return false
		}

		if flag == "--undo" {
			return y.State == yak.StateDone
		} else {
			return y.State == yak.StateTodo
		}
	}

	return true
}
