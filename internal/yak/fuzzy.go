package yak

import (
	"strings"
)

func FindYak(store *Store, searchTerm string) (string, error) {
	if store.Exists(searchTerm) {
		return searchTerm, nil
	}

	yakNames, err := store.List()
	if err != nil {
		return "", err
	}

	var matches []string
	for _, yakName := range yakNames {
		if strings.Contains(yakName, searchTerm) {
			matches = append(matches, yakName)
		}
	}

	if len(matches) == 0 {
		return "", ErrYakNotFound(searchTerm)
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	return "", ErrAmbiguousMatch(searchTerm)
}
