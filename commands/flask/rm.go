package flask

import (
	"fmt"

	"github.com/deviantony/labctl/display"
	"github.com/deviantony/labctl/do"
	"github.com/deviantony/labctl/types"
)

// RmCommand removes the given flask - matching an ID or ID prefix.
type RmCommand struct {
	ID int `arg:"" help:"VPS ID." name:"VPS ID" optional:""`
}

// Run executes the rm command.
func (cmd *RmCommand) Run(cmdCtx types.CommandExecutionContext) error {
	vpsBuilder := do.NewDOVPSBuilder(cmdCtx.Context, cmdCtx.Config.DO, cmdCtx.Logger)

	if cmd.ID == 0 {
		vps, err := vpsBuilder.ListVPS()
		if err != nil {
			return err
		}

		if len(vps) == 0 {
			cmdCtx.Logger.Info("No flask found")
			return nil
		}

		fmt.Printf("Are you sure you want to remove %d flask(s)? y/N\n", len(vps))
		confirm, err := display.AskForConfirmation()
		if err != nil {
			return err
		}

		if confirm {
			for _, v := range vps {
				err := vpsBuilder.RemoveVPS(v.ID)
				if err != nil {
					return err
				}
			}
		}

		return nil
	}

	vps, err := vpsBuilder.GetVPS(cmd.ID)
	if err != nil {
		return err
	}

	return vpsBuilder.RemoveVPS(vps.ID)
}
