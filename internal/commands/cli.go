package commands

var CLI struct {
	Debug   bool   `help:"Enable debug mode."`
	JSON    bool   `help:"Output in JSON format." name:"json"`
	Version VersionFlag `name:"version" help:"Print version information and quit."`

	Create  CreateCommand  `cmd:"" help:"Create a new droplet." default:"withargs"`
	Ls      LsCommand      `cmd:"" help:"List droplets."`
	Rm      RmCommand      `cmd:"" help:"Remove one or more droplets."`
	Status  StatusCommand  `cmd:"" help:"Show version and API connectivity."`
	Options OptionsCommand `cmd:"" help:"Show available region and size options."`
}
