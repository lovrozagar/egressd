#!/usr/bin/env bash
set -euo pipefail
echo "home-exit (Linux) — install Tailscale, then:"
echo "  curl -fsSL https://tailscale.com/install.sh | sh"
echo "  sudo tailscale up --advertise-exit-node"
echo "Approve the exit node in the admin console."
