package commands

import (
	"fmt"

	"github.com/deviantony/labctl/internal/display"
	"github.com/deviantony/labctl/internal/do"
	"github.com/deviantony/labctl/types"
)

// StatusCommand shows version and API connectivity.
type StatusCommand struct{}

// Run executes the status command.
func (cmd *StatusCommand) Run(client *do.Client, globals *Globals) error {
	apiStatus := "ok"
	if err := client.CheckAPI(); err != nil {
		apiStatus = fmt.Sprintf("error — %s", err)
	}

	if globals.JSON {
		return display.PrintJSON(struct {
			Version string `json:"version"`
			API     string `json:"api"`
		}{
			Version: types.VERSION,
			API:     apiStatus,
		})
	}

	fmt.Printf("Version:  %s\n", types.VERSION)
	fmt.Printf("API:      %s\n", apiStatus)
	return nil
}
