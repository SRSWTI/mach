# bodega-deploy

**The simplest way to deploy anything.**

bodega wraps the entire "service file → nginx config → SSL cert" pipeline into one command.

## What is bodega?

```bash
# Before (manual process):
# 1. Write systemd service file
# 2. Write nginx config
# 3. Run certbot
# 4. Pray it works

# After (bodega):
bodega init   # 4 questions, done
bodega deploy # live on the internet with HTTPS
```

## Installation

```bash
pip install bodega-deploy
```

## Quick Start

```bash
# 1. Initialize your deployment
bodega init
# → Service name: my-app
# → Port: 8000
# → Domain: api.example.com
# → Command: python main.py
# → Enable HTTPS? yes

# 2. Deploy
bodega deploy
# → Deployed my-app to https://api.example.com
```

## Configuration (bodega.toml)

```toml
[service]
name = "my-api"
command = "python main.py"
port = 8000
domain = "api.example.com"
working_dir = "/home/user/my-api"
auto_tls = true
```

## Commands

| Command | Description |
|---------|-------------|
| `bodega init` | Create bodega.toml interactively |
| `bodega deploy` | Deploy services from bodega.toml |
| `bodega status` | Check deployment status |
| `bodega logs [service]` | View service logs |
| `bodega remove <service>` | Remove a deployed service |
| `bodega doctor` | Check if everything is healthy |

## Features

- **Automatic HTTPS** - Let's Encrypt certificates obtained automatically
- **Reverse Proxy** - Maps domains to your app ports
- **Service Management** - Process supervision built-in
- **Fuzzy Commands** - `bodega depoly` → "Did you mean: bodega deploy?"
- **Laptop Hosting** - Works from your laptop with `--tunnel` (coming soon)

## How it Works

```
bodega.toml → bodega CLI → mach daemon → Live HTTPS website
                     ↑
            (this Python package)  (Go binary)
```

## License

Apache 2.0 - See LICENSE file.
