package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/mattwynne/yaks/internal/git"
	"github.com/mattwynne/yaks/internal/yak"
	"github.com/spf13/cobra"
)

func NewContextCmd(store *yak.Store) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context [flags] [name...]",
		Short: "View or edit yak context",
		RunE: func(cmd *cobra.Command, args []string) error {
			showFlag, _ := cmd.Flags().GetBool("show")

			if len(args) == 0 {
				return fmt.Errorf("yak name required")
			}

			name := strings.Join(args, " ")

			// Resolve the yak name
			resolvedName, err := yak.FindYak(store, name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				return err
			}

			yakDir := os.ExpandEnv("$YAKS_PATH")
			if yakDir == "$YAKS_PATH" {
				yakDir = ".yaks"
			}
			contextPath := fmt.Sprintf("%s/%s/context.md", yakDir, resolvedName)

		if showFlag {
			return showContext(resolvedName, contextPath)
		}
		return editContext(store, resolvedName, contextPath)
		},
	}

	cmd.Flags().BoolP("edit", "e", false, "Edit context (default)")
	cmd.Flags().BoolP("show", "s", false, "Show context")

	return cmd
}

func showContext(resolvedName, contextPath string) error {
	fmt.Println(resolvedName)

	// Check if context file exists and has content
	content, err := os.ReadFile(contextPath)
	if err == nil && len(content) > 0 {
		fmt.Println()
		fmt.Print(string(content))
	}

	return nil
}

func editContext(store *yak.Store, resolvedName, contextPath string) error {
	fi, _ := os.Stdin.Stat()
	isTTY := (fi.Mode() & os.ModeCharDevice) != 0

	if isTTY {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}
		cmd := exec.Command(editor, contextPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
		git.LogCommand(store.BasePath, "context "+resolvedName)
		return nil
	}

	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
		return err
	}

	if err := os.WriteFile(contextPath, content, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing context: %v\n", err)
		return err
	}
	git.LogCommand(store.BasePath, "context "+resolvedName)

	return nil
}
