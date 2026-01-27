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

type yakInfo struct {
	name  string
	state yak.State
	mtime int64
}

func NewListCmd(store *yak.Store) *cobra.Command {
	var format string
	var only string

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all yaks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listYaks(store, format, only)
		},
	}

	cmd.Flags().StringVar(&format, "format", "markdown", "Output format: markdown, md, plain, raw")
	cmd.Flags().StringVar(&only, "only", "", "Filter: not-done, done")

	return cmd
}

func listYaks(store *yak.Store, format, only string) error {
	yaks, err := store.List()
	if err != nil {
		return err
	}

	if len(yaks) == 0 {
		if format == "plain" || format == "raw" {
			return nil
		}
		fmt.Println("You have no yaks. Are you done?")
		return nil
	}

	var yakInfos []yakInfo
	for _, name := range yaks {
		y, err := store.Get(name)
		if err != nil {
			continue
		}

		// Filter by state if requested
		if only == "not-done" && y.State == yak.StateDone {
			continue
		}
		if only == "done" && y.State == yak.StateTodo {
			continue
		}

		// Get mtime
		yakPath := filepath.Join(store.BasePath, name)
		info, err := os.Stat(yakPath)
		if err != nil {
			continue
		}

		yakInfos = append(yakInfos, yakInfo{
			name:  name,
			state: y.State,
			mtime: info.ModTime().Unix(),
		})
	}

	// Build hierarchy and display recursively
	displayHierarchy(yakInfos, "", format)

	return nil
}

func displayHierarchy(yakInfos []yakInfo, parent string, format string) {
	// Get children of this parent
	var children []yakInfo
	for _, info := range yakInfos {
		yakParent := filepath.Dir(info.name)
		if yakParent == "." {
			yakParent = ""
		}
		if yakParent == parent {
			children = append(children, info)
		}
	}

	// Sort children: done first, then by mtime
	sort.Slice(children, func(i, j int) bool {
		if children[i].state == yak.StateDone && children[j].state != yak.StateDone {
			return true
		}
		if children[i].state != yak.StateDone && children[j].state == yak.StateDone {
			return false
		}
		return children[i].mtime < children[j].mtime
	})

	// Display children and recurse
	for _, child := range children {
		displayYak(child.name, child.state, format)
		displayHierarchy(yakInfos, child.name, format)
	}
}

func displayYak(name string, state yak.State, format string) {
	switch format {
	case "plain", "raw":
		fmt.Println(name)
	default:
		depth := strings.Count(name, "/")
		indent := strings.Repeat("  ", depth)
		displayName := filepath.Base(name)

		if state == yak.StateDone {
			fmt.Printf("\033[90m%s- [x] %s\033[0m\n", indent, displayName)
		} else {
			fmt.Printf("%s- [ ] %s\n", indent, displayName)
		}
	}
}
