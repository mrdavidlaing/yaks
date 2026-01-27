package yak

import (
	"errors"
	"strings"
)

var ErrInvalidName = errors.New("Invalid yak name: contains forbidden characters (\\ : * ? | < > \")")

// ValidateName checks if a yak name contains forbidden characters.
func ValidateName(name string) error {
	if strings.ContainsAny(name, "\\:*?|<>\"") {
		return ErrInvalidName
	}
	return nil
}
