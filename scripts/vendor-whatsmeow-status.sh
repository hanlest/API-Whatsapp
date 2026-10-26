#!/bin/sh
set -e

ROOT="$(CDPATH= cd "$(dirname "$0")/.." && pwd)"
GOMOD="${GOMOD:-$ROOT/src/go.mod}"
if [ ! -f "$GOMOD" ]; then
  GOMOD="$ROOT/go.mod"
fi
if [ ! -f "$GOMOD" ]; then
  echo "go.mod not found under $ROOT" >&2
  exit 1
fi

MODROOT="$(CDPATH= cd "$(dirname "$GOMOD")" && pwd)"
DEST="$(CDPATH= cd "$MODROOT/.." && pwd)/third_party/whatsmeow"

VER="$(grep 'go.mau.fi/whatsmeow' "$GOMOD" | sed -n 's/.*\(v0\.0\.0-[0-9]*-[a-f0-9]*\).*/\1/p' | head -1)"
if [ -z "$VER" ]; then
  echo "Could not parse whatsmeow version from $GOMOD" >&2
  exit 1
fi

(cd "$MODROOT" && go mod download "go.mau.fi/whatsmeow@$VER")
SRC="$(go env GOMODCACHE)/go.mau.fi/whatsmeow@$VER"
if [ ! -d "$SRC" ]; then
  echo "Module cache missing: $SRC" >&2
  exit 1
fi

rm -rf "$DEST"
mkdir -p "$(dirname "$DEST")"
cp -a "$SRC" "$DEST"

PYTHON="${PYTHON:-python3}"
if ! command -v "$PYTHON" >/dev/null 2>&1; then
  echo "python3 required to patch send.go" >&2
  exit 1
fi
"$PYTHON" "$ROOT/scripts/patch_whatsmeow_status_send.py" "$DEST/send.go"

echo "whatsmeow vendored and patched at $DEST"
