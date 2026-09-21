#!/bin/bash
# *****************************************************************************
# Nehonix FileOnix Build Script
# 
# ACCESS RESTRICTIONS:
# - This software is exclusively for use by Authorized Personnel of NEHONIX
# - Intended for Internal Use only within NEHONIX operations
# - No rights granted to unauthorized individuals or entities
# - All modifications are works made for hire assigned to NEHONIX
# *****************************************************************************

set -e

APP_NAME="fileonix"
BUILD_DIR="bin"

# ANSI Colors
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
MAGENTA='\033[0;35m'
RED='\033[0;31m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m'

# Ensure we operate from FileOnix root directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Extract version from package.json if available
if [ -f "package.json" ]; then
    VERSION=$(grep '"version":' package.json | head -n1 | sed -E 's/.*"version": *"([^"]+)".*/\1/')
fi
if [ -z "$VERSION" ]; then
    VERSION="2.0.7"
fi

BUILD_TIME=$(date -u +'%Y-%m-%dT%H:%M:%SZ')

# Parse arguments
COMPRESS=false
PARALLEL=false
CURRENT_ONLY=false
USE_GARBLE=true

for arg in "$@"; do
    case "$arg" in
        --compress) COMPRESS=true ;;
        --parallel) PARALLEL=true ;;
        --current|-c) CURRENT_ONLY=true ;;
        --no-garble) USE_GARBLE=false ;;
        --help|-h)
            echo "Usage: ./build.sh [OPTIONS]"
            echo "Options:"
            echo "  --current, -c   Build only for current platform"
            echo "  --compress      Compress binaries with UPX"
            echo "  --parallel      Build platforms in parallel"
            echo "  --no-garble     Disable garble obfuscator (standard go build)"
            exit 0
            ;;
    esac
done

# Garble tool detection
GARBLE_BIN=""
if [ "$USE_GARBLE" = true ]; then
    export GOTOOLCHAIN="go1.27.1"
    if command -v garble &> /dev/null; then
        GARBLE_BIN="garble"
    elif [ -x "$HOME/go/bin/garble" ]; then
        GARBLE_BIN="$HOME/go/bin/garble"
    elif [ -x "$(go env GOPATH 2>/dev/null)/bin/garble" ]; then
        GARBLE_BIN="$(go env GOPATH)/bin/garble"
    fi

    if [ -n "$GARBLE_BIN" ]; then
        # Check if garble requires go1.27 toolchain
        if ! "$GARBLE_BIN" version >/dev/null 2>&1; then
            
            if "$GARBLE_BIN" version >/dev/null 2>&1; then
                echo -e " ${GREEN}🛡️  Garble AST obfuscator detected (GOTOOLCHAIN=go1.27.1)${NC}"
            else
                echo -e " ${YELLOW}⚠️  Garble incompatible with current go toolchain. Falling back to hardened go build.${NC}"
                USE_GARBLE=false
            fi
        else
            echo -e " ${GREEN}🛡️  Garble AST obfuscator detected${NC}"
        fi
    else
        echo -e " ${YELLOW}⚠️  Garble not found. Falling back to hardened standard build.${NC}"
        USE_GARBLE=false
    fi
fi

if [ "$COMPRESS" = true ]; then
    if ! command -v upx &> /dev/null; then
        echo "❌ upx command not found. Please install UPX to use the --compress flag."
        exit 1
    fi
fi

mkdir -p "$BUILD_DIR"

# Function to scrub embedded Go buildinfo / module dependency tables (path\t, mod\t, dep\t)
strip_go_buildinfo() {
    local target_file="$1"
    python3 -c "
import sys, re
target = sys.argv[1]
try:
    with open(target, 'r+b') as f:
        data = bytearray(f.read())
        changed = False
        # 1. Neutralize magic header '\xff Go buildinf:'
        for m in re.finditer(b'\\xff Go buildinf:', data):
            data[m.start():m.start()+32] = b'\x00' * 32
            changed = True
        # 2. Neutralize embedded module metadata tables (path\t, mod\t, dep\t)
        for m in re.finditer(b'path\\t', data):
            end = data.find(b'\x00', m.start())
            if end != -1:
                data[m.start():end] = b'\x00' * (end - m.start())
                changed = True
        if changed:
            f.seek(0)
            f.write(data)
            f.truncate()
except Exception:
    pass
" "$target_file" 2>/dev/null || true
}

# Platforms to build for
if [ "$CURRENT_ONLY" = true ]; then
    HOST_OS=$(go env GOOS)
    HOST_ARCH=$(go env GOARCH)
    PLATFORMS=("${HOST_OS}/${HOST_ARCH}")
    echo -e " ${CYAN}🎯 Building target:${NC} ${BOLD}${HOST_OS}/${HOST_ARCH}${NC} ${DIM}(current platform)${NC}\n"
else
    PLATFORMS=(
        "linux/amd64"
        "linux/arm64"
        "darwin/amd64"
        "darwin/arm64"
        "windows/amd64"
        "windows/arm64"
    )
    echo -e " ${CYAN}🎯 Building all 6 multi-platform targets...${NC}\n"
fi

echo -e " ${BOLD}🚀 Starting Nehonix FileOnix hardened build (v${VERSION})...${NC}"
[ "$PARALLEL" = true ] && echo -e " ${YELLOW}⚡ Parallel mode enabled — builds will run concurrently.${NC}"

LDFLAGS="-X main.VERSION=$VERSION -s -w -buildid="
GO_FLAGS=(-trimpath -buildvcs=false -ldflags="$LDFLAGS")

build_platform() {
    local PLATFORM="$1"
    local OS="${PLATFORM%/*}"
    local ARCH="${PLATFORM#*/}"
    local SUFFIX=""
    [ "$OS" == "windows" ] && SUFFIX=".exe"

    local OUTPUT_NAME="${APP_NAME}-${OS}-${ARCH}${SUFFIX}"
    local TARGET_PATH="${BUILD_DIR}/${OUTPUT_NAME}"
    local TAR_NAME="${APP_NAME}-${OS}-${ARCH}.tar.gz"
    local TAR_PATH="${BUILD_DIR}/${TAR_NAME}"

    printf "  ${CYAN}%-22s${NC} ${DIM}»${NC} " "Building ${OS}/${ARCH}"

    rm -f "$TARGET_PATH"

    if [ "$USE_GARBLE" = true ]; then
    export GOTOOLCHAIN="go1.27.1"
        if CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH "$GARBLE_BIN" -literals -tiny build "${GO_FLAGS[@]}" -o "$TARGET_PATH" ./internal; then
            strip_go_buildinfo "$TARGET_PATH"
            printf "${GREEN}✓ OBFUSCATED${NC} "
        else
            printf "${RED}❌ FAILED (Garble)${NC}\n"
            return 1
        fi
    else
        if CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH go build "${GO_FLAGS[@]}" -o "$TARGET_PATH" ./internal; then
            strip_go_buildinfo "$TARGET_PATH"
            printf "${GREEN}✓ STRIPPED${NC} "
        else
            printf "${RED}❌ FAILED (Go)${NC}\n"
            return 1
        fi
    fi

    # Native ELF strip on Linux
    if [ "$OS" == "linux" ] && command -v strip &> /dev/null; then
        strip --strip-all "$TARGET_PATH" 2>/dev/null || true
    fi

    # UPX compression
    if [ "$COMPRESS" = true ]; then
        if [ "$OS" == "darwin" ]; then
            printf "${DIM}[UPX skip macOS]${NC} "
        else
            if upx --best "$TARGET_PATH" > /dev/null 2>&1; then
                printf "${MAGENTA}✨ UPX${NC} "
            fi
        fi
    fi

    # Create tar.gz archive for release compatibility
    if [ "$OS" != "windows" ]; then
        chmod +x "$TARGET_PATH"
    fi
    tar -czf "$TAR_PATH" -C "$BUILD_DIR" "$OUTPUT_NAME" 2>/dev/null || true

    echo -e "${DIM}→ ${OUTPUT_NAME}${NC}"

    # If host platform, update ./fileonix
    local HOST_OS=$(go env GOOS)
    local HOST_ARCH=$(go env GOARCH)
    if [ "$OS" = "$HOST_OS" ] && [ "$ARCH" = "$HOST_ARCH" ]; then
        cp -f "$TARGET_PATH" "./${APP_NAME}"
        chmod +x "./${APP_NAME}"
    fi
}

if [ "$PARALLEL" = true ]; then
    PIDS=()
    FAILED=()
    for PLATFORM in "${PLATFORMS[@]}"; do
        build_platform "$PLATFORM" &
        PIDS+=($!)
    done

    FAIL=0
    for i in "${!PIDS[@]}"; do
        if ! wait "${PIDS[$i]}"; then
            FAILED+=("${PLATFORMS[$i]}")
            FAIL=1
        fi
    done

    if [ $FAIL -ne 0 ]; then
        echo -e "\n${RED}❌ Builds failed for: ${FAILED[*]}${NC}"
        exit 1
    fi
else
    for PLATFORM in "${PLATFORMS[@]}"; do
        build_platform "$PLATFORM"
    done
fi

echo -e "\n${GREEN}${BOLD}✨ All builds completed successfully!${NC}"
echo -e "${DIM}Artifacts saved in: ./${BUILD_DIR}${NC}\n"
