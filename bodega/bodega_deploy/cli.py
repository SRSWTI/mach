"""bodega CLI - The simplest way to deploy anything."""

# TODO: Future CLI enhancements:
# TODO: Add interactive dashboard (bodega dashboard) - Real-time metrics with rich/live
# TODO: Implement bodega tunnel - ngrok/Cloudflare Tunnel integration for local hosting
# TODO: Add bodega backup - Backup and restore configurations
# TODO: Implement bodega migrate - Zero-downtime service migrations
# TODO: Add bodega secrets - HashiCorp Vault/AWS Secrets integration
# TODO: Implement bodega test - Pre-deployment config validation
# TODO: Add bodega clone - Copy service config from existing deployment
# TODO: Implement bodega compose - Docker Compose file import
# TODO: Add bodega rollback - One-click rollback to previous version
# TODO: Implement bodega scale - Horizontal scaling across multiple servers
# TODO: Add bodega metrics - Built-in Prometheus metrics export

import os
import sys
import json
import subprocess
import platform
from pathlib import Path
from typing import Optional, List
from dataclasses import dataclass
from difflib import get_close_matches

import click
from rich.console import Console
from rich.table import Table
from rich.panel import Panel
from rich.prompt import Prompt, Confirm, IntPrompt

from .mach_client import MachClient

console = Console()

MACH_VERSION = "0.1.0"


@dataclass
class ServiceConfig:
    name: str
    command: str
    port: int
    domain: str
    working_dir: str
    env: dict
    auto_tls: bool


def fuzzy_command(user_input: str, valid_commands: List[str]) -> Optional[str]:
    """Find the closest matching command."""
    matches = get_close_matches(user_input.lower(), valid_commands, n=1, cutoff=0.4)
    return matches[0] if matches else None


def print_suggestions(user_input: str):
    """Print command suggestions for typos."""
    valid_commands = ["deploy", "certify", "status", "logs", "remove", "init", "help", "doctor"]
    suggestion = fuzzy_command(user_input, valid_commands)
    if suggestion:
        console.print(f"[yellow]Did you mean: [bold]bodega {suggestion}[/bold]?[/yellow]")


def ensure_mach_running() -> bool:
    """Check if mach daemon is running, return True if yes."""
    client = MachClient()
    if client.health():
        return True

    console.print("[yellow]mach daemon not running. Starting it now...[/yellow]")

    # Start mach in background
    mach_bin = get_mach_binary_path()
    if not mach_bin.exists():
        console.print(f"[red]mach binary not found at {mach_bin}[/red]")
        console.print("[yellow]Run: bodega install-mach[/yellow]")
        return False

    try:
        subprocess.Popen(
            [str(mach_bin)],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            start_new_session=True,
        )
        import time
        time.sleep(1)
        return client.health()
    except Exception as e:
        console.print(f"[red]Failed to start mach: {e}[/red]")
        return False


def get_mach_binary_path() -> Path:
    """Get the path to the mach binary."""
    from shutil import which
    mach_path = which("mach")
    if mach_path:
        return Path(mach_path)

    system = platform.system().lower()
    machine = platform.machine().lower()

    if system == "darwin":
        plat = f"darwin-{machine}"
    elif system == "linux":
        plat = f"linux-{machine}"
    else:
        plat = f"{system}-{machine}"

    pkg_dir = Path(__file__).parent
    bundled = pkg_dir / "bin" / f"mach-{plat}"
    if bundled.exists():
        return bundled

    home_bin = Path.home() / ".bodega" / "mach"
    if home_bin.exists():
        return home_bin

    return bundled


def load_bodega_toml(path: Path) -> dict:
    """Load bodega.toml configuration."""
    try:
        import tomllib
    except ImportError:
        import tomli as tomllib

    with open(path, "rb") as f:
        return tomllib.load(f)


def save_bodega_toml(path: Path, config: dict):
    """Save bodega.toml configuration."""
    import tomli_w

    with open(path, "wb") as f:
        tomli_w.dump(config, f)


@click.group(invoke_without_command=True)
@click.option("--version", is_flag=True, help="Show version")
@click.pass_context
def cli(ctx, version):
    """bodega - The simplest way to deploy anything.

    Examples:
        bodega init                # Create a new bodega.toml
        bodega deploy              # Deploy services from bodega.toml
        bodega deploy-static       # Deploy static file server
        bodega status              # Check deployment status
        bodega logs myapp          # View logs for a service
        bodega doctor              # Check system health
    """
    if version:
        console.print(f"bodega-deploy v{MACH_VERSION}")
        return

    if ctx.invoked_subcommand is None:
        console.print(cli.get_help(ctx))


@cli.command()
@click.option("--advanced", is_flag=True, help="Show advanced options")
def init(advanced):
    """Initialize a new bodega.toml configuration."""
    console.print(Panel.fit(
        "[bold blue]Welcome to bodega![/bold blue]\n"
        "Let's set up your deployment configuration."
    ))

    if Path("bodega.toml").exists():
        if not Confirm.ask("bodega.toml already exists. Overwrite?"):
            return

    # Basic prompts
    name = Prompt.ask("Service name", default="my-app")
    port = IntPrompt.ask("Port your app runs on", default=8000)
    domain = Prompt.ask("Domain name (e.g., api.example.com)")
    command = Prompt.ask("Command to start your app", default="python main.py")
    working_dir = Prompt.ask("Working directory", default=os.getcwd())
    auto_tls = Confirm.ask("Enable automatic HTTPS?", default=True)

    config = {
        "service": {
            "name": name,
            "port": port,
            "domain": domain,
            "command": command,
            "working_dir": working_dir,
            "auto_tls": auto_tls,
            "handler": "reverse_proxy",
        }
    }

    if advanced:
        console.print("\n[bold]Advanced Options:[/bold]")
        
        # Additional domains
        if Confirm.ask("Add additional domains? (e.g., www.example.com)", default=False):
            domains_str = Prompt.ask("Additional domains (comma-separated)")
            domains = [d.strip() for d in domains_str.split(",") if d.strip()]
            config["service"]["domains"] = domains

        # Load balancing
        if Confirm.ask("Enable load balancing?", default=False):
            lb_type = Prompt.ask("Load balance policy", default="round_robin")
            config["service"]["load_balance"] = lb_type
            upstreams_str = Prompt.ask("Additional upstreams (comma-separated host:port)")
            if upstreams_str:
                upstreams = [{"address": u.strip()} for u in upstreams_str.split(",") if u.strip()]
                config["service"]["upstreams"] = upstreams

        # Health checks
        if Confirm.ask("Enable health checks?", default=False):
            health_path = Prompt.ask("Health check path", default="/health")
            config["service"]["health_check"] = {
                "path": health_path,
                "interval": "10s"
            }

        # Compression
        if Confirm.ask("Enable compression (gzip/zstd)?", default=True):
            config["service"]["compression"] = {
                "enable": True,
                "formats": ["zstd", "gzip"]
            }

        # WebSocket support
        if Confirm.ask("Enable WebSocket support?", default=False):
            config["service"]["websocket"] = True

        # Headers
        if Confirm.ask("Add custom headers?", default=False):
            headers = []
            while Confirm.ask("Add another header?", default=True):
                header_name = Prompt.ask("Header name (e.g., X-Frame-Options)")
                header_value = Prompt.ask("Header value")
                headers.append({"name": header_name, "value": header_value})
            if headers:
                config["service"]["headers_up"] = headers

        # Authentication
        if Confirm.ask("Enable basic authentication?", default=False):
            username = Prompt.ask("Username")
            password = Prompt.ask("Password (will be hashed)")
            import subprocess
            try:
                result = subprocess.run(
                    ["htpasswd", "-nbB", username, password],
                    capture_output=True, text=True, check=True
                )
                hashed = result.stdout.strip().split(":")[1]
                config["service"]["basic_auth"] = [{
                    "username": username,
                    "hashed_password": hashed
                }]
            except:
                console.print("[yellow]Note: Install apache2-utils to hash passwords[/yellow]")

        # Buffering
        if Confirm.ask("Enable request/response buffering?", default=False):
            config["service"]["buffer_requests"] = True
            config["service"]["buffer_responses"] = True

        # Logging
        if Confirm.ask("Enable access logging?", default=False):
            config["service"]["logging"] = {
                "enable": True,
                "output": "stdout"
            }

    save_bodega_toml(Path("bodega.toml"), config)
    console.print(f"[green]Created bodega.toml![/green]")
    console.print(f"[dim]Run 'bodega deploy' to deploy {name}[/dim]")


@cli.command()
@click.option("--config", "-c", default="bodega.toml", help="Path to config file")
def deploy(config):
    """Deploy services from bodega.toml."""
    config_path = Path(config)
    if not config_path.exists():
        console.print(f"[red]Config file not found: {config}[/red]")
        console.print("[yellow]Run 'bodega init' to create one[/yellow]")
        sys.exit(1)

    if not ensure_mach_running():
        sys.exit(1)

    try:
        toml_config = load_bodega_toml(config_path)
    except Exception as e:
        console.print(f"[red]Failed to parse {config}: {e}[/red]")
        sys.exit(1)

    if "service" not in toml_config:
        console.print(f"[red]No [service] section found in {config}[/red]")
        sys.exit(1)

    svc = toml_config["service"]
    client = MachClient()

    with console.status("[bold green]Deploying...[/bold green]"):
        try:
            # Build kwargs from all possible fields
            kwargs = {
                "name": svc["name"],
                "domain": svc["domain"],
                "auto_tls": svc.get("auto_tls", True),
            }

            if "port" in svc:
                kwargs["port"] = svc["port"]
            if "domains" in svc:
                kwargs["domains"] = svc["domains"]
            if "command" in svc:
                kwargs["command"] = svc["command"]
            if "working_dir" in svc:
                kwargs["working_dir"] = svc["working_dir"]
            if "env" in svc:
                kwargs["env"] = svc["env"]
            if "handler" in svc:
                kwargs["handler"] = svc["handler"]
            if "upstreams" in svc:
                kwargs["upstreams"] = svc["upstreams"]
            if "load_balance" in svc:
                kwargs["load_balance"] = svc["load_balance"]
            if "health_check" in svc:
                kwargs["health_check"] = svc["health_check"]
            if "websocket" in svc:
                kwargs["websocket"] = svc["websocket"]
            if "compression" in svc:
                kwargs["compression"] = svc["compression"]
            if "headers_up" in svc:
                kwargs["headers_up"] = svc["headers_up"]
            if "headers_down" in svc:
                kwargs["headers_down"] = svc["headers_down"]
            if "basic_auth" in svc:
                kwargs["basic_auth"] = svc["basic_auth"]
            if "static" in svc:
                kwargs["static"] = svc["static"]
            if "redirect_to" in svc:
                kwargs["redirect_to"] = svc["redirect_to"]
                kwargs["redirect_code"] = svc.get("redirect_code", 302)
            if "buffer_requests" in svc:
                kwargs["buffer_requests"] = svc["buffer_requests"]
            if "buffer_responses" in svc:
                kwargs["buffer_responses"] = svc["buffer_responses"]
            if "read_timeout" in svc:
                kwargs["read_timeout"] = svc["read_timeout"]
            if "write_timeout" in svc:
                kwargs["write_timeout"] = svc["write_timeout"]
            if "error_pages" in svc:
                kwargs["error_pages"] = svc["error_pages"]
            if "logging" in svc:
                kwargs["logging"] = svc["logging"]

            result = client.deploy(**kwargs)
            console.print(f"[green]Deployed {result['service']} to {result['domain']}[/green]")
            
            if result.get("handler"):
                console.print(f"[dim]Handler: {result['handler']}[/dim]")
        except Exception as e:
            console.print(f"[red]Deployment failed: {e}[/red]")
            sys.exit(1)


@cli.command()
@click.option("--root", "-r", required=True, help="Root directory to serve")
@click.option("--domain", "-d", required=True, help="Domain name")
@click.option("--name", "-n", default="static-site", help="Service name")
@click.option("--browse", is_flag=True, help="Enable directory browsing")
@click.option("--auto-tls", is_flag=True, default=True, help="Enable HTTPS")
@click.option("--compress", is_flag=True, default=True, help="Enable compression")
def deploy_static(root, domain, name, browse, auto_tls, compress):
    """Deploy a static file server."""
    if not ensure_mach_running():
        sys.exit(1)

    if not Path(root).exists():
        console.print(f"[red]Directory not found: {root}[/red]")
        sys.exit(1)

    client = MachClient()

    with console.status("[bold green]Deploying static site...[/bold green]"):
        try:
            kwargs = {
                "name": name,
                "domain": domain,
                "handler": "static",
                "auto_tls": auto_tls,
                "static": {
                    "root": str(Path(root).absolute()),
                    "browse": browse,
                }
            }

            if compress:
                kwargs["compression"] = {
                    "enable": True,
                    "formats": ["zstd", "gzip"]
                }

            result = client.deploy(**kwargs)
            console.print(f"[green]Deployed static site {result['service']} to {result['domain']}[/green]")
        except Exception as e:
            console.print(f"[red]Deployment failed: {e}[/red]")
            sys.exit(1)


@cli.command()
@click.option("--from-domain", "-f", required=True, help="Source domain")
@click.option("--to-url", "-t", required=True, help="Target URL to redirect to")
@click.option("--code", "-c", default=302, help="HTTP redirect code (301 or 302)")
@click.option("--name", "-n", default="redirect", help="Service name")
@click.option("--auto-tls", is_flag=True, default=True, help="Enable HTTPS")
def deploy_redirect(from_domain, to_url, code, name, auto_tls):
    """Deploy a redirect service."""
    if not ensure_mach_running():
        sys.exit(1)

    client = MachClient()

    with console.status("[bold green]Deploying redirect...[/bold green]"):
        try:
            result = client.deploy(
                name=name,
                domain=from_domain,
                handler="redirect",
                redirect_to=to_url,
                redirect_code=code,
                auto_tls=auto_tls,
            )
            console.print(f"[green]Deployed redirect {result['service']}: {from_domain} → {to_url}[/green]")
        except Exception as e:
            console.print(f"[red]Deployment failed: {e}[/red]")
            sys.exit(1)


@cli.command()
def status():
    """Check deployment status."""
    if not ensure_mach_running():
        sys.exit(1)

    client = MachClient()

    try:
        mach_status = client.status()
        services = client.list_services()
    except Exception as e:
        console.print(f"[red]Failed to get status: {e}[/red]")
        sys.exit(1)

    table = Table(title="bodega Status")
    table.add_column("Service", style="cyan")
    table.add_column("Domain", style="green")
    table.add_column("Handler", style="magenta")
    table.add_column("HTTPS", style="blue")

    for svc in services:
        handler = svc.get("handler", "reverse_proxy")
        table.add_row(
            svc["name"],
            svc["domain"],
            handler,
            "✓" if svc.get("auto_tls") else "✗",
        )

    console.print(table)
    console.print(f"[dim]Uptime: {mach_status.get('uptime', 'unknown')}[/dim]")


@cli.command()
@click.argument("name")
def remove(name):
    """Remove a deployed service."""
    if not ensure_mach_running():
        sys.exit(1)

    client = MachClient()

    with console.status(f"[bold yellow]Removing {name}...[/bold yellow]"):
        try:
            result = client.remove(name)
            console.print(f"[green]Removed {result['service']}[/green]")
        except Exception as e:
            console.print(f"[red]Failed to remove: {e}[/red]")
            sys.exit(1)


@cli.command()
def certify():
    """Setup SSL certificates for all services."""
    if not ensure_mach_running():
        sys.exit(1)

    console.print("[green]SSL certificates are managed automatically[/green]")
    console.print("[dim]Certificates obtained when deploying with auto_tls = true[/dim]")


@cli.command()
@click.argument("name", required=False)
def logs(name):
    """View logs for a service (or all services if no name given)."""
    if not ensure_mach_running():
        sys.exit(1)

    if name:
        console.print(f"[dim]Showing logs for {name}...[/dim]")
        console.print("[yellow]Log streaming coming in v0.3![/yellow]")
    else:
        console.print("[dim]Showing logs for all services...[/dim]")
        console.print("[yellow]Log streaming coming in v0.3![/yellow]")


@cli.command()
def doctor():
    """Check if everything is healthy."""
    checks = []

    mach_bin = get_mach_binary_path()
    if mach_bin.exists():
        checks.append(("mach binary", "✓", "green"))
    else:
        checks.append(("mach binary", "✗", "red"))

    client = MachClient()
    if client.health():
        checks.append(("mach daemon", "✓ running", "green"))
    else:
        checks.append(("mach daemon", "✗ stopped", "yellow"))

    if Path("bodega.toml").exists():
        checks.append(("bodega.toml", "✓", "green"))
    else:
        checks.append(("bodega.toml", "✗ not found", "yellow"))

    table = Table(title="bodega Doctor")
    table.add_column("Check", style="cyan")
    table.add_column("Status")

    for check, status, color in checks:
        table.add_row(check, f"[{color}]{status}[/{color}]")

    console.print(table)


@cli.command()
def config():
    """View current mach configuration."""
    if not ensure_mach_running():
        sys.exit(1)

    client = MachClient()
    try:
        cfg = client.get_config()
        console.print(json.dumps(cfg, indent=2))
    except Exception as e:
        console.print(f"[red]Failed to get config: {e}[/red]")


def main():
    """Entry point with fuzzy command matching."""
    if len(sys.argv) > 1:
        cmd = sys.argv[1]
        valid_commands = [
            "deploy", "deploy-static", "deploy-redirect", "certify",
            "status", "logs", "remove", "init", "help", "doctor", "config",
            "--version", "-v"
        ]

        if cmd not in valid_commands and not cmd.startswith("-"):
            suggestion = fuzzy_command(cmd, valid_commands)
            if suggestion:
                console.print(f"[yellow]Unknown command: {cmd}[/yellow]")
                console.print(f"[green]Did you mean: [bold]bodega {suggestion}[/bold]?[/green]")
                sys.argv[1] = suggestion

    cli()


if __name__ == "__main__":
    main()
