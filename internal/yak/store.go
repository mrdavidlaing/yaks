package yak

import (
	"os"
	"path/filepath"
	"strings"
)

type Store struct {
	BasePath string
}

func NewStore(path string) *Store {
	absPath := path
	if !filepath.IsAbs(path) {
		cwd, _ := os.Getwd()
		absPath = filepath.Join(cwd, path)
	}
	return &Store{BasePath: absPath}
}

func (s *Store) Create(name string) error {
	if err := ValidateName(name); err != nil {
		return err
	}
	yakDir := filepath.Join(s.BasePath, name)
	if err := os.MkdirAll(yakDir, 0755); err != nil {
		return err
	}
	stateFile := filepath.Join(yakDir, "state")
	if err := os.WriteFile(stateFile, []byte("todo\n"), 0644); err != nil {
		return err
	}
	contextFile := filepath.Join(yakDir, "context.md")
	f, err := os.Create(contextFile)
	if err != nil {
		return err
	}
	f.Close()
	return nil
}

func (s *Store) Exists(name string) bool {
	yakDir := filepath.Join(s.BasePath, name)
	_, err := os.Stat(yakDir)
	return err == nil
}

func (s *Store) Get(name string) (*Yak, error) {
	yakDir := filepath.Join(s.BasePath, name)
	stateFile := filepath.Join(yakDir, "state")
	content, err := os.ReadFile(stateFile)
	if err != nil {
		return nil, err
	}
	stateStr := strings.TrimSpace(string(content))
	return &Yak{
		Name:        name,
		State:       State(stateStr),
		ContextPath: filepath.Join(yakDir, "context.md"),
	}, nil
}

func (s *Store) List() ([]string, error) {
	if _, err := os.Stat(s.BasePath); os.IsNotExist(err) {
		return []string{}, nil
	}
	var names []string
	err := filepath.WalkDir(s.BasePath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && path != s.BasePath {
			relPath, _ := filepath.Rel(s.BasePath, path)
			names = append(names, relPath)
		}
		return nil
	})
	return names, err
}

func (s *Store) Delete(name string) error {
	yakDir := filepath.Join(s.BasePath, name)
	return os.RemoveAll(yakDir)
}

func (s *Store) Migrate() error {
	if _, err := os.Stat(s.BasePath); os.IsNotExist(err) {
		return nil
	}
	return filepath.WalkDir(s.BasePath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == "done" {
			yakDir := filepath.Dir(path)
			stateFile := filepath.Join(yakDir, "state")
			if err := os.WriteFile(stateFile, []byte("done\n"), 0644); err != nil {
				return err
			}
			if err := os.Remove(path); err != nil {
				return err
			}
		}
		return nil
	})
}
