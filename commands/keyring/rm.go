package keyring

import (
	"github.com/deviantony/labctl/commands/context"
	"github.com/deviantony/labctl/dockerhub"
	"github.com/deviantony/labctl/terminal"
)

// RemoveCommand removes a key from the keyring.
type RemoveCommand struct {
	ID string `arg:"" help:"UUID associated with the key." name:"Key UUID"`
}

// Run executes the rm command.
func (cmd *RemoveCommand) Run(cmdCtx context.CommandExecutionContext) error {
	code, err := terminal.AskFor2FACode()
	if err != nil {
		return err
	}

	client := dockerhub.NewDockerHubClient(cmdCtx.Config.DockerHub, cmdCtx.Logger, code)

	err = client.DeleteAccessToken(cmd.ID)
	if err != nil {
		return err
	}

	cmdCtx.Logger.Infow("Key successfully removed from the keyring",
		"uuid", cmd.ID,
	)

	return nil
}
