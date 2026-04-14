"""HTTP client for the mach daemon."""

# TODO: Future - Client enhancements:
# TODO: Add retry logic with exponential backoff for 5xx errors
# TODO: Implement connection pooling for better performance
# TODO: Add async support (aiohttp) for concurrent operations
# TODO: Implement WebSocket client for real-time log streaming
# TODO: Add request/response interceptors for auth tokens
# TODO: Implement request caching for list/status calls
# TODO: Add connection health checking and automatic reconnection

import json
import requests
from typing import List, Dict, Any, Optional


class MachClient:
    """Client for the mach HTTP API."""

    def __init__(self, base_url: str = "http://localhost:8765"):
        self.base_url = base_url.rstrip("/")

    def health(self) -> bool:
        """Check if mach is running."""
        try:
            resp = requests.get(f"{self.base_url}/health", timeout=2)
            return resp.status_code == 200
        except requests.RequestException:
            return False

    def status(self) -> Dict[str, Any]:
        """Get deployment status."""
        resp = requests.get(f"{self.base_url}/status")
        resp.raise_for_status()
        return resp.json()

    def list_services(self) -> List[Dict[str, Any]]:
        """List all deployed services."""
        resp = requests.get(f"{self.base_url}/list")
        resp.raise_for_status()
        return resp.json()

    def get_config(self) -> Dict[str, Any]:
        """Get full mach configuration."""
        resp = requests.get(f"{self.base_url}/config")
        resp.raise_for_status()
        return resp.json()

    def update_config(self, config: Dict[str, Any]) -> Dict[str, str]:
        """Update mach configuration."""
        resp = requests.put(f"{self.base_url}/config", json=config)
        resp.raise_for_status()
        return resp.json()

    def deploy(
        self,
        name: str,
        domain: str,
        port: Optional[int] = None,
        domains: Optional[List[str]] = None,
        command: str = "",
        working_dir: str = "",
        env: Optional[Dict[str, str]] = None,
        auto_tls: bool = True,
        handler: str = "reverse_proxy",
        # Reverse proxy settings
        upstreams: Optional[List[Dict[str, Any]]] = None,
        load_balance: str = "",
        health_check: Optional[Dict[str, str]] = None,
        websocket: bool = False,
        # Compression
        compression: Optional[Dict[str, Any]] = None,
        # Headers
        headers_up: Optional[List[Dict[str, Any]]] = None,
        headers_down: Optional[List[Dict[str, Any]]] = None,
        # Authentication
        basic_auth: Optional[List[Dict[str, str]]] = None,
        # Static files
        static: Optional[Dict[str, Any]] = None,
        # Redirect
        redirect_to: str = "",
        redirect_code: int = 302,
        # Buffering
        buffer_requests: bool = False,
        buffer_responses: bool = False,
        # Timeouts
        read_timeout: str = "",
        write_timeout: str = "",
        # Error pages
        error_pages: Optional[Dict[str, str]] = None,
        # Logging
        logging: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, Any]:
        """Deploy a service with ALL features."""
        payload: Dict[str, Any] = {
            "name": name,
            "domain": domain,
            "auto_tls": auto_tls,
            "handler": handler,
        }

        if port:
            payload["port"] = port
        if domains:
            payload["domains"] = domains
        if command:
            payload["command"] = command
        if working_dir:
            payload["working_dir"] = working_dir
        if env:
            payload["env"] = env

        # Reverse proxy settings
        if upstreams:
            payload["upstreams"] = upstreams
        if load_balance:
            payload["load_balance"] = load_balance
        if health_check:
            payload["health_check"] = health_check
        if websocket:
            payload["websocket"] = True

        # Compression
        if compression:
            payload["compression"] = compression

        # Headers
        if headers_up:
            payload["headers_up"] = headers_up
        if headers_down:
            payload["headers_down"] = headers_down

        # Authentication
        if basic_auth:
            payload["basic_auth"] = basic_auth

        # Static files
        if static:
            payload["static"] = static

        # Redirect
        if redirect_to:
            payload["redirect_to"] = redirect_to
            payload["redirect_code"] = redirect_code

        # Buffering
        if buffer_requests:
            payload["buffer_requests"] = True
        if buffer_responses:
            payload["buffer_responses"] = True

        # Timeouts
        if read_timeout:
            payload["read_timeout"] = read_timeout
        if write_timeout:
            payload["write_timeout"] = write_timeout

        # Error pages
        if error_pages:
            payload["error_pages"] = error_pages

        # Logging
        if logging:
            payload["logging"] = logging

        resp = requests.post(f"{self.base_url}/deploy", json=payload)
        resp.raise_for_status()
        return resp.json()

    def remove(self, name: str) -> Dict[str, str]:
        """Remove a deployed service."""
        resp = requests.delete(f"{self.base_url}/remove?name={name}")
        resp.raise_for_status()
        return resp.json()

    def reload(self) -> Dict[str, str]:
        """Reload configuration from disk."""
        resp = requests.post(f"{self.base_url}/reload")
        resp.raise_for_status()
        return resp.json()
