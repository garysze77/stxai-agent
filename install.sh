#!/bin/sh
set -e

REPO="garysze77/stxai-agent"
BINARY="stxai"
VERSION="${VERSION:-latest}"

echo "╔══════════════════════════════════╗"
echo "║   STX AI Agent — Installer      ║"
echo "╚══════════════════════════════════╝"
echo ""

# Detect OS and arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unsupported arch: $ARCH"; exit 1 ;;
esac

case "$OS" in
    darwin|linux) ;;
    *) echo "Unsupported OS: $OS. Try Windows binary from GitHub Releases."; exit 1 ;;
esac

TARBALL="${BINARY}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/${VERSION}/download/${TARBALL}"

if [ "$VERSION" = "latest" ]; then
    URL="https://github.com/${REPO}/releases/latest/download/${TARBALL}"
fi

echo "→ Downloading ${TARBALL}..."
if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$URL" -o "/tmp/${TARBALL}"
elif command -v wget >/dev/null 2>&1; then
    wget -q "$URL" -O "/tmp/${TARBALL}"
else
    echo "Please install curl or wget."
    exit 1
fi

echo "→ Extracting ${TARBALL}..."
TMPDIR="$(mktemp -d)"
tar -xzf "/tmp/${TARBALL}" -C "$TMPDIR"
rm "/tmp/${TARBALL}"

echo "→ Installing to /usr/local/bin/${BINARY}..."
chmod +x "${TMPDIR}/${BINARY}"
sudo mv "${TMPDIR}/${BINARY}" /usr/local/bin/${BINARY}
rm -rf "$TMPDIR"

echo ""
echo "✅ STX AI Agent installed!"
echo "   Run 'stxai setup' to configure your API key."
echo "   Run 'stxai chat' to start chatting."
echo "   Run 'stxai start' to launch the Telegram bot."
