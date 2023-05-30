package filesystem

import (
	"errors"
	"fmt"
	"os"
)

// FileExists checks if a file exists on the filesystem
func FileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, fmt.Errorf("an error occurred while checking if the file exists: %w", err)
	}
}
