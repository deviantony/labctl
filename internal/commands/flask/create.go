package flask

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/atotto/clipboard"
	"github.com/deviantony/labctl/internal/commands/context"
	"github.com/deviantony/labctl/internal/config"
	terminal "github.com/deviantony/labctl/internal/display"
	"github.com/deviantony/labctl/pkg/random"
	"github.com/deviantony/labctl/pkg/ssh"
	"github.com/deviantony/labctl/types"
)

// CreateOptionsFlag is used to display a list of available options for flask creation.
type CreateOptionsFlag bool

func (o CreateOptionsFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (o CreateOptionsFlag) IsBool() bool                         { return true }
func (o CreateOptionsFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	terminal.DisplayOptionList()
	app.Exit(0)
	return nil
}

// CreateCommand creates a new flask.
type CreateCommand struct {
	Options    CreateOptionsFlag `help:"Displays a list of available options for flask creation and quit." short:"o"`
	AutoRemove bool              `help:"Automatically removes the flask after SSH session ends." default:"true" negatable:""`
	Background bool              `help:"Returns the flask ID after creation and do not run a SSH session" short:"d" default:"false"`
	Name       string            `help:"Name of the flask." short:"n"`
	Profile    string            `help:"Profile to use for flask creation (lxd only). Overrides the default flask profile." short:"p"`
	Image      string            `help:"Image to use for flask creation (lxd only). Overrides the default flask image." short:"i"`
	Region     string            `help:"Region of the flask (cloud only)." short:"r" enum:"usw,use,eu,ap,nz" default:"eu"`
	Size       string            `help:"Size of the flask." short:"s" enum:"xs,s,m,l,xl" default:"xs"`
}

// Run executes the create command.
func (cmd *CreateCommand) Run(cmdCtx context.CommandExecutionContext) error {
	flaskManager, err := context.BuildManagerFromProvider(cmdCtx)
	if err != nil {
		return err
	}

	flaskName := cmd.Name
	if flaskName == "" {
		flaskName = random.GeneratePetName(2, "-")
	}

	if cmdCtx.Config.GetProvider() == config.PROVIDER_DO {
		cmdCtx.Logger.Infow("Creating new flask",
			"Name", flaskName,
			"Region", cmd.Region,
			"Size", cmd.Size,
			"Auto remove", cmd.AutoRemove && !cmd.Background,
		)
	} else {
		cmdCtx.Logger.Infow("Creating new flask",
			"Name", flaskName,
			"Size", cmd.Size,
			"Profile", cmd.Profile,
			"Image", cmd.Image,
			"Auto remove", cmd.AutoRemove && !cmd.Background,
		)
	}

	flaskCfg := types.FlaskConfig{
		Size:    cmd.Size,
		Region:  cmd.Region,
		Profile: cmd.Profile,
		Image:   cmd.Image,
	}

	flask, err := flaskManager.CreateFlask(flaskName, flaskCfg)
	if err != nil {
		return err
	}

	err = flaskManager.WaitUntilFlaskIsReady(&flask)
	if err != nil {
		return err
	}

	if cmd.Background {
		if cmdCtx.Config.GetProvider() == config.PROVIDER_DO {
			cmdCtx.Logger.Infow("Flask created",
				"ID", flask.DO.ID,
				"IP", flask.Ipv4,
			)
		} else {
			cmdCtx.Logger.Infow("Flask created",
				"ID", flask.LXD.ID,
				"IP", flask.Ipv4,
			)
		}

		sshCommand := fmt.Sprintf("ssh -o StrictHostKeyChecking=no root@%s", flask.Ipv4)

		err = clipboard.WriteAll(sshCommand)
		if err != nil {
			cmdCtx.Logger.Warnf("Unable to add command to clipboard. Error: %s", err.Error())
			cmdCtx.Logger.Infoln("Use the command below to SSH into the flask")
			cmdCtx.Logger.Infoln(sshCommand)
			return err
		}

		cmdCtx.Logger.Infoln("Command added to clipboard. Paste or use the command below to SSH into the flask")
		cmdCtx.Logger.Infoln(sshCommand)

		return nil
	}

	err = ssh.ExecuteSSHSession(flask.Ipv4)
	if err != nil {
		return err
	}

	if !cmd.AutoRemove {
		return nil
	}

	cmdCtx.Logger.Infow("Automatically removing flask")
	return flaskManager.RemoveFlask(flask)
}
