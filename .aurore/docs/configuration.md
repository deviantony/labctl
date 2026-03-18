---
title: Configuration
summary: YAML configuration file structure, properties, defaults, and environment override.
covers:
  - internal/config/**
  - config.example.yml
scanned_at_commit: 98491e9d01c27506d1add05ed4b935117de7ecf0
layer: core
order: 3
---

## Overview

labctl loads its configuration from a single YAML file. The config drives all DigitalOcean API interactions — there are no CLI flags for credentials or infrastructure settings. This keeps the CLI interface clean (flags are for per-invocation choices like region and size) while persistent settings live in the config file.

## File Location

The config file is loaded from `~/.labctl/config.yml` by default. This can be overridden by setting the `LABCTL_CONFIG` environment variable to an absolute path, which is useful for CI/automation or when running multiple labctl configurations side by side.

The config is loaded at startup in `cmd/labctl.go` and passed to the DigitalOcean client constructor. If the file is missing or malformed, labctl exits with a fatal error.

## Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `apiToken` | string | yes | — | DigitalOcean API token. Used to authenticate all API calls. |
| `projectID` | string | yes | — | DigitalOcean project ID. New droplets are assigned to this project for organizational grouping. |
| `sshKeyFingerprint` | string | yes | — | Fingerprint of the SSH key to inject into new droplets. Must match a key already registered in your DigitalOcean account. |
| `baseImage` | string | yes | — | Image slug for new droplets (e.g. `ubuntu-22-04-x64`). This is the OS image every droplet starts from. |
| `pollInterval` | duration | yes | — | How often to check if a newly created droplet is ready. Parsed as a Go duration (e.g. `5s`, `10s`). |
| `pollTimeout` | duration | yes | — | Maximum time to wait for a droplet to become ready before giving up. Parsed as a Go duration (e.g. `2m`, `5m`). |
| `tagName` | string | no | `labctl` | Tag applied to all droplets created by labctl. Used to scope `ls` to only show labctl-managed droplets. If omitted, defaults to `"labctl"` in the DO client constructor (`internal/do/client.go`). |

## Example

From `config.example.yml`:

```yaml
apiToken: your-digitalocean-api-token
projectID: your-digitalocean-project-id
sshKeyFingerprint: your-ssh-key-fingerprint
baseImage: ubuntu-22-04-x64
pollInterval: 5s
pollTimeout: 2m
tagName: labctl
```

## How Defaults Work

The config struct in `internal/config/config.go` does not apply defaults at load time — `NewConfig` simply decodes the YAML into the struct. The only default is `tagName`, which is applied in `internal/do/client.go` at client construction time: if `cfg.TagName` is empty, it's set to `"labctl"`. All other properties have no defaults and must be provided.

This means that if `pollInterval` or `pollTimeout` are omitted, they'll be zero-valued (`0s`), which would cause the readiness polling to either spin continuously or time out immediately. In practice, all properties except `tagName` should be treated as required.

## Duration Format

The `pollInterval` and `pollTimeout` fields use Go's `time.Duration` parsing via the `yaml.v3` decoder. Valid formats include:

- `5s` — 5 seconds
- `2m` — 2 minutes
- `1m30s` — 1 minute 30 seconds
- `500ms` — 500 milliseconds
