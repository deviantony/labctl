package types

import (
	"context"

	"github.com/deviantony/labctl/config"
	"go.uber.org/zap"
)

const VERSION = "0.2.0-dev"

// A flask is an environment that can run in LXC or in the cloud
type Flask struct {
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
