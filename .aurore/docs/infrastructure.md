---
title: Infrastructure
summary: DigitalOcean client, two-phase readiness polling, petname generation, and type definitions.
covers:
  - internal/do/**
  - pkg/**
  - types/**
scanned_at_commit: 712a73a7df12e354f38b27a0f3d6093e410ca096
layer: core
order: 2
---

## Overview

The infrastructure layer is intentionally thin. There's no provider abstraction or interface — the [CLI commands](./commands.md) receive a `*do.Client` directly and call its methods. With only DigitalOcean supported, an abstraction layer would add complexity without value.

## DigitalOcean Client

Defined in `internal/do/client.go`. Wraps a `godo.Client` (v1.178.0) with labctl-specific operations.

### Construction

`NewClient(ctx, cfg, logger)` creates the client from config. If no tag name is configured, it defaults to `"labctl"`. The godo client is created via `godo.NewFromToken(cfg.APIToken)`.

### Operations

**`CreateDroplet(name, region, size)`** — creates a droplet with the configured base image and SSH key, tags it, enables monitoring, and fires off a goroutine to assign it to the configured project (fire-and-forget with a detached context — failure is non-fatal). Region and size aliases (like `eu` and `xs`) are resolved to DO slugs by `mapRegion` and `mapSize` before the API call. Returns a `types.Droplet` with the ID and creation parameters, plus an action HREF extracted from the API response links for use in readiness polling.

**`ListDroplets()`** — paginated listing of all droplets with the configured tag (up to 200 per page). For each droplet, extracts the public IPv4 address (falls back to `"-"` if unavailable) and the creation timestamp.

**`RemoveDroplet(id)`** — deletes a single droplet by ID.

**`CheckAPI()`** — calls `Account.Get()` to verify the API token is valid. Used by the [status command](./commands.md#status).

**`WaitUntilReady(droplet, actionHREF)`** — two-phase readiness polling:

1. **Phase 1 — `waitForActive`**: polls the DigitalOcean Actions API using the action HREF returned from droplet creation. This is more efficient than polling `Droplets.Get` repeatedly because the action endpoint tells you exactly when provisioning completes. Once the action status is `"completed"`, it fetches the droplet to get the public IP. If the action errors, it fails immediately. Falls back to `Droplets.Get` polling if no action HREF is available.

2. **Phase 2 — `waitForSSH`**: once the droplet is active with an IP, switches to a tighter polling loop (every 2 seconds) that dials TCP port 22 with a 1-second timeout. This phase is purely local network I/O — no API calls — so it can poll more aggressively. Checks immediately before the first tick to avoid an unnecessary 2-second wait.

Both phases share a single deadline derived from `config.PollTimeout`.

### Region and Size Mapping

The `Option` type represents a CLI alias mapped to a DigitalOcean slug. Two exported functions return the full mapping tables:

- `RegionOptions()` — returns all region mappings: `usw`->`sfo3`, `use`->`nyc1`, `eu`->`fra1`, `ap`->`sgp1`, `au`->`syd1`
- `SizeOptions()` — returns all size mappings: `xs`->`s-1vcpu-512mb-10gb`, `s`->`s-1vcpu-1gb`, `m`->`s-2vcpu-4gb`, `l`->`s-4vcpu-8gb`, `xl`->`s-8vcpu-16gb`

These are used by the [options command](./commands.md#options) to display available aliases, and internally by `mapRegion`/`mapSize` (which build lookup maps from the same data) to resolve aliases during droplet creation. If an unrecognized value is passed, both map functions return it as-is — allowing direct DO slugs as a fallback.

## Types

Defined in `types/droplet.go`. Two things live here:

- `VERSION` variable (default `"dev"`) — the application version, overridden at build time via `-ldflags -X`. Referenced by the version flag and status command.
- `Droplet` struct — represents a DigitalOcean droplet with fields: `ID`, `Name`, `IPv4`, `Region`, `Size`, `CreatedAt`. All fields have JSON tags for the `--json` output mode.

## Petname Generator

Located in `pkg/random/petname.go`. A self-contained random name generator forked from `dustinkirkland/petname`. Contains embedded word lists (adjectives, adverbs, animal names) and a `GeneratePetName(words, separator)` function:

- 1 word: just an animal name (`falcon`)
- 2 words: adjective + name (`brave-falcon`)
- 3+ words: adverbs + adjective + name (`eagerly-brave-falcon`)

The default usage in the create command is `GeneratePetName(2, "-")`, producing names like `calm-firefly`. Uses `math/rand` for selection — not cryptographically secure, but that's fine for display names.
