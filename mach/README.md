# mach

[![Go Report Card](https://goreportcard.com/badge/github.com/srswti/mach)](https://goreportcard.com/report/github.com/srswti/mach)

mach is the high-performance deployment engine. It provides reverse proxy, automatic TLS, and HTTP serving capabilities through a clean HTTP API.

## Installation

```bash
# Using go install
go install github.com/srswti/mach@latest

# Or download pre-built binary
curl -fsSL https://github.com/srswti/mach/releases/latest/download/mach-$(uname -s)-$(uname -m) -o /usr/local/bin/mach
chmod +x /usr/local/bin/mach
```

## What is mach?

mach is a single binary that exposes a JSON HTTP API for managing deployments. The bodega CLI handles everything - you don't need to know the internals.

## Architecture

```
User runs: bodega deploy
           ↓
bodega.toml (human-friendly)
           ↓
bodega CLI (Python) - fuzzy search, interactive prompts
           ↓
mach HTTP API (this binary, Go)
           ↓
Reverse proxy + auto-TLS engine
           ↓
Your app on port 8000 now live with HTTPS
```

## API Endpoints


| Endpoint  | Method | Description             |
| --------- | ------ | ----------------------- |
| `/health` | GET    | Health check            |
| `/status` | GET    | Get deployment status   |
| `/list`   | GET    | List all services       |
| `/deploy` | POST   | Deploy a new service    |
| `/remove` | DELETE | Remove a service        |
| `/reload` | POST   | Reload config from disk |


### Deploy Example

```bash
curl -X POST http://localhost:8765/deploy \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-api",
    "command": "python main.py",
    "port": 8000,
    "domain": "api.example.com",
    "working_dir": "/home/user/my-api",
    "auto_tls": true
  }'
```

## Configuration

mach stores its configuration in `~/.config/mach/config.json` (platform-specific).

Environment variables:

- `MACH_CONFIG` - Path to config file
- `MACH_ADDR` - Listen address (default: `localhost:8765`)

## Building

```bash
./build.sh
```

This builds for the current platform. Cross-compilation available for:

- darwin/amd64, darwin/arm64
- linux/amd64, linux/arm64

## License

Apache 2.0. See LICENSE file.

## Third Party Notices

This software uses components from Caddy (Apache 2.0 licensed). See LICENSES/CADDY_LICENSE for details.