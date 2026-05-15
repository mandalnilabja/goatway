#!/bin/bash
# Goatway Installation Script
# Usage: curl -fsSL https://raw.githubusercontent.com/mandalnilabja/goatway/main/install.sh | bash
#
# Environment variables:
#   INSTALL_DIR - Installation directory (default: /usr/local/bin)
#   VERSION     - Specific version to install (default: latest)

set -e

REPO="mandalnilabja/goatway"
BINARY_NAME="goatway"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() { echo -e "${GREEN}[INFO]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1" >&2; exit 1; }

cleanup() { rm -rf "${TMP_DIR}"; }

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Darwin*) echo "darwin" ;;
        Linux*)  echo "linux" ;;
        MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
        *) error "Unsupported operating system: $(uname -s)" ;;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *) error "Unsupported architecture: $(uname -m)" ;;
    esac
}

# Get latest release version from GitHub
get_latest_version() {
    curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | \
        grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/' || echo ""
}

main() {
    info "Installing Goatway..."

    OS=$(detect_os)
    ARCH=$(detect_arch)

    info "Detected OS: ${OS}, Architecture: ${ARCH}"

    # Determine version
    if [ -z "${VERSION}" ]; then
        VERSION=$(get_latest_version)
        if [ -z "${VERSION}" ]; then
            error "Could not determine latest version. Set VERSION env var or check your internet connection."
        fi
    fi
    info "Installing version: ${VERSION}"

    # Build download URL
    if [ "${OS}" = "windows" ]; then
        FILENAME="${BINARY_NAME}-${OS}-${ARCH}.exe"
    else
        FILENAME="${BINARY_NAME}-${OS}-${ARCH}"
    fi

    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"

    info "Downloading from: ${DOWNLOAD_URL}"

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    trap cleanup EXIT

    # Download binary
    if ! curl -fsSL "${DOWNLOAD_URL}" -o "${TMP_DIR}/${BINARY_NAME}"; then
        error "Failed to download ${DOWNLOAD_URL}"
    fi

    # Make executable
    chmod +x "${TMP_DIR}/${BINARY_NAME}"

    # Install
    info "Installing to ${INSTALL_DIR}/${BINARY_NAME}"

    if [ -w "${INSTALL_DIR}" ]; then
        mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        warn "Elevated permissions required to install to ${INSTALL_DIR}"
        sudo mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    fi

    # Verify installation
    INSTALLED_BIN="${INSTALL_DIR}/${BINARY_NAME}"
    if "${INSTALLED_BIN}" -version &> /dev/null; then
        info "Installation successful!"
        echo ""
        "${INSTALLED_BIN}" -version
        echo ""
        info "Run 'goatway' to start the server"
        info "Web UI will be available at http://localhost:8080"
    else
        error "Installation verification failed. Binary at ${INSTALLED_BIN} is not executable."
    fi
}

main "$@"
