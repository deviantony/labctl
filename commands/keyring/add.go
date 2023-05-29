package keyring

import (
	"github.com/deviantony/labctl/commands/context"
	"github.com/deviantony/labctl/dockerhub"
	"github.com/deviantony/labctl/terminal"
)

// AddCommand adds a new key to the keyring.
type AddCommand struct {
	Label string `arg:"" help:"Label associated to the key." name:"Key label" optional:""`
}

// Run executes the add command.
func (cmd *AddCommand) Run(cmdCtx context.CommandExecutionContext) error {
	code, err := terminal.AskFor2FACode()
	if err != nil {
		return err
	}

	client := dockerhub.NewDockerHubClient(cmdCtx.Config.DockerHub, cmdCtx.Logger, code)

	token, err := client.CreateAccessToken(cmd.Label)
	if err != nil {
		return err
	}

	cmdCtx.Logger.Infow("Key successfully added to the keyring",
		"token", token,
	)

	return nil
}
