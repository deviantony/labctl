# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is labctl

labctl is a CLI tool for managing ephemeral DigitalOcean droplets. It handles the full lifecycle: create (with auto-generated petnames), poll until SSH-ready, copy SSH command to clipboard, list, and destroy.

## Build & Run

```bash
make build          # Build static binary to dist/labctl
make install        # Build and install to $GOPATH/bin
make clean          # Remove dist/
```

Releases are handled by GoReleaser via GitHub Actions on `v*` tag push. See `.goreleaser.yml` and `.github/workflows/release.yml`.

There are no tests or linters configured.

## Architecture

**Entry point**: `cmd/labctl.go` — initializes zap logger, parses CLI with Kong, loads config, creates DigitalOcean client, routes to command.

**Package layout**:
- `internal/commands/` — CLI commands (create, ls, rm, status, options) and shared state (`globals.go` holds JSON flag + logger)
- `internal/config/` — YAML config loader (`~/.labctl/config.yml` or `$LABCTL_CONFIG`)
- `internal/do/` — DigitalOcean API client wrapping `godo`
- `internal/display/` — Table rendering with `go-pretty` and JSON output mode
- `types/` — `Droplet` struct and `VERSION` variable (set via ldflags at build time)
- `pkg/random/` — Petname generator for auto-naming droplets

**CLI framework**: Kong with struct-tag-based command definitions. Commands implement `Run(globals *Globals)` methods. The command router is in `internal/commands/cli.go`.

**Key flow (create command)**: Resolves region/size aliases → calls DO API → polls droplet status + IP + SSH port 22 readiness → copies `ssh root@<ip>` to clipboard.

## Configuration

Config file at `~/.labctl/config.yml` with fields: `apiToken`, `projectID`, `sshKeyFingerprint`, `baseImage`, `pollInterval`, `pollTimeout`, `tagName`. See `config.example.yml`.

## Version

Version is a `var` in `types/droplet.go`, defaulting to `dev`. It is overridden at build time via `-ldflags -X` in the Makefile and GoReleaser config.
