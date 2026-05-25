#!/usr/bin/env bash
# Capture screenshots of the running PIdisk app for the README.
#
# Usage:
#   1. Start the app (wails dev) and open it.
#   2. Run this script. It will prompt you to bring each screen into view,
#      then capture the foreground window.
#   3. PNGs land in docs/screenshots/.
#
# Requires:
#   - macOS (uses screencapture)
#   - Screen Recording permission for your terminal:
#     System Settings -> Privacy & Security -> Screen Recording -> add your terminal app.

set -euo pipefail
cd "$(dirname "$0")/.."

OUT=docs/screenshots
mkdir -p "$OUT"

if ! command -v screencapture >/dev/null 2>&1; then
  echo "screencapture not found. macOS only." >&2
  exit 1
fi

capture() {
  local file=$1
  local description=$2
  echo
  echo "==> $description"
  echo "    Bring the relevant PIdisk view to the foreground, then press Enter."
  read -r
  # -o suppress shadow, -w wait for window click. Use -W to capture frontmost window without click.
  screencapture -o -W "$OUT/$file"
  echo "    Saved $OUT/$file"
}

capture "login.png"          "Login screen with at least one profile listed"
capture "create-profile.png" "Create-profile form (Advanced collapsed)"
capture "files-dark.png"     "Files page in dark theme, folder tree on the left"
capture "transfers.png"      "Transfer drawer with at least one active or finished transfer"
capture "sync.png"           "Sync panel drawer (open from the toolbar)"
capture "hostkey.png"        "Host-key confirmation dialog (open by connecting to a new host)"

echo
echo "All screenshots captured."
echo "Optimise size with: pngcrush -reduce -brute $OUT/*.png"
