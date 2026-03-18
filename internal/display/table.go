package display

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/deviantony/labctl/internal/do"
	"github.com/deviantony/labctl/types"
	"github.com/jedib0t/go-pretty/v6/table"
)

// DisplayDroplets renders a list of droplets as a table.
func DisplayDroplets(droplets []types.Droplet) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"ID", "Name", "IPv4", "Region", "Size"})

	for _, d := range droplets {
		t.AppendRow(table.Row{d.ID, d.Name, d.IPv4, d.Region, d.Size})
	}

	t.Render()
}

// DisplayOptions renders a labeled table of alias-to-slug mappings.
func DisplayOptions(label string, options []do.Option) {
	fmt.Printf("%s:\n", label)
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Alias", "DigitalOcean Slug"})

	for _, o := range options {
		t.AppendRow(table.Row{o.Alias, o.Slug})
	}

	t.Render()
	fmt.Println()
}

// PrintJSON writes v as indented JSON to stdout.
func PrintJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("unable to encode JSON: %w", err)
	}
	return nil
}
