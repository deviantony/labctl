package flask

import (
	"github.com/alecthomas/kong"
	"github.com/deviantony/labctl/display"
	"github.com/deviantony/labctl/lxd"
	"github.com/deviantony/labctl/random"
	"github.com/deviantony/labctl/ssh"
	"github.com/deviantony/labctl/types"
)

// CreateOptionsFlag is used to display a list of available options for flask creation.
type CreateOptionsFlag bool

func (o CreateOptionsFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (o CreateOptionsFlag) IsBool() bool                         { return true }
func (o CreateOptionsFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	display.DisplayOptionList()
	app.Exit(0)
	return nil
}

// CreateCommand creates a new flask.
type CreateCommand struct {
	Options    CreateOptionsFlag `help:"Displays a list of available options for flask creation and quit." short:"o"`
	AutoRemove bool              `help:"Automatically removes the flask after SSH session ends." default:"true" negatable:""`
	Background bool              `help:"Returns the flask ID after creation and do not run a SSH session" short:"d" default:"false"`
	Name       string            `help:"Name of the flask." short:"n"`
	Region     string            `help:"Region of the flask." short:"r" enum:"usw,use,eu,ap,nz" default:"eu"`
	Size       string            `help:"Size of the flask." short:"s" enum:"xs,s,m,l,xl" default:"xs"`
}

// Run executes the create command.
func (cmd *CreateCommand) Run(cmdCtx types.CommandExecutionContext) error {
	flaskName := cmd.Name
	if flaskName == "" {
		flaskName = random.GeneratePetName(2, "-")
	}

	cmdCtx.Logger.Infow("Creating new flask",
		"Name", flaskName,
		"Region", cmd.Region,
		"Size", cmd.Size,
		"Auto remove", cmd.AutoRemove && !cmd.Background,
	)

	// flaskManager := do.NewFlaskManager(cmdCtx.Context, cmdCtx.Config.DO, cmdCtx.Logger)
	flaskManager, err := lxd.NewFlaskManager(cmdCtx.Context, cmdCtx.Config.LXD, cmdCtx.Logger)
	if err != nil {
		return err
	}

	// flaskCfg := types.FlaskConfig{
	// 	Size:   cmd.Size,
	// 	Region: cmd.Region,
	// } // Only for DO

	flask, err := flaskManager.CreateFlask(flaskName)
	if err != nil {
		return err
	}

	err = flaskManager.WaitUntilFlaskIsReady(&flask)
	if err != nil {
		return err
	}

	// flaskIP, err := flaskManager.WaitUntilFlaskIsReady(flaskID)
	// if err != nil {
	// 	return err
	// }

	if cmd.Background {
		cmdCtx.Logger.Infow("Flask created",
			"ID", flask.LXD.ID,
			"IP", flask.Ipv4,
		)
		return nil
	}

	err = ssh.ExecuteSSHSession(cmdCtx.Logger, flask.Ipv4)
	if err != nil {
		return err
	}

	if !cmd.AutoRemove {
		return nil
	}

	cmdCtx.Logger.Infow("Automatically removing flask")
	return flaskManager.RemoveFlask(flask)
}
