package yak

import (
	"testing"
)

func TestFindYak_ExactMatch(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	// Create a yak
	err := store.Create("parent/child")
	if err != nil {
		t.Fatalf("Failed to create yak: %v", err)
	}

	// Find by exact name
	result, err := FindYak(store, "parent/child")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "parent/child" {
		t.Errorf("Expected 'parent/child', got '%s'", result)
	}
}

func TestFindYak_FuzzyUniqueMatch(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	// Create multiple yaks
	store.Create("ideas/buy a pony")
	store.Create("ideas/fix the build")
	store.Create("ideas/fix the fridge")

	// Find by unique substring
	result, err := FindYak(store, "build")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "ideas/fix the build" {
		t.Errorf("Expected 'ideas/fix the build', got '%s'", result)
	}
}

func TestFindYak_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	// Create a yak
	store.Create("ideas/buy a pony")

	// Try to find non-existent yak
	_, err := FindYak(store, "nonexistent")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if err.Error() != "Error: yak 'nonexistent' not found" {
		t.Errorf("Expected error message \"Error: yak 'nonexistent' not found\", got \"%s\"", err.Error())
	}
}

func TestFindYak_AmbiguousMatch(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	// Create multiple yaks with overlapping names
	store.Create("ideas/buy a pony")
	store.Create("ideas/fix the build")
	store.Create("ideas/fix the fridge")

	// Try to find with ambiguous substring
	_, err := FindYak(store, "fix")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if err.Error() != "Error: yak name 'fix' is ambiguous" {
		t.Errorf("Expected error message \"Error: yak name 'fix' is ambiguous\", got \"%s\"", err.Error())
	}
}
