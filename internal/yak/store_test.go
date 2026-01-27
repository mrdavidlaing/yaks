package yak

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewStore_ConvertsRelativePathToAbsolute(t *testing.T) {
	store := NewStore(".yaks")
	cwd, _ := os.Getwd()
	expected := filepath.Join(cwd, ".yaks")
	if store.BasePath != expected {
		t.Errorf("expected '%s', got '%s'", expected, store.BasePath)
	}
}

func TestNewStore_KeepsAbsolutePathUnchanged(t *testing.T) {
	store := NewStore("/tmp/yaks")
	if store.BasePath != "/tmp/yaks" {
		t.Errorf("expected '/tmp/yaks', got '%s'", store.BasePath)
	}
}

func TestCreate_CreatesDirectoryAndStateFile(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	err := store.Create("my task")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	yakDir := filepath.Join(tmpDir, "my task")
	if _, err := os.Stat(yakDir); os.IsNotExist(err) {
		t.Error("yak directory was not created")
	}

	stateFile := filepath.Join(yakDir, "state")
	content, err := os.ReadFile(stateFile)
	if err != nil {
		t.Fatalf("state file not found: %v", err)
	}
	if string(content) != "todo\n" {
		t.Errorf("expected state 'todo\\n', got '%s'", string(content))
	}
}

func TestCreate_CreatesContextMdFile(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	err := store.Create("my task")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	contextFile := filepath.Join(tmpDir, "my task", "context.md")
	if _, err := os.Stat(contextFile); os.IsNotExist(err) {
		t.Error("context.md file was not created")
	}
}

func TestCreate_ValidatesName(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	err := store.Create("foo:bar")
	if err == nil {
		t.Error("expected error for invalid name, got nil")
	}
}

func TestCreate_NestedYak(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	err := store.Create("parent/child")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	yakDir := filepath.Join(tmpDir, "parent", "child")
	if _, err := os.Stat(yakDir); os.IsNotExist(err) {
		t.Error("nested yak directory was not created")
	}

	stateFile := filepath.Join(yakDir, "state")
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		t.Error("state file was not created for nested yak")
	}
}

func TestExists_ReturnsTrueForExistingYak(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)
	store.Create("my task")

	if !store.Exists("my task") {
		t.Error("expected Exists to return true for existing yak")
	}
}

func TestExists_ReturnsFalseForNonExistingYak(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	if store.Exists("nonexistent") {
		t.Error("expected Exists to return false for non-existing yak")
	}
}

func TestGet_LoadsYakFromDisk(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)
	store.Create("my task")

	yak, err := store.Get("my task")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if yak.Name != "my task" {
		t.Errorf("expected name 'my task', got '%s'", yak.Name)
	}
	if yak.State != StateTodo {
		t.Errorf("expected state todo, got '%s'", yak.State)
	}
}

func TestGet_LoadsDoneState(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)
	store.Create("my task")
	os.WriteFile(filepath.Join(tmpDir, "my task", "state"), []byte("done\n"), 0644)

	yak, err := store.Get("my task")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if yak.State != StateDone {
		t.Errorf("expected state done, got '%s'", yak.State)
	}
}

func TestList_ReturnsAllYakNames(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)
	store.Create("task1")
	store.Create("task2")

	names, err := store.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 2 {
		t.Errorf("expected 2 yaks, got %d", len(names))
	}
}

func TestList_IncludesNestedYaks(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)
	store.Create("parent")
	store.Create("parent/child")

	names, err := store.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 2 {
		t.Errorf("expected 2 yaks (parent and parent/child), got %d: %v", len(names), names)
	}
}

func TestList_EmptyWhenNoYaks(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	names, err := store.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 0 {
		t.Errorf("expected 0 yaks, got %d", len(names))
	}
}

func TestDelete_RemovesYakDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)
	store.Create("my task")

	err := store.Delete("my task")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if store.Exists("my task") {
		t.Error("yak should not exist after deletion")
	}
}

func TestMigrate_ConvertsDoneFileToStateFile(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	yakDir := filepath.Join(tmpDir, "old yak")
	os.MkdirAll(yakDir, 0755)
	os.WriteFile(filepath.Join(yakDir, "done"), []byte{}, 0644)

	err := store.Migrate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stateFile := filepath.Join(yakDir, "state")
	content, err := os.ReadFile(stateFile)
	if err != nil {
		t.Fatalf("state file not found: %v", err)
	}
	if string(content) != "done\n" {
		t.Errorf("expected state 'done\\n', got '%s'", string(content))
	}

	doneFile := filepath.Join(yakDir, "done")
	if _, err := os.Stat(doneFile); !os.IsNotExist(err) {
		t.Error("done file should have been removed")
	}
}
