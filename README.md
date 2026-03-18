# labctl

Manage DigitalOcean droplets from the command line. Spin up ephemeral lab environments, SSH in, tear them down when you're done.

## Configuration

Create `~/.labctl/config.yml`:

```yaml
apiToken: your-digitalocean-api-token
projectID: your-digitalocean-project-id
sshKeyFingerprint: your-ssh-key-fingerprint
baseImage: ubuntu-22-04-x64
pollInterval: 5s
pollTimeout: 2m
tagName: labctl
```

| Property | Description |
|----------|-------------|
| `apiToken` | DigitalOcean API token |
| `projectID` | Project to assign droplets to |
| `sshKeyFingerprint` | Fingerprint of an SSH key registered in your DO account |
| `baseImage` | Image slug for new droplets |
| `pollInterval` | How often to check droplet readiness (e.g. `5s`) |
| `pollTimeout` | Max wait time for readiness (e.g. `2m`) |
| `tagName` | Tag for labctl-managed droplets (default: `labctl`) |

Override the config path with `LABCTL_CONFIG`:

```
export LABCTL_CONFIG=/path/to/config.yml
```

## Usage

Create a droplet:

```
labctl create
```

Options: `-r` region (`usw`, `use`, `eu`, `ap`, `au`), `-s` size (`xs`, `s`, `m`, `l`, `xl`), `-n` name.

The SSH command is copied to your clipboard and printed to stdout.

List droplets:

```
labctl ls
```

Remove droplets:

```
labctl rm <id> [<id>...]
```

Check version and API connectivity:

```
labctl status
```

Add `--json` to any command for JSON output.

## Building

```
make build
```
