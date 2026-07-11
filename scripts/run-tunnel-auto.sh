#!/usr/bin/env bash
# Run on the LAPTOP, not the phone - cloudflared's binary doesn't run on
# Android 5.1. The phone still runs the actual blog server on :8080; this
# just proxies it to the internet over LAN.
#
# Unlike run-tunnel.sh, this script closes the loop: it starts the Quick
# Tunnel, scrapes the random https://*.trycloudflare.com URL from its logs,
# writes it into cloudflare/worker/wrangler.toml as PHONE_URL, and runs
# `wrangler deploy` so the Worker picks it up - no manual dashboard edit.
#
# Requires: cloudflare/worker/wrangler.toml already exists (copy from
# wrangler.toml.example) and `wrangler` is already authenticated
# (wrangler login).
#
# Usage: run-tunnel-auto.sh <phone-lan-ip>
# Find the phone's LAN IP in Termux with: ifconfig | grep wlan0 -A1

set -euo pipefail

PHONE_IP="${1:?Usage: run-tunnel-auto.sh <phone-lan-ip>}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
WORKER_DIR="$REPO_ROOT/cloudflare/worker"
WRANGLER_TOML="$WORKER_DIR/wrangler.toml"
LOG_FILE="$(mktemp -t cloudflared-tunnel.XXXXXX.log)"

if [ ! -f "$WRANGLER_TOML" ]; then
    echo "wrangler.toml not found at $WRANGLER_TOML. Copy wrangler.toml.example to wrangler.toml and fill in RENDER_URL first." >&2
    exit 1
fi

cleanup() {
    if [ -n "${TUNNEL_PID:-}" ] && kill -0 "$TUNNEL_PID" 2>/dev/null; then
        kill "$TUNNEL_PID" 2>/dev/null || true
    fi
}
trap cleanup EXIT

echo "Starting Cloudflare Quick Tunnel -> http://${PHONE_IP}:8080 ..."
cloudflared tunnel --url "http://${PHONE_IP}:8080" >"$LOG_FILE" 2>&1 &
TUNNEL_PID=$!

TUNNEL_URL=""
for _ in $(seq 1 60); do
    if ! kill -0 "$TUNNEL_PID" 2>/dev/null; then
        echo "cloudflared exited early. Check $LOG_FILE" >&2
        exit 1
    fi
    TUNNEL_URL="$(grep -oE 'https://[a-zA-Z0-9-]+\.trycloudflare\.com' "$LOG_FILE" | head -n1 || true)"
    if [ -n "$TUNNEL_URL" ]; then
        break
    fi
    sleep 0.5
done

if [ -z "$TUNNEL_URL" ]; then
    echo "Timed out waiting for tunnel URL after 30s. Check $LOG_FILE" >&2
    exit 1
fi

echo "Tunnel URL: $TUNNEL_URL"

if grep -qE '^PHONE_URL\s*=' "$WRANGLER_TOML"; then
    sed -i.bak -E "s|^PHONE_URL\s*=.*|PHONE_URL = \"$TUNNEL_URL\"|" "$WRANGLER_TOML"
else
    sed -i.bak -E "s|^\[vars\]$|[vars]\nPHONE_URL = \"$TUNNEL_URL\"|" "$WRANGLER_TOML"
fi
rm -f "$WRANGLER_TOML.bak"

echo "Deploying Worker with updated PHONE_URL..."
(cd "$WORKER_DIR" && wrangler deploy)

echo ""
echo "Worker updated. Tunnel running (PID $TUNNEL_PID)."
echo "Press Ctrl+C to stop the tunnel."
trap - EXIT
wait "$TUNNEL_PID"
