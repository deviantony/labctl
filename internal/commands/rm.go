package commands

import (
	"fmt"
	"strings"

	"github.com/deviantony/labctl/internal/do"
)

// RmCommand removes one or more droplets.
type RmCommand struct {
	IDs []int `arg:"" help:"Droplet ID(s) to remove." name:"id" required:""`
}

// Run executes the rm command.
func (cmd *RmCommand) Run(client *do.Client, globals *Globals) error {
	var errs []string

	for _, id := range cmd.IDs {
		globals.Logger.Infow("Removing droplet", "id", id)

		if err := client.RemoveDroplet(id); err != nil {
			globals.Logger.Errorw("Failed to remove droplet", "id", id, "error", err)
			errs = append(errs, fmt.Sprintf("droplet %d: %s", id, err))
			continue
		}

		globals.Logger.Infow("Droplet removed", "id", id)
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to remove some droplets:\n  %s", strings.Join(errs, "\n  "))
	}

	return nil
}
