package flask

import (
	"github.com/deviantony/labctl/commands/context"
	"github.com/deviantony/labctl/ssh"
)

// CpCommand copies a file or a directory to a flask.
type CpCommand struct {
	ID         string `arg:"" help:"Flask ID." name:"Flask ID"`
	LocalPath  string `arg:"" help:"Path to local folder or file." name:"Local path"`
	RemotePath string `arg:"" help:"Path to remote folder or file." name:"Remote path"`
}

// Run executes the cp command.
func (cmd *CpCommand) Run(cmdCtx context.CommandExecutionContext) error {
	flaskManager, err := context.BuildManagerFromProvider(cmdCtx)
	if err != nil {
		return err
	}

	flask, err := flaskManager.GetFlask(cmd.ID)
	if err != nil {
		return err
	}

	return ssh.CopyToRemote(cmdCtx.Logger, flask.Ipv4, cmd.LocalPath, cmd.RemotePath)
}
