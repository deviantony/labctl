package context

import (
	"context"
	"fmt"

	"github.com/deviantony/labctl/config"
	"github.com/deviantony/labctl/do"
	"github.com/deviantony/labctl/lxd"
	"github.com/deviantony/labctl/types"
	"go.uber.org/zap"
)

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

// BuildManagerFromProvider returns a FlaskManager based on the specified provider.
func BuildManagerFromProvider(cmdCtx CommandExecutionContext) (types.FlaskManager, error) {
	switch cmdCtx.Config.GetProvider() {
	case config.PROVIDER_LXD:
		return lxd.NewFlaskManager(cmdCtx.Context, cmdCtx.Config.LXD, cmdCtx.Logger)
	case config.PROVIDER_DO:
		return do.NewFlaskManager(cmdCtx.Context, cmdCtx.Config.DO, cmdCtx.Logger)
	default:
		return nil, fmt.Errorf("unknown provider: %s", cmdCtx.Config.GetProvider())
	}
}
