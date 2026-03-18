package commands

import (
	"github.com/deviantony/labctl/internal/display"
	"github.com/deviantony/labctl/internal/do"
)

// LsCommand lists all droplets.
type LsCommand struct{}

// Run executes the ls command.
func (cmd *LsCommand) Run(client *do.Client, globals *Globals) error {
	droplets, err := client.ListDroplets()
	if err != nil {
		return err
	}

	if globals.JSON {
		return display.PrintJSON(droplets)
	}

	if len(droplets) == 0 {
		globals.Logger.Info("No droplets found")
		return nil
	}

	display.DisplayDroplets(droplets)
	return nil
}
