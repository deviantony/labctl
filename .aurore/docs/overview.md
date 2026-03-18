---
type: overview
title: Project Overview
summary: CLI tool for managing ephemeral DigitalOcean droplets for lab environments.
scanned_at_commit: 712a73a7df12e354f38b27a0f3d6093e410ca096
---

# labctl

labctl solves a specific problem: spinning up and tearing down DigitalOcean droplets quickly from the command line. The typical workflow is `labctl create`, which provisions a droplet, waits until it's SSH-reachable, copies the SSH command to your clipboard, and prints it. When you're done, `labctl rm <id>` destroys it. These are disposable lab machines — create, use, destroy.

## How It Works

The CLI is intentionally minimal. Five commands cover the full lifecycle:

- **`create`** — provisions a droplet with a random petname (like `brave-falcon`), polls until SSH is reachable, copies the SSH command to clipboard
- **`ls`** — lists all droplets tagged with `labctl`, showing uptime for each
- **`rm`** — deletes one or more droplets by ID, in parallel
- **`status`** — health check showing version and whether the DO API token is valid (useful for automation)
- **`options`** — displays available region and size aliases with their DigitalOcean slug mappings

All droplets are tagged (default: `labctl`) so labctl only sees its own resources. Droplets are assigned to a configured DigitalOcean project for organizational clarity.

## Design Choices

**No abstraction layers.** The code talks directly to the `godo` client — there's no provider interface or abstraction. With only DigitalOcean to support, an indirection layer would add complexity without value.

**Two-phase readiness polling.** After creating a droplet, labctl first polls the DigitalOcean Actions API until the creation action completes (droplet is active with an IP), then switches to a tighter loop that dials TCP port 22 until SSH accepts connections. This two-phase approach avoids unnecessary Droplets.Get calls during provisioning while ensuring the SSH command actually works when it's handed to the user.

**Random naming.** Droplet names default to petnames (adjective-animal combinations) generated from embedded word lists. This avoids naming conflicts without requiring user input.

## How It's Built

labctl is a single Go binary using [Kong](./commands.md) for CLI parsing and the [DigitalOcean godo SDK](./infrastructure.md) for API calls.

- **[CLI Commands](./commands.md)** — `cmd/labctl.go` and `internal/commands/` define the user-facing interface
- **[Infrastructure](./infrastructure.md)** — `internal/do/`, `internal/display/`, `pkg/random/`, and `types/` implement the DO client, output formatting, and name generation
- **[Configuration](./configuration.md)** — `internal/config/` and `~/.labctl/config.yml` define credentials, infrastructure settings, and polling behavior

## Build & Release

The project uses a Makefile for local development (`make build`, `make install`) and GoReleaser via GitHub Actions for releases. Pushing a `v*` tag triggers the release workflow, which builds static binaries for Linux and macOS (amd64/arm64) and creates a GitHub release. An `install.sh` script provides curl-based installation for end users. Version is injected at build time via ldflags — local builds get the `git describe` output, releases get the tag version.
