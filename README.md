# bd-tracker Plugin for Genie

This plugin provides integration with the [bd/beads](https://github.com/okayest-dev/bd) issue tracker for Genie.

## Installation

```bash
# From source
git clone https://github.com/genie-plugins/bd-tracker
cd bd-tracker
go build -o bd-tracker ./cmd/bd-tracker
# Copy bd-tracker to your genie plugins directory or add to PATH
```

Or install via your package manager when available.

## Configuration

Add to your `config.toml`:

```toml
[plugins]
enable = ["bd-tracker"]
```

Or enable automatically if it's the only tracker plugin:

```toml
[plugins]
# bd-tracker will be loaded automatically if no other tracker is configured
```

## Requirements

- `bd` CLI installed and available in PATH (or set `BD_TRACKER_BINARY` environment variable)
- A bd workspace initialized (`bd init`)

## Features

Implements the full `TrackerSource` interface:

- `tracker/frontier` - Get next ready task (`bd ready`)
- `tracker/claim` - Claim a task (`bd update <id> --claim`)
- `tracker/read` - Read task details (`bd show <id>`)
- `tracker/close` - Close a task (`bd close <id> --reason`)
- `tracker/comment` - Add comment (`bd comment <id> --message`)
- `tracker/create` - Create new task (`bd create --title ...`)

## Environment Variables

- `BD_TRACKER_BINARY` - Path to bd binary (default: "bd")

## Protocol

This plugin implements the Genie plugin protocol v1 with JSON-RPC 2.0 over stdin/stdout.

### Methods

| Method | Description |
|--------|-------------|
| `capabilities/list` | Returns plugin capabilities |
| `tracker/frontier` | Get next ready task |
| `tracker/claim` | Claim a task by ID |
| `tracker/read` | Read task by ID |
| `tracker/close` | Close task with reason |
| `tracker/comment` | Add comment to task |
| `tracker/create` | Create new task |
| `ping` | Health check |
| `shutdown` | Graceful shutdown |

## Development

```bash
# Build
go build -o bd-tracker ./cmd/bd-tracker

# Test capabilities
./bd-tracker --capabilities

# Test manifest
./bd-tracker --manifest

# Run manually (JSON-RPC on stdin/stdout)
echo '{"jsonrpc":"2.0","method":"tracker/frontier","id":1}' | ./bd-tracker
```