package commands

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/deviantony/labctl/internal/do"
	"github.com/deviantony/labctl/internal/display"
	"github.com/deviantony/labctl/pkg/random"
)

// CreateCommand creates a new droplet.
type CreateCommand struct {
	Region string `help:"Region of the droplet." short:"r" enum:"usw,use,eu,ap,au" default:"eu"`
	Size   string `help:"Size of the droplet." short:"s" enum:"xs,s,m,l,xl" default:"xs"`
	Name   string `help:"Name of the droplet." short:"n"`
}

// Run executes the create command.
func (cmd *CreateCommand) Run(client *do.Client, globals *Globals) error {
	name := cmd.Name
	if name == "" {
		name = random.GeneratePetName(2, "-")
	}

	globals.Logger.Infow("Creating droplet", "name", name, "region", cmd.Region, "size", cmd.Size)

	droplet, actionHREF, err := client.CreateDroplet(name, cmd.Region, cmd.Size)
	if err != nil {
		return err
	}

	err = client.WaitUntilReady(&droplet, actionHREF)
	if err != nil {
		return err
	}

	globals.Logger.Infow("Droplet ready", "id", droplet.ID, "ip", droplet.IPv4)

	sshCommand := fmt.Sprintf("ssh -o StrictHostKeyChecking=no root@%s", droplet.IPv4)

	if err := clipboard.WriteAll(sshCommand); err != nil {
		globals.Logger.Warnf("Unable to copy SSH command to clipboard: %s", err.Error())
	} else {
		globals.Logger.Infof("SSH command copied to clipboard")
	}

	if globals.JSON {
		return display.PrintJSON(droplet)
	}

	fmt.Println(sshCommand)
	return nil
}
