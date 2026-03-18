---
title: Infrastructure
summary: DigitalOcean client, petname generation, and type definitions.
covers:
  - internal/do/**
  - pkg/**
  - types/**
scanned_at_commit: 98491e9d01c27506d1add05ed4b935117de7ecf0
layer: core
order: 2
---

## Overview

The infrastructure layer is intentionally thin. There's no provider abstraction or interface — the [CLI commands](./commands.md) receive a `*do.Client` directly and call its methods. This is a deliberate simplification from the previous multi-provider architecture; with only DigitalOcean supported, an abstraction layer would add complexity without value.

## DigitalOcean Client

Defined in `internal/do/client.go`. Wraps a `godo.Client` (v1.178.0) with labctl-specific operations.

### Construction

`NewClient(ctx, cfg, logger)` creates the client from config. If no tag name is configured, it defaults to `"labctl"`. The godo client is created via `godo.NewFromToken(cfg.APIToken)`.

### Operations

**`CreateDroplet(name, region, size)`** — creates a droplet with the configured base image and SSH key, tags it, enables monitoring, and assigns it to the configured project. Region and size aliases (like `eu` and `xs`) are resolved to DO slugs by `mapRegion` and `mapSize` before the API call. Returns a `types.Droplet` with the ID and creation parameters (no IP yet — that comes after polling).

**`ListDroplets()`** — paginated listing of all droplets with the configured tag (up to 200 per page). For each droplet, extracts the public IPv4 address (falls back to `"-"` if unavailable).

**`RemoveDroplet(id)`** — deletes a single droplet by ID.

**`CheckAPI()`** — calls `Account.Get()` to verify the API token is valid. Used by the [status command](./commands.md#status).

**`WaitUntilReady(droplet)`** — polls until the droplet is active, has a public IPv4, and accepts TCP connections on port 22 (2-second dial timeout). Uses a ticker at `config.PollInterval` with a deadline at `config.PollTimeout`. Updates the droplet's `IPv4` field in place when ready.

### Region and Size Mapping

Two private functions at the bottom of `client.go` map CLI aliases to DigitalOcean slugs:

- Regions: `usw`->`sfo3`, `use`->`nyc1`, `eu`->`fra1`, `ap`->`sgp1`, `au`->`syd1`
- Sizes: `xs`->`s-1vcpu-512mb-10gb`, `s`->`s-1vcpu-1gb`, `m`->`s-2vcpu-4gb`, `l`->`s-4vcpu-8gb`, `xl`->`s-8vcpu-16gb`

If an unrecognized value is passed, both functions return it as-is — allowing direct DO slugs as a fallback.

## Types

Defined in `types/droplet.go`. Two things live here:

- `VERSION` constant (`"0.8.0"`) — the application version, referenced by the version flag and status command
- `Droplet` struct — represents a DigitalOcean droplet with fields: `ID`, `Name`, `IPv4`, `Region`, `Size`. All fields have JSON tags for the `--json` output mode.

## Petname Generator

Located in `pkg/random/petname.go`. A self-contained random name generator forked from `dustinkirkland/petname`. Contains embedded word lists (adjectives, adverbs, animal names) and a `GeneratePetName(words, separator)` function:

- 1 word: just an animal name (`falcon`)
- 2 words: adjective + name (`brave-falcon`)
- 3+ words: adverbs + adjective + name (`eagerly-brave-falcon`)

The default usage in the create command is `GeneratePetName(2, "-")`, producing names like `calm-firefly`. Uses `math/rand` for selection — not cryptographically secure, but that's fine for display names.
