package flask

import (
	"github.com/deviantony/labctl/commands/context"
	"github.com/deviantony/labctl/ssh"
)

// ExecCommand creates a SSH session to the given flask - matching an ID or ID prefix.
type ExecCommand struct {
	ID string `arg:"" help:"Flask ID." name:"Flask ID"`
}

// Run executes the exec command.
func (cmd *ExecCommand) Run(cmdCtx context.CommandExecutionContext) error {
	flaskManager, err := context.BuildManagerFromProvider(cmdCtx)
	if err != nil {
		return err
	}

	flask, err := flaskManager.GetFlask(cmd.ID)
	if err != nil {
		return err
	}

	return ssh.ExecuteSSHSession(cmdCtx.Logger, flask.Ipv4)
}
