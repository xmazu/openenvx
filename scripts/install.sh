#!/bin/bash
# OpenEnvX Installer
# Installation: /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/xmazu/openenvx/main/scripts/install.sh)"

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="xmazu/openenvx"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="openenvx"

print_info() {
    echo -e "${BLUE}ℹ${NC} $1" >&2
}

print_success() {
    echo -e "${GREEN}✓${NC} $1" >&2
}

print_error() {
    echo -e "${RED}✗${NC} $1" >&2
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1" >&2
}

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Linux*)     echo "linux";;
        Darwin*)    echo "darwin";;
        *)          echo "unsupported";;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64)     echo "amd64";;
        aarch64)    echo "arm64";;
        arm64)      echo "arm64";;
        *)          echo "unsupported";;
    esac
}

# Get the latest release version
get_latest_version() {
    version=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | head -1 | cut -d'"' -f4 2>/dev/null)

    if [ -z "$version" ]; then
        echo ""
        return 1
    fi

    version="${version#v}"
    echo "$version"
}

# Main installation
main() {
    print_info "Installing OpenEnvX"
    echo "" >&2

    OS=$(detect_os)
    ARCH=$(detect_arch)

    if [ "$OS" = "unsupported" ] || [ "$ARCH" = "unsupported" ]; then
        print_error "Unsupported OS/Architecture: $(uname -s) $(uname -m)"
        exit 1
    fi

    print_success "Detected: $(uname -s) $(uname -m) ($OS/$ARCH)"

    print_info "Fetching latest version..."
    VERSION=$(get_latest_version)
    if [ -z "$VERSION" ]; then
        print_error "Could not fetch latest version from GitHub"
        exit 1
    fi
    print_success "Latest version: v${VERSION}"

    BINARY_FILE="${BINARY_NAME}-${OS}-${ARCH}"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${VERSION}/${BINARY_FILE}"

    if [ -x "${INSTALL_DIR}/${BINARY_NAME}" ]; then
        INSTALLED_VERSION=$("${INSTALL_DIR}/${BINARY_NAME}" --version 2>/dev/null | sed -nE 's/.*v([0-9.]+).*/\1/p' || echo "unknown")
        if [ "$INSTALLED_VERSION" = "$VERSION" ]; then
            print_success "OpenEnvX v${VERSION} is already installed at ${INSTALL_DIR}/${BINARY_NAME}"
            echo "" >&2
            print_info "Next steps:"
            echo "  1. Initialize: openenvx init" >&2
            echo "  2. Store a secret: openenvx set DATABASE_URL" >&2
            echo "  3. Run with decrypted env: openenvx run -- npm start" >&2
            exit 0
        fi
    fi

    print_info "Downloading openenvx v${VERSION}..."

    TEMP_FILE=$(mktemp)
    if ! curl -fsSL -o "$TEMP_FILE" "$DOWNLOAD_URL"; then
        print_error "Failed to download from $DOWNLOAD_URL"
        rm -f "$TEMP_FILE"
        exit 1
    fi

    if [ ! -s "$TEMP_FILE" ]; then
        print_error "Downloaded file is empty"
        rm -f "$TEMP_FILE"
        exit 1
    fi

    print_success "Downloaded successfully"

    if [ ! -w "$INSTALL_DIR" ]; then
        print_info "Need sudo to write to $INSTALL_DIR"

        if ! sudo -v 2>/dev/null; then
            print_error "sudo access required but not available"
            rm -f "$TEMP_FILE"
            exit 1
        fi

        sudo mv "$TEMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
        sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    else
        mv "$TEMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
        chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    fi

    print_success "Installed openenvx v${VERSION} to ${INSTALL_DIR}/${BINARY_NAME}"

    echo "" >&2
    print_info "Installation complete!"
    echo "" >&2
    print_info "Next steps:"
    echo "  1. Initialize: openenvx init" >&2
    echo "  2. Store a secret: openenvx set DATABASE_URL" >&2
    echo "  3. Run with decrypted env: openenvx run -- npm start" >&2
    echo "" >&2
    echo "For more info: openenvx help" >&2
}

main "$@"
