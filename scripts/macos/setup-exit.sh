#!/usr/bin/env bash
# v0: guide Tailscale exit-node enablement on macOS.
set -euo pipefail
echo "egressd (macOS) — v0 uses Tailscale as the transport."
echo
echo "1) Install Tailscale from https://tailscale.com/download/mac"
echo "2) Log in, then enable this Mac as an Exit Node:"
echo "   - Admin console → Machine → Edit route settings → Exit node"
echo "   - Or: sudo tailscale set --advertise-exit-node"
echo "3) Approve the exit node in the Tailscale admin console."
echo "4) On the client (bot), use tailsocks or Tailscale SOCKS against this exit."
echo
echo "Do NOT expose an unauthenticated SOCKS port on your LAN."
