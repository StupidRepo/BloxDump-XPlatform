package utils

import (
	"fmt"
)

var (
	EmptyDirectory   = fmt.Errorf("a directory path was blank")
	InvalidDirectory = func(msg string) error { return fmt.Errorf("invalid directory: %s", msg) }
)
