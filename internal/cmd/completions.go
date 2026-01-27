package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mattwynne/yaks/internal/yak"
	"github.com/spf13/cobra"
)

func NewCompletionsCmd(store *yak.Store) *cobra.Command {
	return &cobra.Command{
		Use:                "completions [command] [flags]",
		Short:              "Generate completion suggestions or install completions",
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if this is an install command
			if len(args) > 0 && args[0] == "install" {
				return installCompletions(args[1:])
			}

			// Otherwise, list yaks for completion (existing logic)
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

func installCompletions(args []string) error {
	dryRun := false
	for _, arg := range args {
		if arg == "--dry-run" {
			dryRun = true
		}
	}

	shell := os.Getenv("SHELL")
	rcFile, err := getShellRcFile(shell)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not detect shell from $SHELL=%s\n", shell)
		fmt.Fprintf(os.Stderr, "Please manually source the completion script for your shell\n")
		return err
	}

	shellName := filepath.Base(shell)
	completionFile := fmt.Sprintf("completions/yx.%s", shellName)
	completionLine := fmt.Sprintf("source %s", completionFile)

	if dryRun {
		fmt.Printf("Would add to %s:\n", rcFile)
		fmt.Printf("  %s\n", completionLine)
		return nil
	}

	// Check if file is writable
	if _, err := os.Stat(rcFile); err == nil {
		// File exists, check if writable
		f, err := os.OpenFile(rcFile, os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to write to %s\n", rcFile)
			fmt.Fprintf(os.Stderr, "Please add the following line manually to %s:\n", rcFile)
			fmt.Fprintf(os.Stderr, "  %s\n", completionLine)
			return err
		}
		defer f.Close()

		// Check if already installed
		content, _ := os.ReadFile(rcFile)
		if strings.Contains(string(content), completionLine) {
			fmt.Printf("Completion already installed in %s\n", rcFile)
			return nil
		}

		// Add completion
		fmt.Fprintln(f, "")
		fmt.Fprintln(f, "# yx completion")
		fmt.Fprintln(f, completionLine)

		fmt.Printf("Completion installed to %s\n", rcFile)
		fmt.Printf("Run 'source %s' or restart your shell to enable completions\n", rcFile)
	} else {
		fmt.Fprintf(os.Stderr, "Error: Failed to write to %s\n", rcFile)
		fmt.Fprintf(os.Stderr, "Please add the following line manually to %s:\n", rcFile)
		fmt.Fprintf(os.Stderr, "  %s\n", completionLine)
		return err
	}

	return nil
}

func getShellRcFile(shell string) (string, error) {
	shellName := filepath.Base(shell)
	home := os.Getenv("HOME")

	switch shellName {
	case "bash":
		return filepath.Join(home, ".bashrc"), nil
	case "zsh":
		return filepath.Join(home, ".zshrc"), nil
	default:
		return "", fmt.Errorf("unsupported shell")
	}
}
