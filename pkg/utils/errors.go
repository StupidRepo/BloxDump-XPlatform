package utils

import (
	"fmt"
)

var (
	EmptyDirectory   = fmt.Errorf("a directory path was blank")
	InvalidDirectory = func(msg string) error { return fmt.Errorf("invalid directory: %s", msg) }
)

func VersionMismatch(got, expected int) error {
	return fmt.Errorf("configuration version mismatch. got %d, expected %d", got, expected)
}
