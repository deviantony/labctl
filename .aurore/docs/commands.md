---
title: CLI Commands
summary: Kong-based CLI structure covering droplet lifecycle, status, and options commands.
covers:
  - cmd/**
  - internal/commands/**
  - internal/display/**
scanned_at_commit: 7a525d5bc240aec68ade915265c1dd139b70bb2f
layer: cli
order: 1
---

## Overview

labctl uses [Kong](https://github.com/alecthomas/kong) (v1.14.0) for CLI parsing. The top-level CLI struct in `internal/commands/cli.go` exposes five commands — `create`, `ls`, `rm`, `status`, and `options` — plus global flags `--debug` (verbose logging), `--json` (JSON output), and `--version`.

The entry point in `cmd/labctl.go` handles initialization: parse CLI args, set up a zap logger, load the YAML config, create the DigitalOcean client, and build a `Globals` struct that every command receives alongside the client. Commands that don't need a config or API client (currently `options`) are routed separately — they run before config loading, receiving only `*Globals`.

## Globals

Defined in `internal/commands/globals.go`. A simple struct carrying shared state passed to all command `Run` methods:

- `JSON bool` — whether to output in JSON format
- `Logger *zap.SugaredLogger` — structured logger

Kong's `Run` method injects both the `*do.Client` and `*Globals` into each command's `Run(client *do.Client, globals *Globals) error` signature via its dependency injection. Commands that don't need the client (like `options`) accept only `*Globals`.

## Commands

### `create`

Defined in `internal/commands/create.go`. The primary command and the default when `labctl` is invoked with arguments (`default:"withargs"` tag).

**Flags:**
- `-r` / `--region` — region alias (`usw`, `use`, `eu`, `ap`, `au`), defaults to `eu`
- `-s` / `--size` — size alias (`xs`, `s`, `m`, `l`, `xl`), defaults to `xs`
- `-n` / `--name` — custom droplet name; if omitted, a random petname is generated via `pkg/random`

**Flow:** generate name (petname if not provided) -> create droplet via DO client -> receive action HREF -> poll via two-phase readiness check (Actions API then SSH) -> copy SSH command to clipboard (warns on failure, doesn't abort) -> print SSH command or JSON output.

The clipboard integration uses `atotto/clipboard`. If the clipboard is unavailable (e.g. headless server), it logs a warning and continues — the SSH command is always printed to stdout regardless.

### `ls`

Defined in `internal/commands/ls.go`. Lists all droplets with the configured tag. Renders a table with columns: ID, Name, IPv4, Region, Size, Uptime. The Uptime column computes a human-readable duration from the droplet's creation timestamp (e.g. `5m`, `2h30m`, `3d12h`). With `--json`, outputs the droplet array as indented JSON. If no droplets exist, logs "No droplets found" and exits cleanly.

### `rm`

Defined in `internal/commands/rm.go`. Takes one or more droplet IDs as positional arguments (`arg:"" required:""`). Deletes all droplets in parallel using goroutines with a `sync.WaitGroup`. Each deletion logs progress independently. If some deletions fail, it collects errors under a mutex and returns a combined error message — it doesn't stop on the first failure.

### `status`

Defined in `internal/commands/status.go`. A health-check command that outputs the labctl version and whether the DigitalOcean API is reachable. Calls `client.CheckAPI()` which hits `Account.Get()` — a lightweight call that validates the API token. Supports `--json` output for automation.

### `options`

Defined in `internal/commands/options.go`. Displays available region and size alias-to-slug mappings by calling `do.RegionOptions()` and `do.SizeOptions()`. This command doesn't need a config file or API client — it runs before config loading in `cmd/labctl.go`. With `--json`, outputs the mappings as a JSON object with `regions` and `sizes` arrays.

## Version Flag

Defined in `internal/commands/version.go`. A custom `VersionFlag` type (a `string`) that implements Kong's `BeforeApply` hook — it prints the version from `kong.Vars` and exits before any command runs. This keeps `--version` working as a standalone flag independent of the `status` command.

## Display

Table rendering lives in `internal/display/table.go`. Three functions:

- `DisplayDroplets` — renders a `go-pretty/v6` table with ID, Name, IPv4, Region, Size, Uptime columns. The Uptime column is computed by `formatUptime`, which parses the RFC 3339 creation timestamp and formats the duration in a compact style (seconds, minutes, hours+minutes, or days+hours).
- `DisplayOptions` — renders a labeled table of alias-to-slug mappings with Alias and DigitalOcean Slug columns. Used by the `options` command.
- `PrintJSON` — writes any value as indented JSON to stdout.

All write directly to `os.Stdout`. The display package has no provider-specific knowledge — it works with `types.Droplet` structs and `do.Option` structs.
