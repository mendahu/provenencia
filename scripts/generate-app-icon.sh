#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if ! command -v rsvg-convert >/dev/null; then
  echo "rsvg-convert is required (e.g. brew install librsvg)" >&2
  exit 1
fi

# Regenerates macos/App/Resources/Assets.xcassets/AppIcon.appiconset from the
# same mark used by LogoMark.imageset, composed onto a warm-parchment
# background (PVPalette.paper50 / PVColor.surfacePage's light value) at ~80%
# fill, matching Apple's icon-grid margin convention. macOS applies its own
# corner-rounding and drop shadow to a full-bleed square icon automatically,
# so no masking is done here.
#
# Rerun this whenever the mark or brand color changes; it fully regenerates
# every PNG in AppIcon.appiconset.

SRC="macos/App/Resources/Assets.xcassets/LogoMark.imageset/logo-mark.svg"
DEST="macos/App/Resources/Assets.xcassets/AppIcon.appiconset"
BG="#F8F4ED"
MARK_FRACTION="0.8"

# The mark's own markup (everything between its <svg ...> and </svg> tags),
# so it can be re-embedded inside a per-size canvas rather than hand-copied.
inner="$(sed -n '2,/<\/svg>/p' "$SRC" | sed '$d')"

render() {
  local size="$1" name="$2"
  local mark scale margin tmp
  mark=$(awk -v s="$size" -v f="$MARK_FRACTION" 'BEGIN { printf "%.4f", s * f }')
  scale=$(awk -v m="$mark" 'BEGIN { printf "%.6f", m / 48 }')
  margin=$(awk -v s="$size" -v m="$mark" 'BEGIN { printf "%.4f", (s - m) / 2 }')
  tmp="$(mktemp /tmp/pv-app-icon-XXXX.svg)"
  cat > "$tmp" <<EOF
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 $size $size">
  <rect width="$size" height="$size" fill="$BG"/>
  <g transform="translate($margin,$margin) scale($scale)">
    $inner
  </g>
</svg>
EOF
  rsvg-convert -w "$size" -h "$size" "$tmp" -o "$DEST/$name"
  rm -f "$tmp"
}

render 16   icon_16x16.png
render 32   icon_16x16@2x.png
render 32   icon_32x32.png
render 64   icon_32x32@2x.png
render 128  icon_128x128.png
render 256  icon_128x128@2x.png
render 256  icon_256x256.png
render 512  icon_256x256@2x.png
render 512  icon_512x512.png
render 1024 icon_512x512@2x.png

echo "Generated 10 AppIcon PNGs into $DEST"
