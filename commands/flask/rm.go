package flask

import (
	"fmt"

	"github.com/deviantony/labctl/display"
	"github.com/deviantony/labctl/do"
	"github.com/deviantony/labctl/types"
)

// RmCommand removes the given flask - matching an ID or ID prefix.
type RmCommand struct {
	ID int `arg:"" help:"Flask ID." name:"Flask ID" optional:""`
}

// Run executes the rm command.
func (cmd *RmCommand) Run(cmdCtx types.CommandExecutionContext) error {
	flaskManager := do.NewFlaskManager(cmdCtx.Context, cmdCtx.Config.DO, cmdCtx.Logger)

	if cmd.ID == 0 {
		flasks, err := flaskManager.ListFlasks()
		if err != nil {
			return err
		}

		if len(flasks) == 0 {
			cmdCtx.Logger.Info("No flask found")
			return nil
		}

		fmt.Printf("Are you sure you want to remove %d flask(s)? y/N\n", len(flasks))
		confirm, err := display.AskForConfirmation()
		if err != nil {
			return err
		}

		if confirm {
			for _, v := range flasks {
				err := flaskManager.RemoveFlask(v.ID)
				if err != nil {
					return err
				}
			}
		}

		return nil
	}

	flask, err := flaskManager.GetFlask(cmd.ID)
	if err != nil {
		return err
	}

	return flaskManager.RemoveFlask(flask.ID)
}
