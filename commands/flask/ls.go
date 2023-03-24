package flask

import (
	"github.com/deviantony/labctl/display"
	"github.com/deviantony/labctl/lxd"
	"github.com/deviantony/labctl/types"
)

// LsCommand lists all running flasks.
type LsCommand struct{}

// Run executes the ls command.
func (cmd *LsCommand) Run(cmdCtx types.CommandExecutionContext) error {

	// flaskManager := do.NewFlaskManager(cmdCtx.Context, cmdCtx.Config.DO, cmdCtx.Logger)
	flaskManager, err := lxd.NewFlaskManager(cmdCtx.Context, cmdCtx.Config.LXD, cmdCtx.Logger)
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

	display.DisplayFlaskList(flasks)
	return nil
}
