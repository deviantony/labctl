package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/alecthomas/kong"
	"github.com/deviantony/labctl/internal/commands"
	"github.com/deviantony/labctl/internal/config"
	"github.com/deviantony/labctl/internal/do"
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

	zapCfg := zap.NewProductionConfig()
	zapCfg.Encoding = "console"
	zapCfg.DisableStacktrace = true
	zapCfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.TimeOnly)
	logger, err := zapCfg.Build()
	if err != nil {
		return nil, err
	}

	return logger.Sugar(), nil
}

func main() {
	cliCtx := kong.Parse(&commands.CLI,
		kong.Name("labctl"),
		kong.Description("Manage DigitalOcean droplets from the command line."),
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

	configPath := os.Getenv(config.ConfigEnvOverride)
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			logger.Fatalf("Unable to determine home directory: %s", err)
		}
		configPath = filepath.Join(home, config.ConfigPath)
	}

	cfg, err := config.NewConfig(configPath)
	if err != nil {
		logger.Fatalf("Unable to read configuration file: %s", err)
	}

	client := do.NewClient(context.Background(), cfg, logger)
	globals := &commands.Globals{
		JSON:   commands.CLI.JSON,
		Logger: logger,
	}

	err = cliCtx.Run(client, globals)
	cliCtx.FatalIfErrorf(err)
}
