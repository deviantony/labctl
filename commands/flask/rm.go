package flask

import (
	"fmt"

	"github.com/deviantony/labctl/commands/context"
	"github.com/deviantony/labctl/terminal"
)

// RmCommand removes the given flask - matching an ID or ID prefix.
type RmCommand struct {
	ID string `arg:"" help:"Flask ID." name:"Flask ID" optional:""`
}

// Run executes the rm command.
func (cmd *RmCommand) Run(cmdCtx context.CommandExecutionContext) error {
	flaskManager, err := context.BuildManagerFromProvider(cmdCtx)
	if err != nil {
		return err
	}

	if cmd.ID == "" {
		flasks, err := flaskManager.ListFlasks()
		if err != nil {
			return err
		}

		if len(flasks) == 0 {
			cmdCtx.Logger.Info("No flask found")
			return nil
		}

		fmt.Printf("Are you sure you want to remove %d flask(s)? y/N\n", len(flasks))
		confirm, err := terminal.AskForConfirmation()
		if err != nil {
			return err
		}

		if confirm {
			for _, flask := range flasks {
				err := flaskManager.RemoveFlask(flask)
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

	return flaskManager.RemoveFlask(flask)
}
