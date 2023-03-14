package types

import (
	"context"

	"github.com/deviantony/labctl/config"
	"go.uber.org/zap"
)

const VERSION = "0.1.0"

// VPS represents a VPS instance.
type VPS struct {
	ID     int
	Name   string
	Region string
	Size   string
	Ipv4   string
}

// CommandExecutionContext holds the context and logger for a command execution.
type CommandExecutionContext struct {
	Context context.Context
	Config  config.Config
	Logger  *zap.SugaredLogger
}

// NewCommandExecutionContext creates a new command execution context.
func NewCommandExecutionContext(ctx context.Context, cfg config.Config, logger *zap.SugaredLogger) CommandExecutionContext {
	return CommandExecutionContext{
		Context: ctx,
		Config:  cfg,
		Logger:  logger,
	}
}
