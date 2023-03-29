package filesystem

import (
	"errors"
	"os"

	"go.uber.org/zap"
)

// FileExists checks if a file exists on the filesystem
func FileExists(path string, logger *zap.SugaredLogger) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		logger.Errorf("An error occurred while checking if the file exists: %s", err)
		return true
	}
}
