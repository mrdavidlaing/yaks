package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestIsGitRepository_InGitRepo(t *testing.T) {
	tempDir := t.TempDir()
	
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}
	
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tempDir)
	
	if !IsGitRepository() {
		t.Error("expected IsGitRepository() to return true in git repo")
	}
}

func TestIsGitRepository_NotInGitRepo(t *testing.T) {
	tempDir := t.TempDir()
	
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tempDir)
	
	if IsGitRepository() {
		t.Error("expected IsGitRepository() to return false outside git repo")
	}
}

func TestRunGit_SimpleCommand(t *testing.T) {
	tempDir := t.TempDir()
	
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}
	
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tempDir)
	
	output, err := RunGit([]string{"rev-parse", "--is-inside-work-tree"}, nil)
	if err != nil {
		t.Fatalf("RunGit failed: %v", err)
	}
	if output != "true" {
		t.Errorf("expected 'true', got '%s'", output)
	}
}

func TestRunGit_WithEnv(t *testing.T) {
	tempDir := t.TempDir()
	
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}
	
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tempDir)
	
	tempIndex := filepath.Join(tempDir, "test-index")
	env := map[string]string{
		"GIT_INDEX_FILE": tempIndex,
	}
	
	err := RunGitSilent([]string{"read-tree", "--empty"}, env)
	if err != nil {
		t.Fatalf("RunGitSilent failed: %v", err)
	}
	
	if _, err := os.Stat(tempIndex); os.IsNotExist(err) {
		t.Error("expected temp index file to be created")
	}
}

func TestLogCommand_CreatesCommit(t *testing.T) {
	tempDir := t.TempDir()
	
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}
	
	configCmd := exec.Command("git", "config", "user.email", "test@example.com")
	configCmd.Dir = tempDir
	configCmd.Run()
	
	configCmd2 := exec.Command("git", "config", "user.name", "Test User")
	configCmd2.Dir = tempDir
	configCmd2.Run()
	
	yaksPath := filepath.Join(tempDir, ".yaks")
	os.MkdirAll(yaksPath, 0755)
	os.WriteFile(filepath.Join(yaksPath, "test-yak", "state"), []byte("todo\n"), 0644)
	os.MkdirAll(filepath.Join(yaksPath, "test-yak"), 0755)
	os.WriteFile(filepath.Join(yaksPath, "test-yak", "state"), []byte("todo\n"), 0644)
	
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tempDir)
	
	LogCommand(yaksPath, "add test-yak")
	
	checkCmd := exec.Command("git", "rev-parse", "refs/notes/yaks")
	checkCmd.Dir = tempDir
	if err := checkCmd.Run(); err != nil {
		t.Error("expected refs/notes/yaks to exist after LogCommand")
	}
}

func TestLogCommand_NotInGitRepo(t *testing.T) {
	tempDir := t.TempDir()
	yaksPath := filepath.Join(tempDir, ".yaks")
	os.MkdirAll(yaksPath, 0755)
	
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tempDir)
	
	LogCommand(yaksPath, "add test-yak")
}

func TestLogCommand_YaksPathNotExist(t *testing.T) {
	tempDir := t.TempDir()
	
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}
	
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tempDir)
	
	LogCommand(filepath.Join(tempDir, "nonexistent"), "add test-yak")
}
