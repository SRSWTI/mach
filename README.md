# mach

[![Go Report Card](https://goreportcard.com/badge/github.com/srswti/mach)](https://goreportcard.com/report/github.com/srswti/mach)

The simplest way to deploy web services. Period.

```bash
bodega init     # Create configuration
bodega deploy   # Deploy to production
```

That's it. SSL certificates, reverse proxy, load balancing, compression — all handled automatically.

## What is bodega mach?

it is a deployment tool that makes self-hosting as easy as:

1. **Write a simple TOML config**
2. **Run `bodega deploy`**
3. **Your site is live with HTTPS**

No nginx configs. No Certbot commands. No systemd units. Just deploy.

## Installation

### Quick Install (Recommended)

```bash
# Using the install script
curl -fsSL https://raw.githubusercontent.com/srswti/mach/main/install.sh | bash
```

### Manual Installation

**Install mach (Go binary):**
```bash
# Using go install
go install github.com/srswti/mach@latest

# Or clone and build
git clone https://github.com/srswti/mach
cd mach
./build.sh
```

**Install bodega (Python CLI):**
```bash
cd bodega
pip install -e .
# Or from PyPI: pip install bodega-deploy
```

## Quick Start

### 1. Initialize your project

```bash
bodega init
```

This creates a `bodega.toml` file:

```toml
[service]
name = "my-app"
port = 8000
domain = "api.example.com"
command = "python main.py"
working_dir = "."
auto_tls = true
```

### 2. Deploy

```bash
bodega deploy
```

Your service is now live at `https://api.example.com` with automatic SSL.

## Features

### Core Features

- **Automatic HTTPS** - SSL certificates via Let's Encrypt
- **Reverse Proxy** - Multiple upstreams with load balancing
- **Static File Serving** - Host documentation or SPAs
- **Redirects** - Domain migrations made easy

### Performance

- **Compression** - gzip + zstd automatic compression
- **HTTP/3** - Modern protocol support
- **WebSocket** - Full WebSocket proxy support
- **Load Balancing** - Round-robin, least-connections, IP hash

### Security

- **Basic Auth** - Password protection
- **Header Manipulation** - Add/remove/modify any headers
- **Security Headers** - XSS, CSP, HSTS out of the box
- **Request Buffering** - DDoS protection

### Observability

- **Health Checks** - Automatic backend monitoring
- **Access Logging** - JSON or common log format
- **Status Dashboard** - View all deployments

## Commands

```bash
bodega init                  # Create bodega.toml interactively
bodega deploy                # Deploy from bodega.toml
bodega deploy-static         # Deploy static files
bodega deploy-redirect       # Deploy a redirect
bodega status                # Show all deployments
bodega logs [service]        # View service logs
bodega remove [name]         # Remove a deployment
bodega certify               # Setup SSL certificates
bodega doctor                # Check system health
```

## Configuration Examples

### Load Balanced API

```toml
[service]
name = "api"
domain = "api.example.com"
port = 8080

[[service.upstreams]]
address = "localhost:8080"

[[service.upstreams]]
address = "localhost:8081"

[[service.upstreams]]
address = "localhost:8082"

load_balance = "round_robin"

[service.health_check]
path = "/health"
interval = "10s"
```

### Static Website

```toml
[service]
name = "docs"
domain = "docs.example.com"
handler = "static"

[service.static]
root = "/var/www/docs"
browse = false
```

### Secure Admin Panel

```toml
[service]
name = "admin"
domain = "admin.example.com"
port = 3000

[[service.basic_auth]]
username = "admin"
hashed_password = "$2y$10$..."
realm = "Admin Area"

[[service.headers_down]]
name = "X-Frame-Options"
value = "DENY"
```

## Architecture

```
User runs: bodega deploy
         ↓
    bodega.toml (TOML)
         ↓
    bodega CLI (Python) - prompts, fuzzy search
         ↓
    HTTP API → localhost:8765
         ↓
    mach daemon (Go) - reverse proxy, TLS, static files
         ↓
    Live HTTPS website
```

## Full Feature List


| Feature         | Description                   | Status |
| --------------- | ----------------------------- | ------ |
| Reverse Proxy   | Forward to backend services   | ✅      |
| Load Balancing  | Multiple backends             | ✅      |
| Health Checks   | Automatic backend monitoring  | ✅      |
| WebSocket       | Full WebSocket support        | ✅      |
| Compression     | gzip + zstd                   | ✅      |
| Static Files    | Serve static content          | ✅      |
| Redirects       | HTTP redirects                | ✅      |
| Basic Auth      | Password protection           | ✅      |
| Headers         | Request/response manipulation | ✅      |
| Buffering       | Request/response buffering    | ✅      |
| Timeouts        | Connection timeouts           | ✅      |
| Error Pages     | Custom error handling         | ✅      |
| Logging         | Access logs                   | ✅      |
| Automatic HTTPS | Let's Encrypt integration     | ✅      |
| HTTP/3          | Modern protocol               | ✅      |


## Project Structure

```
mach-deploy/
├── mach/              # Go daemon (HTTP engine)
├── bodega/            # Python CLI (user interface)
├── examples/          # Example configurations
├── install.sh         # One-line installer
└── README.md          # This file
```

## Development

```bash
# Build mach
cd mach
./build.sh

# Install bodega
cd ../bodega
pip install -e .

# Run tests
./test.sh
```

## License

Apache 2.0 - See [LICENSE](LICENSE) file.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing`)
3. Commit your changes (`git commit -am 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing`)
5. Open a Pull Request

## Support

- Issues: [GitHub Issues](https://github.com/srswti/mach/issues)
- Discussions: [GitHub Discussions](https://github.com/srswti/mach/discussions)

