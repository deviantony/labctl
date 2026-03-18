---
type: overview
title: Project Overview
summary: CLI tool for managing ephemeral DigitalOcean droplets for lab environments.
scanned_at_commit: 98491e9d01c27506d1add05ed4b935117de7ecf0
---

# labctl

labctl solves a specific problem: spinning up and tearing down DigitalOcean droplets quickly from the command line. The typical workflow is `labctl create`, which provisions a droplet, waits until it's SSH-reachable, copies the SSH command to your clipboard, and prints it. When you're done, `labctl rm <id>` destroys it. These are disposable lab machines — create, use, destroy.

## How It Works

The CLI is intentionally minimal. Four commands cover the full lifecycle:

- **`create`** — provisions a droplet with a random petname (like `brave-falcon`), polls until SSH is reachable, copies the SSH command to clipboard
- **`ls`** — lists all droplets tagged with `labctl`
- **`rm`** — deletes one or more droplets by ID
- **`status`** — health check showing version and whether the DO API token is valid (useful for automation)

All droplets are tagged (default: `labctl`) so labctl only sees its own resources. Droplets are assigned to a configured DigitalOcean project for organizational clarity.

## Design Choices

**No abstraction layers.** Earlier versions of labctl supported multiple providers (LXD, DigitalOcean) behind a `FlaskManager` interface. The revamp removed that abstraction entirely — the code talks directly to the `godo` client. This makes the codebase dramatically simpler since there's only one provider to support.

**Polling for readiness.** After creating a droplet, labctl polls until three conditions are met: the droplet status is "active", it has a public IPv4 address, and TCP port 22 accepts connections. This ensures the SSH command actually works when it's handed to the user.

**Random naming.** Droplet names default to petnames (adjective-animal combinations) generated from embedded word lists. This avoids naming conflicts without requiring user input.

## How It's Built

labctl is a single Go binary using [Kong](./commands.md) for CLI parsing and the [DigitalOcean godo SDK](./infrastructure.md) for API calls.

- **[CLI Commands](./commands.md)** — `cmd/labctl.go` and `internal/commands/` define the user-facing interface
- **[Infrastructure](./infrastructure.md)** — `internal/do/`, `internal/display/`, `pkg/random/`, and `types/` implement the DO client, output formatting, and name generation
- **[Configuration](./configuration.md)** — `internal/config/` and `~/.labctl/config.yml` define credentials, infrastructure settings, and polling behavior
