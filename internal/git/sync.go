package git

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func CheckGitSetup() error {
	if !IsGitRepository() {
		return fmt.Errorf("not in a git repository")
	}

	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("no origin remote configured")
	}

	return nil
}

func GetLocalRef() string {
	sha, err := RunGit([]string{"rev-parse", "refs/notes/yaks"}, nil)
	if err != nil {
		return ""
	}
	return sha
}

func GetRemoteRef() string {
	sha, err := RunGit([]string{"rev-parse", "refs/remotes/origin/yaks"}, nil)
	if err != nil {
		return ""
	}
	return sha
}

func DetectLocalChanges(yaksPath, localRef string) bool {
	if _, err := os.Stat(yaksPath); os.IsNotExist(err) {
		return false
	}

	refDir, err := os.MkdirTemp("", "yx-sync-ref-*")
	if err != nil {
		return false
	}
	defer os.RemoveAll(refDir)

	if localRef != "" {
		extractArchive(localRef, refDir)
	}

	return !dirsEqual(yaksPath, refDir)
}

func extractArchive(ref, destDir string) {
	archiveCmd := exec.Command("git", "archive", ref)
	tarCmd := exec.Command("tar", "-x", "-C", destDir)

	pipe, err := archiveCmd.StdoutPipe()
	if err != nil {
		return
	}
	tarCmd.Stdin = pipe

	if err := archiveCmd.Start(); err != nil {
		return
	}
	if err := tarCmd.Start(); err != nil {
		archiveCmd.Wait()
		return
	}

	archiveCmd.Wait()
	tarCmd.Wait()
}

func dirsEqual(dir1, dir2 string) bool {
	cmd := exec.Command("diff", "-qr", dir1, dir2)
	cmd.Stderr = nil
	cmd.Stdout = nil
	err := cmd.Run()
	return err == nil
}

func MergeRemoteIntoLocalYaks(yaksPath, remoteRef string) error {
	tempDir, err := os.MkdirTemp("", "yx-sync-merge-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	extractArchive(remoteRef, tempDir)

	if err := copyDir(yaksPath, tempDir); err != nil {
		return err
	}

	if err := os.RemoveAll(yaksPath); err != nil {
		return err
	}
	if err := os.MkdirAll(yaksPath, 0755); err != nil {
		return err
	}
	if err := copyDir(tempDir, yaksPath); err != nil {
		return err
	}

	return nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func MergeLocalAndRemote(localRef, remoteRef string) error {
	if localRef == "" && remoteRef != "" {
		return RunGitSilent([]string{"update-ref", "refs/notes/yaks", remoteRef}, nil)
	}

	if localRef != "" && remoteRef == "" {
		return nil
	}

	if localRef != "" && remoteRef != "" && localRef != remoteRef {
		return mergeWithGitMergeTree(localRef, remoteRef)
	}

	return nil
}

func mergeWithGitMergeTree(localRef, remoteRef string) error {
	mergedTree, err := RunGit([]string{
		"merge-tree",
		"--write-tree",
		"--allow-unrelated-histories",
		localRef,
		remoteRef,
	}, nil)

	if err != nil || mergedTree == "" {
		return fmt.Errorf("git merge-tree failed unexpectedly")
	}

	return createMergeCommit(localRef, remoteRef, mergedTree)
}

func createMergeCommit(localRef, remoteRef, mergedTree string) error {
	commitSha, err := RunGit([]string{
		"commit-tree", mergedTree,
		"-p", localRef,
		"-p", remoteRef,
		"-m", "Merge yaks",
	}, nil)

	if err != nil {
		return err
	}

	return RunGitSilent([]string{"update-ref", "refs/notes/yaks", commitSha}, nil)
}

func ExtractYaksToWorkingDir(yaksPath string) error {
	if err := os.RemoveAll(yaksPath); err != nil {
		return err
	}

	if err := os.MkdirAll(yaksPath, 0755); err != nil {
		return err
	}

	if _, err := RunGit([]string{"rev-parse", "refs/notes/yaks"}, nil); err != nil {
		return nil
	}

	extractArchive("refs/notes/yaks", yaksPath)
	return nil
}

func Sync(yaksPath string) error {
	if err := CheckGitSetup(); err != nil {
		return err
	}

	RunGitSilent([]string{
		"fetch", "origin",
		"refs/notes/yaks:refs/remotes/origin/yaks",
	}, nil)

	remoteRef := GetRemoteRef()
	localRef := GetLocalRef()

	hasLocalChanges := DetectLocalChanges(yaksPath, localRef)

	if hasLocalChanges && remoteRef != "" {
		MergeRemoteIntoLocalYaks(yaksPath, remoteRef)
	}

	if hasLocalChanges {
		LogCommand(yaksPath, "sync")
		localRef = GetLocalRef()
	}

	MergeLocalAndRemote(localRef, remoteRef)

	if _, err := RunGit([]string{"rev-parse", "refs/notes/yaks"}, nil); err == nil {
		RunGitSilent([]string{
			"push", "origin",
			"refs/notes/yaks:refs/notes/yaks",
		}, nil)
	}

	ExtractYaksToWorkingDir(yaksPath)

	RunGitSilent([]string{"update-ref", "-d", "refs/remotes/origin/yaks"}, nil)

	return nil
}
