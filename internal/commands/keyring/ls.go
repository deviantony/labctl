package keyring

import (
	"github.com/deviantony/labctl/internal/commands/context"
	"github.com/deviantony/labctl/internal/dockerhub"
	"github.com/deviantony/labctl/internal/terminal"
)

// LsCommand lists all keys in the keyring.
type LsCommand struct{}

// Run executes the ls command.
func (cmd *LsCommand) Run(cmdCtx context.CommandExecutionContext) error {
	code, err := terminal.AskFor2FACode()
	if err != nil {
		return err
	}

	client := dockerhub.NewDockerHubClient(cmdCtx.Config.DockerHub, cmdCtx.Logger, code)

	tokens, err := client.ListAccessTokens()
	if err != nil {
		return err
	}

	terminal.DisplayDockerHubAccessTokens(tokens)

	return nil
}
