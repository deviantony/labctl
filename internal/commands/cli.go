package commands

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/deviantony/labctl/internal/commands/flask"
	"github.com/deviantony/labctl/internal/commands/keyring"
)

// VersionFlag is used to display the version of the CLI.
type VersionFlag string

func (v VersionFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                         { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	fmt.Println(vars["version"])
	app.Exit(0)
	return nil
}

var CLI struct {
	// Generic options
	Debug    bool        `help:"Enable debug mode."`
	Provider string      `help:"Provider to use." enum:"do,lxd"  default:"lxd"`
	Version  VersionFlag `name:"version" help:"Print version information and quit"`

	// Flasks
	Flask struct {
		Create flask.CreateCommand `cmd:"" help:"Create a new flask." default:"withargs"`
		Ls     flask.LsCommand     `cmd:"" help:"List existing flasks."`
		Cp     flask.CpCommand     `cmd:"" help:"Copy a file or a directory to a flask."`
		Exec   flask.ExecCommand   `cmd:"" help:"Create a SSH session to the given flask ID."`
		Rm     flask.RmCommand     `cmd:"" help:"Remove a flask."`
	} `cmd:"" help:"Manage flasks."`

	// Keyring
	Keyring struct {
		Add keyring.AddCommand    `cmd:"" help:"Add a new key to the keyring."`
		Rm  keyring.RemoveCommand `cmd:"" help:"Remove a key from the keyring."`
		Ls  keyring.LsCommand     `cmd:"" help:"List keys in the keyring."`
	} `cmd:"" help:"Manage keyring."`
}
