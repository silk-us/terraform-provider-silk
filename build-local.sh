#!/usr/bin/env bash
# build-local.sh — Build terraform-provider-silk against the LOCAL silk-sdp-go-sdk
# checkout instead of the version pinned in go.mod. Restores go.mod on exit so
# nothing accidentally gets committed.
#
# Usage:
#   ./build-local.sh                      # build for current host (darwin_arm64 etc.)
#   ./build-local.sh --all                # build all release targets
#   ./build-local.sh --install            # build for host, install to ~/.terraform.d
#   ./build-local.sh --sdk ../some/path   # override SDK path (default: ../silk-sdp-go-sdk)

set -euo pipefail

PROVIDER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SDK_DIR="${PROVIDER_DIR}/../silk-sdp-go-sdk"
MODE="host"

while [[ $# -gt 0 ]]; do
    case "$1" in
        --all)     MODE="all"; shift ;;
        --install) MODE="install"; shift ;;
        --sdk)     SDK_DIR="$2"; shift 2 ;;
        -h|--help)
            sed -n '2,11p' "$0"; exit 0 ;;
        *) echo "Unknown arg: $1" >&2; exit 2 ;;
    esac
done

if [[ ! -d "$SDK_DIR/silksdp" ]]; then
    echo "ERROR: SDK not found at $SDK_DIR (expected silksdp/ inside)" >&2
    exit 1
fi
SDK_DIR="$(cd "$SDK_DIR" && pwd)"

# Pull VERSION/BINARY from the Makefile so we don't drift.
VERSION="$(awk -F= '/^VERSION/ {gsub(/ /,"",$2); print $2; exit}' "$PROVIDER_DIR/Makefile")"
BINARY="$(awk -F= '/^BINARY/ {gsub(/ /,"",$2); print $2; exit}' "$PROVIDER_DIR/Makefile" | sed 's/\${NAME}/silk/')"

cd "$PROVIDER_DIR"
mkdir -p ./bin

# Snapshot go.mod / go.sum so we can restore them no matter what.
cp go.mod go.mod.bak
cp go.sum go.sum.bak
restore() {
    mv go.mod.bak go.mod
    mv go.sum.bak go.sum
    echo "Restored go.mod / go.sum"
}
trap restore EXIT

# Add (or rewrite) the replace directive to point at the local SDK.
go mod edit -replace "github.com/silk-us/silk-sdp-go-sdk=${SDK_DIR}"
go mod tidy

echo "Building against local SDK at: $SDK_DIR"
echo "Provider version: $VERSION"

case "$MODE" in
    host)
        OUT="./bin/${BINARY}_${VERSION}_$(go env GOOS)_$(go env GOARCH)"
        go build -o "$OUT"
        echo "Built: $OUT"
        ;;
    install)
        # Mimic `make install` but using the local SDK.
        HOST_OS_ARCH="$(go env GOOS)_$(go env GOARCH)"
        INSTALL_DIR="$HOME/.terraform.d/plugins/localdomain/provider/silk/${VERSION}/${HOST_OS_ARCH}"
        mkdir -p "$INSTALL_DIR"
        go build -o "$INSTALL_DIR/${BINARY}"
        OUT="./bin/${BINARY}_${VERSION}_${HOST_OS_ARCH}"
        cp "$INSTALL_DIR/${BINARY}" "$OUT"
        echo "Built: $OUT"
        echo "Installed: $INSTALL_DIR/${BINARY}"
        ;;
    all)
        targets=(
            "darwin amd64"
            "darwin arm64"
            "freebsd 386"
            "freebsd amd64"
            "freebsd arm"
            "linux 386"
            "linux amd64"
            "linux arm"
            "openbsd 386"
            "openbsd amd64"
            "solaris amd64"
            "windows 386"
            "windows amd64"
        )
        for t in "${targets[@]}"; do
            read -r os arch <<<"$t"
            ext=""
            [[ "$os" == "windows" ]] && ext=".exe"
            OUT="./bin/${BINARY}_${VERSION}_${os}_${arch}${ext}"
            GOOS="$os" GOARCH="$arch" go build -o "$OUT"
            echo "Built: $OUT"
        done
        chmod +x ./bin/* || true
        ;;
esac

echo "Done."
