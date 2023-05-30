package keyring

import (
	"time"

	"github.com/deviantony/labctl/internal/commands/context"
	"github.com/deviantony/labctl/internal/dockerhub"
	"github.com/deviantony/labctl/pkg/prompt"
)

// AddCommand adds a new key to the keyring.
type AddCommand struct {
	Label    string        `arg:"" help:"Label associated to the key." name:"Key label" optional:""`
	Validity time.Duration `help:"Validity of the key. Program will automatically pause when specified and remove the key after the duration."`
}

// Run executes the add command.
func (cmd *AddCommand) Run(cmdCtx context.CommandExecutionContext) error {
	code, err := prompt.AskFor2FACode()
	if err != nil {
		return err
	}

	client := dockerhub.NewDockerHubClient(cmdCtx.Config.DockerHub, cmdCtx.Logger, code)

	token, err := client.CreateAccessToken(cmd.Label)
	if err != nil {
		return err
	}

	cmdCtx.Logger.Infow("Key successfully added to the keyring",
		"token", token.Token,
	)

	if cmd.Validity > 0 {
		cmdCtx.Logger.Infow("Key validity specified, program will enter pause mode",
			"validity", cmd.Validity,
		)

		time.Sleep(cmd.Validity)

		cmdCtx.Logger.Infow("Key validity expired, removing key from the keyring",
			"validity", cmd.Validity,
		)

		err = client.DeleteAccessToken(token.Uuid)
		if err != nil {
			return err
		}

		cmdCtx.Logger.Info("Key successfully removed from the keyring")
	}

	return nil
}
