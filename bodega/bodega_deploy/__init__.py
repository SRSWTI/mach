"""bodega-deploy - The simplest way to deploy anything."""

# TODO: Future - Module structure improvements:
# TODO: Split into submodules: config, deploy, server, utils
# TODO: Implement plugin system for custom deployment strategies
# TODO: Add abstract base classes for cloud providers
# TODO: Create separate package for CLI vs library usage
# TODO: Implement async API client for better performance
# TODO: Add telemetry module (opt-in) for usage analytics
# TODO: Implement config validation schema with pydantic
# TODO: Create built-in templates for common frameworks

__version__ = "0.1.0"
__all__ = ["deploy", "certify", "status", "logs", "remove"]
