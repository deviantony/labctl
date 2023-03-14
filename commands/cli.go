package commands

import (
	"context"
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/deviantony/labctl/config"
	"go.uber.org/zap"
)

// VersionFlag is used to display the version of the CLI.
type VersionFlag string

func (v VersionFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                         { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	fmt.Println(vars["version"])
	app.Exit(0)
	return nil
}

var CLI struct {
	// Generic options
	Debug   bool        `help:"Enable debug mode."`
	Version VersionFlag `name:"version" help:"Print version information and quit"`

	// Commands
	Create CreateCommand `cmd:"" help:"Create a new VPS." default:"withargs"`
	Cp     CpCommand     `cmd:"" help:"Copy a file or a directory to a remote VPS."`
	Exec   ExecCommand   `cmd:"" help:"Create a SSH session to the given VPS ID."`
	Ls     LsCommand     `cmd:"" help:"List all running VPS."`
	Rm     RmCommand     `cmd:"" help:"Remove a VPS."`
}

// CommandExecutionContext holds the context and logger for a command execution.
type CommandExecutionContext struct {
	context context.Context
	config  config.Config
	logger  *zap.SugaredLogger
}

// NewCommandExecutionContext creates a new command execution context.
func NewCommandExecutionContext(ctx context.Context, cfg config.Config, logger *zap.SugaredLogger) *CommandExecutionContext {
	return &CommandExecutionContext{
		context: ctx,
		config:  cfg,
		logger:  logger,
	}
}
