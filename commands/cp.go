package commands

import (
	"github.com/deviantony/labctl/do"
	"github.com/deviantony/labctl/ssh"
)

// CpCommand copies a file or a directory to a remote VPS.
type CpCommand struct {
	ID         int    `arg:"" help:"VPS ID." name:"VPS ID"`
	LocalPath  string `arg:"" help:"Path to local folder or file." name:"Local path"`
	RemotePath string `arg:"" help:"Path to remote folder or file." name:"Remote path"`
}

// Run executes the cp command.
func (cmd *CpCommand) Run(cmdCtx *CommandExecutionContext) error {
	vpsBuilder := do.NewDOVPSBuilder(cmdCtx.context, cmdCtx.config.DO, cmdCtx.logger)

	vps, err := vpsBuilder.GetVPS(cmd.ID)
	if err != nil {
		return err
	}

	return ssh.CopyToRemote(cmdCtx.logger, vps.Ipv4, cmd.LocalPath, cmd.RemotePath)
}
