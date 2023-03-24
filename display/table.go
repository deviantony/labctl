package display

import (
	"os"

	"github.com/deviantony/labctl/types"
	"github.com/jedib0t/go-pretty/v6/table"
)

// DisplayFlaskList displays a list of flasks in a table format on the standard output.
func DisplayFlaskList(flasks []types.Flask) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"ID", "Name", "IPv4", "Region", "Size"})

	for _, flask := range flasks {
		t.AppendRow(table.Row{flask.ID, flask.Name, flask.Ipv4, flask.Config.Region, flask.Config.Size})
	}

	t.Render()
}

func DisplayOptionList() {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"CLI OPTION", "DO EQUIVALENT"})

	t.AppendRow(table.Row{"REGIONS"})
	t.AppendSeparator()
	t.AppendRow(table.Row{"usw", "sfo1"})
	t.AppendRow(table.Row{"use", "nyc1"})
	t.AppendRow(table.Row{"eu", "fra1"})
	t.AppendRow(table.Row{"ap", "sgp1"})
	t.AppendRow(table.Row{"nz", "syd1"})

	t.AppendSeparator()
	t.AppendRow(table.Row{"SIZES (https://slugs.do-api.dev/)"})
	t.AppendSeparator()

	t.AppendRow(table.Row{"xs", "s-1vcpu-512mb-10gb"})
	t.AppendRow(table.Row{"s", "s-1vcpu-1gb"})
	t.AppendRow(table.Row{"m", "s-2vcpu-4gb"})
	t.AppendRow(table.Row{"l", "s-4vcpu-8gb"})
	t.AppendRow(table.Row{"xl", "s-8vcpu-16gb"})

	t.Render()
}
