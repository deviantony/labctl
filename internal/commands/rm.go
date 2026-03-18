package commands

import (
	"fmt"
	"strings"
	"sync"

	"github.com/deviantony/labctl/internal/do"
)

// RmCommand removes one or more droplets.
type RmCommand struct {
	IDs []int `arg:"" help:"Droplet ID(s) to remove." name:"id" required:""`
}

// Run executes the rm command.
func (cmd *RmCommand) Run(client *do.Client, globals *Globals) error {
	var (
		mu   sync.Mutex
		errs []string
		wg   sync.WaitGroup
	)

	for _, id := range cmd.IDs {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			globals.Logger.Infow("Removing droplet", "id", id)

			if err := client.RemoveDroplet(id); err != nil {
				globals.Logger.Errorw("Failed to remove droplet", "id", id, "error", err)
				mu.Lock()
				errs = append(errs, fmt.Sprintf("droplet %d: %v", id, err))
				mu.Unlock()
				return
			}

			globals.Logger.Infow("Droplet removed", "id", id)
		}(id)
	}

	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("failed to remove droplets:\n  %s", strings.Join(errs, "\n  "))
	}

	return nil
}
