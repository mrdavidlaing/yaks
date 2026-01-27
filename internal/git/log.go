package git

import (
	"os"
)

// LogCommand creates a commit in refs/notes/yaks with the yaks directory state.
// It silently returns if not in a git repository or if yaksPath doesn't exist.
// Errors are logged to stderr but do not cause the function to fail.
func LogCommand(yaksPath, message string) {
	if !IsGitRepository() {
		return
	}

	if _, err := os.Stat(yaksPath); os.IsNotExist(err) {
		return
	}

	tempIndex, err := os.CreateTemp("", "yx-git-index-*")
	if err != nil {
		return
	}
	tempIndexPath := tempIndex.Name()
	tempIndex.Close()
	defer os.Remove(tempIndexPath)

	env := map[string]string{
		"GIT_INDEX_FILE": tempIndexPath,
		"GIT_WORK_TREE":  yaksPath,
	}

	if err := RunGitSilent([]string{"read-tree", "--empty"}, env); err != nil {
		return
	}

	if err := RunGitSilent([]string{"add", "."}, env); err != nil {
		return
	}

	indexEnv := map[string]string{
		"GIT_INDEX_FILE": tempIndexPath,
	}
	treeSha, err := RunGit([]string{"write-tree"}, indexEnv)
	if err != nil {
		return
	}

	var parentArgs []string
	parentSha, err := RunGit([]string{"rev-parse", "refs/notes/yaks"}, nil)
	if err == nil && parentSha != "" {
		parentArgs = []string{"-p", parentSha}
	}

	commitArgs := []string{"commit-tree", treeSha}
	commitArgs = append(commitArgs, parentArgs...)
	commitArgs = append(commitArgs, "-m", message)

	newCommitSha, err := RunGit(commitArgs, nil)
	if err != nil {
		return
	}

	RunGitSilent([]string{"update-ref", "refs/notes/yaks", newCommitSha}, nil)
}
