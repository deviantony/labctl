package commands

import (
	"github.com/deviantony/labctl/do"
	"github.com/deviantony/labctl/ssh"
)

// ExecCommand creates a SSH session to the given VPS matching an ID or ID prefix.
type ExecCommand struct {
	ID int `arg:"" help:"VPS ID." name:"VPS ID"`
}

// Run executes the exec command.
func (cmd *ExecCommand) Run(cmdCtx *CommandExecutionContext) error {
	vpsBuilder := do.NewDOVPSBuilder(cmdCtx.context, cmdCtx.config.DO, cmdCtx.logger)

	vps, err := vpsBuilder.GetVPS(cmd.ID)
	if err != nil {
		return err
	}

	return ssh.ExecuteSSHSession(cmdCtx.logger, vps.Ipv4)
}
