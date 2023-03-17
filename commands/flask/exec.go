package flask

import (
	"github.com/deviantony/labctl/do"
	"github.com/deviantony/labctl/ssh"
	"github.com/deviantony/labctl/types"
)

// ExecCommand creates a SSH session to the given flask - matching an ID or ID prefix.
type ExecCommand struct {
	ID int `arg:"" help:"Flask ID." name:"Flask ID"`
}

// Run executes the exec command.
func (cmd *ExecCommand) Run(cmdCtx types.CommandExecutionContext) error {
	flaskManager := do.NewFlaskManager(cmdCtx.Context, cmdCtx.Config.DO, cmdCtx.Logger)

	flask, err := flaskManager.GetFlask(cmd.ID)
	if err != nil {
		return err
	}

	return ssh.ExecuteSSHSession(cmdCtx.Logger, flask.Ipv4)
}
