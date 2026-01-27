package git

import (
	"os"
	"os/exec"
	"strings"
)

// IsGitRepository checks if the current directory is inside a git repository
func IsGitRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Stderr = nil
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "true"
}

// RunGit executes a git command with the given arguments and optional environment variables
// Returns the output and any error
func RunGit(args []string, env map[string]string) (string, error) {
	cmd := exec.Command("git", args...)
	
	if len(env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}
	
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// RunGitSilent executes a git command without capturing output
// Used for commands where we don't need the output
func RunGitSilent(args []string, env map[string]string) error {
	cmd := exec.Command("git", args...)
	cmd.Stderr = nil
	cmd.Stdout = nil
	
	if len(env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}
	
	return cmd.Run()
}
