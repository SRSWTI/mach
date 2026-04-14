# Bodega Deploy Examples

This directory contains example configurations demonstrating all features of bodega-deploy.

## Available Examples

### 1. `bodega-full.toml` - Full Production API
Complete configuration with all features enabled:
- Load balancing across 3 backends
- Health checks
- WebSocket support
- Compression (zstd + gzip)
- Header manipulation (security headers)
- Basic authentication
- Error pages
- Access logging

### 2. `bodega-static.toml` - Static File Server
Perfect for hosting documentation or SPA builds:
- Static file serving with precompressed assets
- Custom index files
- Hidden file patterns
- Caching headers

### 3. `bodega-redirect.toml` - Domain Redirects
Redirect old domains to new ones:
- 301/302 redirects
- Multiple source domains
- HTTPS enforcement

## Usage

Deploy any example:

```bash
# Deploy full API
bodega deploy --config examples/bodega-full.toml

# Deploy static site
bodega deploy --config examples/bodega-static.toml

# Deploy redirect
bodega deploy --config examples/bodega-redirect.toml
```

Or use the shorthand commands:

```bash
# For static sites
bodega deploy-static --root /var/www/site --domain site.example.com

# For redirects
bodega deploy-redirect --from-domain old.com --to-url https://new.com
```

## Feature Reference

| Feature | TOML Key | Values |
|---------|----------|--------|
| Load Balancing | `load_balance` | `round_robin`, `least_conn`, `ip_hash` |
| Compression | `compression.enable` | `true`/`false` |
| WebSocket | `websocket` | `true`/`false` |
| Headers Up | `headers_up` | Array of `{name, value}` |
| Headers Down | `headers_down` | Array of `{name, value}` |
| Basic Auth | `basic_auth` | Array of `{username, hashed_password, realm}` |
| Buffering | `buffer_requests` | `true`/`false` |
| Timeouts | `read_timeout` | Duration string (e.g., "30s") |
| Error Pages | `error_pages` | Map of code to file path |
| Logging | `logging.enable` | `true`/`false` |

## See Also

- [Main README](../README.md)
- [CLI Help](../bodega/README.md)
- [Mach API](../mach/README.md)
