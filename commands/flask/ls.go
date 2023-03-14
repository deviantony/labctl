package flask

import (
	"github.com/deviantony/labctl/display"
	"github.com/deviantony/labctl/do"
	"github.com/deviantony/labctl/types"
)

// LsCommand lists all running VPS.
type LsCommand struct{}

// Run executes the ls command.
func (cmd *LsCommand) Run(cmdCtx types.CommandExecutionContext) error {
	vpsBuilder := do.NewDOVPSBuilder(cmdCtx.Context, cmdCtx.Config.DO, cmdCtx.Logger)

	vps, err := vpsBuilder.ListVPS()
	if err != nil {
		return err
	}

	if len(vps) == 0 {
		cmdCtx.Logger.Info("No flask found")
		return nil
	}

	display.DisplayVPSList(vps)
	return nil
}
