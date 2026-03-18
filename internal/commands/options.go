package commands

import (
	"github.com/deviantony/labctl/internal/display"
	"github.com/deviantony/labctl/internal/do"
)

// OptionsCommand displays available region and size mappings.
type OptionsCommand struct{}

// Run executes the options command.
func (cmd *OptionsCommand) Run(globals *Globals) error {
	regions := do.RegionOptions()
	sizes := do.SizeOptions()

	if globals.JSON {
		return display.PrintJSON(struct {
			Regions []do.Option `json:"regions"`
			Sizes   []do.Option `json:"sizes"`
		}{
			Regions: regions,
			Sizes:   sizes,
		})
	}

	display.DisplayOptions("Regions", regions)
	display.DisplayOptions("Sizes", sizes)
	return nil
}
