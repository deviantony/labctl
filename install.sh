#!/bin/sh
set -eu

REPO="deviantony/labctl"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS
OS="$(uname -s)"
case "$OS" in
  Linux)  OS="Linux" ;;
  Darwin) OS="Darwin" ;;
  *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64)   ARCH="x86_64" ;;
  aarch64|arm64)   ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

# Resolve version
if [ -z "${VERSION:-}" ]; then
  VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"v([^"]+)".*/\1/')"
  if [ -z "$VERSION" ]; then
    echo "Failed to fetch latest version" >&2
    exit 1
  fi
fi

TARBALL="labctl_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/v${VERSION}/${TARBALL}"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

echo "Downloading labctl v${VERSION} for ${OS}/${ARCH}..."
curl -fsSL "$URL" -o "${TMPDIR}/${TARBALL}"
tar -xzf "${TMPDIR}/${TARBALL}" -C "$TMPDIR"

echo "Installing to ${INSTALL_DIR}/labctl..."
install -m 755 "${TMPDIR}/labctl" "${INSTALL_DIR}/labctl"

echo "labctl v${VERSION} installed successfully"
labctl --version
