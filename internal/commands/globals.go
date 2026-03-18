package commands

import "go.uber.org/zap"

// Globals holds shared state passed to all commands.
type Globals struct {
	JSON   bool
	Logger *zap.SugaredLogger
}
