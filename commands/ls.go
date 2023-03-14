package commands

import (
	"github.com/deviantony/labctl/display"
	"github.com/deviantony/labctl/do"
)

// LsCommand lists all running VPS.
type LsCommand struct{}

// Run executes the ls command.
func (cmd *LsCommand) Run(cmdCtx *CommandExecutionContext) error {
	vpsBuilder := do.NewDOVPSBuilder(cmdCtx.context, cmdCtx.config.DO, cmdCtx.logger)

	vps, err := vpsBuilder.ListVPS()
	if err != nil {
		return err
	}

	if len(vps) == 0 {
		cmdCtx.logger.Info("No VPS found")
		return nil
	}

	display.DisplayVPSList(vps)
	return nil
}
