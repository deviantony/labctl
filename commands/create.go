package commands

import (
	"github.com/alecthomas/kong"
	"github.com/deviantony/labctl/display"
	"github.com/deviantony/labctl/do"
	"github.com/deviantony/labctl/ssh"
)

// CreateOptionsFlag is used to display a list of available options for VPS creation.
type CreateOptionsFlag bool

func (o CreateOptionsFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (o CreateOptionsFlag) IsBool() bool                         { return true }
func (o CreateOptionsFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	display.DisplayOptionList()
	app.Exit(0)
	return nil
}

// CreateCommand creates a new VPS.
type CreateCommand struct {
	Options    CreateOptionsFlag `help:"Display a list of available options for VPS creation and quit." short:"o"`
	AutoRemove bool              `help:"Automatically remove VPS after SSH session ends." default:"true" negatable:""`
	Background bool              `help:"Returns VPS ID after creation and do not run a SSH session" short:"d" default:"false"`
	VPSName    string            `help:"Name of the VPS." short:"n"`
	Region     string            `help:"Region of the VPS." short:"r" enum:"usw,use,eu,ap,nz" default:"eu"`
	Size       string            `help:"Size of the VPS." short:"s" enum:"xs,s,m,l,xl" default:"xs"`
}

// Run executes the create command.
func (cmd *CreateCommand) Run(cmdCtx *CommandExecutionContext) error {
	vpsName := cmd.VPSName
	if vpsName == "" {
		vpsName = GeneratePetName(2, "-")
	}

	cmdCtx.logger.Infow("Creating VPS",
		"VPS name", vpsName,
		"Region", cmd.Region,
		"Size", cmd.Size,
		"Auto remove VPS", cmd.AutoRemove && !cmd.Background,
	)

	vpsBuilder := do.NewDOVPSBuilder(cmdCtx.context, cmdCtx.config.DO, cmdCtx.logger)

	vpsID, err := vpsBuilder.CreateVPS(vpsName, cmd.Region, cmd.Size)
	if err != nil {
		return err
	}

	vpsIPaddr, err := vpsBuilder.WaitForVPSToBeReady(vpsID)
	if err != nil {
		return err
	}

	if cmd.Background {
		cmdCtx.logger.Infow("VPS created",
			"ID", vpsID,
		)
		return nil
	}

	err = ssh.ExecuteSSHSession(cmdCtx.logger, vpsIPaddr)
	if err != nil {
		return err
	}

	if !cmd.AutoRemove {
		return nil
	}

	cmdCtx.logger.Infow("Automatically removing VPS")
	return vpsBuilder.RemoveVPS(vpsID)
}
