#!/bin/bash
set -e

# bodega-deploy installation script
# Downloads and installs the bodega CLI and mach daemon
#
# TODO: Future - Installer improvements:
# TODO: Add support for Windows (PowerShell script)
# TODO: Implement checksum verification for downloaded binaries
# TODO: Add uninstall option (--uninstall flag)
# TODO: Support for system-wide installation (--system flag)
# TODO: Add update/upgrade functionality (--upgrade flag)
# TODO: Implement offline installation mode (bundled binaries)
# TODO: Add Docker-based installation option
# TODO: Support fish shell completions
# TODO: Add man page generation and installation
# TODO: Implement automatic PATH detection and fix suggestions

echo "╔═══════════════════════════════════════╗"
echo "║     mach installer             ║"
echo "║     The ulitmate server        ║"
echo "╚═══════════════════════════════════════╝"
echo ""

# Configuration
REPO="srswti/mach"
VERSION="${VERSION:-0.1.0}"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
MACH_DIR="${MACH_DIR:-$HOME/.bodega}"

# Detect platform
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$ARCH" in
        x86_64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            echo "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    PLATFORM="${OS}-${ARCH}"
}

# Check for required tools
check_dependencies() {
    echo "Checking dependencies..."

    if ! command -v python3 &> /dev/null; then
        echo "❌ Python 3 is required but not installed"
        exit 1
    fi
    echo "✓ Python 3 found"

    if ! command -v pip3 &> /dev/null && ! command -v pip &> /dev/null; then
        echo "❌ pip is required but not installed"
        exit 1
    fi
    echo "✓ pip found"
}

# Download mach binary
download_mach() {
    echo ""
    echo "Downloading mach daemon..."

    local mach_url="https://github.com/${REPO}/releases/download/v${VERSION}/mach-${PLATFORM}"
    local mach_bin="${MACH_DIR}/mach"

    mkdir -p "$MACH_DIR"

    if command -v curl &> /dev/null; then
        curl -fsSL "$mach_url" -o "$mach_bin" || {
            echo "⚠️  Could not download mach binary"
            echo "   You can build from source: cd mach && ./build.sh"
            return 1
        }
    elif command -v wget &> /dev/null; then
        wget -q "$mach_url" -O "$mach_bin" || {
            echo "⚠️  Could not download mach binary"
            echo "   You can build from source: cd mach && ./build.sh"
            return 1
        }
    else
        echo "❌ curl or wget is required"
        exit 1
    fi

    chmod +x "$mach_bin"
    echo "✓ mach downloaded to $mach_bin"
}

# Install Python package
install_bodega() {
    echo ""
    echo "Installing bodega CLI..."

    # Create virtual environment if requested
    if [ "${USE_VENV:-false}" = "true" ]; then
        VENV_DIR="${VENV_DIR:-$HOME/.bodega/venv}"
        python3 -m venv "$VENV_DIR"
        source "$VENV_DIR/bin/activate"
        echo "✓ Created virtual environment at $VENV_DIR"
    fi

    # Install bodega package
    pip install bodega-deploy || {
        echo "❌ Failed to install bodega-deploy"
        exit 1
    }

    echo "✓ bodega CLI installed"
}

# Setup shell integration
setup_shell() {
    echo ""
    echo "Setting up shell integration..."

    SHELL_RC=""
    if [ -n "$ZSH_VERSION" ]; then
        SHELL_RC="$HOME/.zshrc"
    elif [ -n "$BASH_VERSION" ]; then
        SHELL_RC="$HOME/.bashrc"
    fi

    if [ -n "$SHELL_RC" ] && [ -f "$SHELL_RC" ]; then
        # Add mach to PATH if not already there
        if ! grep -q "bodega" "$SHELL_RC" 2>/dev/null; then
            echo "" >> "$SHELL_RC"
            echo "# bodega-deploy" >> "$SHELL_RC"
            echo "export PATH=\"\$PATH:$MACH_DIR\"" >> "$SHELL_RC"
            echo "✓ Added $MACH_DIR to PATH in $SHELL_RC"
        fi
    fi

    # Create bodega completion script
    if command -v bodega &> /dev/null; then
        # Generate completions if click supports it
        _BODEGA_COMPLETE=bash_source bodega 2>/dev/null > "$MACH_DIR/bodega-complete.bash" || true
    fi
}

# Verify installation
verify_install() {
    echo ""
    echo "Verifying installation..."

    if command -v bodega &> /dev/null; then
        echo "✓ bodega CLI installed"
        bodega --version 2>/dev/null || true
    else
        echo "⚠️  bodega CLI not in PATH"
        echo "   You may need to restart your shell or run:"
        echo "   export PATH=\"\$PATH:$MACH_DIR\""
    fi

    if [ -f "$MACH_DIR/mach" ]; then
        echo "✓ mach daemon available"
    else
        echo "⚠️  mach daemon not found"
        echo "   Run: cd mach && ./build.sh to build from source"
    fi
}

# Print usage
print_usage() {
    echo ""
    echo "═══════════════════════════════════════════"
    echo "Installation complete! lfg"
    echo "═══════════════════════════════════════════"
    echo ""
    echo "Quick start:"
    echo "  cd my-project"
    echo "  bodega init       # Create bodega.toml"
    echo "  bodega deploy     # Deploy to the internet"
    echo ""
    echo "Commands:"
    echo "  bodega init       - Create new deployment config"
    echo "  bodega deploy     - Deploy services"
    echo "  bodega status     - Check deployment status"
    echo "  bodega logs       - View logs"
    echo "  bodega doctor     - Check health"
    echo ""
    echo "Documentation: https://github.com/${REPO}"
    echo ""
}

# Main installation flow
main() {
    # TODO: Future - Add command-line argument parsing
    # TODO: Add --verbose and --quiet modes
    # TODO: Implement --dry-run to preview changes
    # TODO: Add --version to pin specific release

    detect_platform
    check_dependencies

    echo "Platform: $PLATFORM"
    echo "Install dir: $INSTALL_DIR"
    echo "Mach dir: $MACH_DIR"
    echo ""

    # Download or build mach
    # TODO: Future - Support for air-gapped environments
    # TODO: Add mirror/CDN selection for faster downloads
    if ! download_mach; then
        echo ""
        echo "Building mach from source instead..."
        if [ -d "mach" ]; then
            cd mach
            ./build.sh
            cp build/mach-* "$MACH_DIR/mach"
            cd ..
        fi
    fi

    # Install Python package
    # TODO: Future - Support for conda/poetry package managers
    # TODO: Add editable vs production install detection
    install_bodega

    # Setup shell integration
    # TODO: Future - Add IDE integrations (VSCode, IntelliJ)
    setup_shell

    # Verify
    verify_install

    # Print usage
    print_usage
}

# Run main
main "$@"
