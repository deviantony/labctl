package flask

import (
	"github.com/deviantony/labctl/lxd"
	"github.com/deviantony/labctl/ssh"
	"github.com/deviantony/labctl/types"
)

// CpCommand copies a file or a directory to a flask.
type CpCommand struct {
	ID         string `arg:"" help:"Flask ID." name:"Flask ID"`
	LocalPath  string `arg:"" help:"Path to local folder or file." name:"Local path"`
	RemotePath string `arg:"" help:"Path to remote folder or file." name:"Remote path"`
}

// Run executes the cp command.
func (cmd *CpCommand) Run(cmdCtx types.CommandExecutionContext) error {
	// flaskManager := do.NewFlaskManager(cmdCtx.Context, cmdCtx.Config.DO, cmdCtx.Logger)
	flaskManager, err := lxd.NewFlaskManager(cmdCtx.Context, cmdCtx.Config.LXD, cmdCtx.Logger)
	if err != nil {
		return err
	}

	flask, err := flaskManager.GetFlask(cmd.ID)
	if err != nil {
		return err
	}

	return ssh.CopyToRemote(cmdCtx.Logger, flask.Ipv4, cmd.LocalPath, cmd.RemotePath)
}
