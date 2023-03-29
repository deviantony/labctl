package types

import (
	"context"

	"github.com/deviantony/labctl/config"
	"go.uber.org/zap"
)

const VERSION = "0.2.0-dev"

// CommandExecutionContext holds the context and logger for a command execution.
type CommandExecutionContext struct {
	Context context.Context
	Config  config.Config
	Logger  *zap.SugaredLogger
}

// A flask is an environment that can run in LXC or in the cloud
type Flask struct {
	Name string
	Ipv4 string
	LXD  FlaskLXDProperties
	DO   FlaskDOProperties
}

// FlaskDOProperties holds the DigitalOcean specific properties for a flask
type FlaskDOProperties struct {
	ID     int
	Region string
	Size   string
}

// FlaskLXDProperties holds the LXD specific properties for a flask
type FlaskLXDProperties struct {
	ID       string
	Status   string
	Profiles []string
}

// NewCommandExecutionContext creates a new command execution context.
func NewCommandExecutionContext(ctx context.Context, cfg config.Config, logger *zap.SugaredLogger) CommandExecutionContext {
	return CommandExecutionContext{
		Context: ctx,
		Config:  cfg,
		Logger:  logger,
	}
}
