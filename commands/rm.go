package commands

import (
	"fmt"

	"github.com/deviantony/labctl/do"
)

// RmCommand removes the given VPS matching an ID or ID prefix.
type RmCommand struct {
	ID int `arg:"" help:"VPS ID." name:"VPS ID" optional:""`
}

// Run executes the rm command.
func (cmd *RmCommand) Run(cmdCtx *CommandExecutionContext) error {
	vpsBuilder := do.NewDOVPSBuilder(cmdCtx.context, cmdCtx.config.DO, cmdCtx.logger)

	if cmd.ID == 0 {
		vps, err := vpsBuilder.ListVPS()
		if err != nil {
			return err
		}

		if len(vps) == 0 {
			cmdCtx.logger.Info("No VPS found")
			return nil
		}

		fmt.Printf("Are you sure you want to remove %d VPS? y/N\n", len(vps))
		confirm, err := askForConfirmation()
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
