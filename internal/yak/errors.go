package yak

import "fmt"

// ErrYakNotFound is returned when a yak cannot be found
func ErrYakNotFound(name string) error {
	return fmt.Errorf("Error: yak '%s' not found", name)
}

// ErrAmbiguousMatch is returned when multiple yaks match the search term
func ErrAmbiguousMatch(name string) error {
	return fmt.Errorf("Error: yak name '%s' is ambiguous", name)
}
