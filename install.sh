#!/usr/bin/env bash
set -e

# GitHub repo (owner/name). Override with OPENENVX_REPO env if needed.
GITHUB_REPO="${OPENENVX_REPO:-xmazu/openenvx}"
BINARY_NAME="openenvx"
VERSION="latest"
BINDIR=""

usage() {
  cat <<EOF
Usage: $0 [--version VERSION] [-d DIR]
Install OpenEnvX CLI from GitHub Releases (macOS and Linux only).

Options:
  --version VERSION   Release tag (e.g. v1.0.0) or 'latest' (default)
  -d, --dir DIR       Install directory (default: \$HOME/.local/bin)
  -h, --help          Show this help

Examples:
  $0
  $0 --version v1.0.0
  $0 -d /usr/local/bin
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      VERSION="${2:?Missing value for --version}"
      shift 2
      ;;
    -d|--dir)
      BINDIR="${2:?Missing value for -d}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

# Platform detection (macOS and Linux only)
OS=$(uname -s)
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac
case "$OS" in
  Darwin) OS="darwin" ;;
  Linux) OS="linux" ;;
  *)
    echo "Unsupported OS: $OS. Only macOS and Linux are supported." >&2
    exit 1
    ;;
esac

# Install directory
BINDIR="${BINDIR:-$HOME/.local/bin}"
mkdir -p "$BINDIR"

# Resolve release and asset URL
ASSET_NAME="${BINARY_NAME}-${OS}-${ARCH}"
if [[ "$VERSION" == "latest" ]]; then
  RELEASE_JSON=$(curl -sSfL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest") || {
    echo "Failed to fetch latest release. Check your network and that ${GITHUB_REPO} has releases." >&2
    exit 1
  }
else
  RELEASE_JSON=$(curl -sSfL "https://api.github.com/repos/${GITHUB_REPO}/releases/tags/${VERSION}") || {
    echo "Failed to fetch release ${VERSION}. Check that the tag exists." >&2
    exit 1
  }
fi

# Extract download URL for this platform (no jq required)
ASSET_URL=$(echo "$RELEASE_JSON" | grep -o '"browser_download_url": "[^"]*"' | sed 's/"browser_download_url": "\([^"]*\)"/\1/' | grep -F "$ASSET_NAME" | head -1)
if [[ -z "$ASSET_URL" ]]; then
  echo "No release or binary for this platform. Expected asset: ${ASSET_NAME}" >&2
  echo "Ensure a GitHub release exists with artifact ${ASSET_NAME}." >&2
  exit 1
fi

# Download and install
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
curl -sSfL -o "$TMP/openenvx" "$ASSET_URL"
chmod +x "$TMP/openenvx"
mv "$TMP/openenvx" "$BINDIR/openenvx"

# Verify
if "$BINDIR/openenvx" --version &>/dev/null; then
  "$BINDIR/openenvx" --version
else
  echo "Installed openenvx to $BINDIR/openenvx"
fi

# Auto-configure PATH if needed
if ! echo "$PATH" | grep -q "$BINDIR"; then
  # Detect shell config file
  case "${SHELL##*/}" in
    zsh) SHELL_CONFIG="$HOME/.zshrc" ;;
    bash) SHELL_CONFIG="$HOME/.bashrc" ;;
    *) SHELL_CONFIG="$HOME/.profile" ;;
  esac

  # Add to PATH in shell config if not already present
  if [[ -f "$SHELL_CONFIG" ]] && ! grep -q "\.local/bin" "$SHELL_CONFIG" 2>/dev/null; then
    echo "" >> "$SHELL_CONFIG"
    echo "# Added by OpenEnvX installer" >> "$SHELL_CONFIG"
    echo 'export PATH="$PATH:$HOME/.local/bin"' >> "$SHELL_CONFIG"
    echo "Added $BINDIR to PATH in $SHELL_CONFIG"
    echo "Run 'source $SHELL_CONFIG' or restart your terminal to use openenvx"
  else
    echo "Add $BINDIR to your PATH by adding this to your shell config:"
    echo '  export PATH="$PATH:$HOME/.local/bin"'
  fi
fi
