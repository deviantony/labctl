package terminal

import (
	"os"
	"strings"
	"time"

	"github.com/deviantony/labctl/dockerhub"
	"github.com/deviantony/labctl/types"
	"github.com/jedib0t/go-pretty/v6/table"
)

// DisplayDockerHubAccessTokens displays a list of DockerHub access tokens in a table format on the standard output.
func DisplayDockerHubAccessTokens(tokens []dockerhub.AccessToken) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"ID", "Label", "Created", "Last used"})

	for _, token := range tokens {
		t.AppendRow(table.Row{token.Uuid, token.Label, token.CreatedAt.Format(time.DateOnly), token.LastUsed.Format(time.DateOnly)})
	}

	t.Render()
}

// DisplayCloudFlasks displays a list of cloud based flasks in a table format on the standard output.
func DisplayCloudFlasks(flasks []types.Flask) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"ID", "Name", "IPv4", "Region", "Size"})

	for _, flask := range flasks {
		t.AppendRow(table.Row{flask.DO.ID, flask.Name, flask.Ipv4, flask.DO.Region, flask.DO.Size})
	}

	t.Render()
}

// DisplayLXDFlasks displays a list of LXD based flasks in a table format on the standard output.
func DisplayLXDFlasks(flasks []types.Flask) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"ID", "Name", "Status", "IPv4", "Profiles"})

	for _, flask := range flasks {
		t.AppendRow(table.Row{flask.LXD.ID, flask.Name, flask.LXD.Status, flask.Ipv4, strings.Join(flask.LXD.Profiles, ",")})
	}

	t.Render()
}

// DisplayOptionList displays a list of CLI options and their DO equivalent.
func DisplayOptionList() {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"CLI", "CLOUD", "LXD"})

	t.AppendRow(table.Row{"REGION"})
	t.AppendSeparator()
	t.AppendRow(table.Row{"usw", "sfo1", "-"})
	t.AppendRow(table.Row{"use", "nyc1", "-"})
	t.AppendRow(table.Row{"eu", "fra1", "-"})
	t.AppendRow(table.Row{"ap", "sgp1", "-"})
	t.AppendRow(table.Row{"nz", "syd1", "-"})

	t.AppendSeparator()
	t.AppendRow(table.Row{"SIZE", "https://slugs.do-api.dev/", ""})
	t.AppendSeparator()

	t.AppendRow(table.Row{"xs", "s-1vcpu-512mb-10gb", "1vcpu-512mb-10gb"})
	t.AppendRow(table.Row{"s", "s-1vcpu-1gb", "1vcpu-1gb-10gb"})
	t.AppendRow(table.Row{"m", "s-2vcpu-4gb", "2vcpu-2gb-20gb"})
	t.AppendRow(table.Row{"l", "s-4vcpu-8gb", "4vcpu-4g-50gb"})
	t.AppendRow(table.Row{"xl", "s-8vcpu-16gb", "-"})

	t.Render()
}
