package flask

import (
	"github.com/deviantony/labctl/internal/commands/context"
	"github.com/deviantony/labctl/internal/config"
	"github.com/deviantony/labctl/internal/terminal"
)

// LsCommand lists all running flasks.
type LsCommand struct{}

// Run executes the ls command.
func (cmd *LsCommand) Run(cmdCtx context.CommandExecutionContext) error {
	flaskManager, err := context.BuildManagerFromProvider(cmdCtx)
	if err != nil {
		return err
	}

	flasks, err := flaskManager.ListFlasks()
	if err != nil {
		return err
	}

	if len(flasks) == 0 {
		cmdCtx.Logger.Info("No flask found")
		return nil
	}

	if cmdCtx.Config.GetProvider() == config.PROVIDER_DO {
		terminal.DisplayCloudFlasks(flasks)
	} else {
		terminal.DisplayLXDFlasks(flasks)
	}

	return nil
}
