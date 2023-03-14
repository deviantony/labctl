package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/alecthomas/kong"
	"github.com/deviantony/labctl/commands"
	"github.com/deviantony/labctl/config"
	"github.com/deviantony/labctl/random"
	"github.com/deviantony/labctl/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func initializeLogger(debug bool) (*zap.SugaredLogger, error) {
	if debug {
		logger, err := zap.NewDevelopment()
		if err != nil {
			return nil, err
		}

		return logger.Sugar(), nil
	}

	config := zap.NewProductionConfig()
	config.Encoding = "console"
	config.DisableStacktrace = true
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.TimeOnly)
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return logger.Sugar(), nil
}

func main() {
	ctx := context.Background()

	cliCtx := kong.Parse(&commands.CLI,
		kong.Name("labctl"),
		kong.Description("Control your lab environment from the command line."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}),
		kong.Vars{
			"version": types.VERSION,
		})

	logger, err := initializeLogger(commands.CLI.Debug)
	if err != nil {
		log.Fatalf("Unable to initialize logger: %s", err)
	}
	defer logger.Sync()

	configPath := os.Getenv(config.CONFIG_ENV_OVERRIDE)
	if configPath == "" {
		configPath = filepath.Join(os.Getenv("HOME"), config.CONFIG_PATH)
	}

	cfg, err := config.NewConfig(configPath)
	if err != nil {
		logger.Fatalf("Unable to read configuration file: %s", err)
	}

	random.NonDeterministicMode()

	cmdCtx := types.NewCommandExecutionContext(ctx, cfg, logger)
	err = cliCtx.Run(cmdCtx)
	cliCtx.FatalIfErrorf(err)
}
