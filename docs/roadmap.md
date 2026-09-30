# Roadmap

## v0 — MacBook as residential exit (solve X login blocks)

- [ ] Document Tailscale exit-node path (Mac + Linux)
- [ ] Client SOCKS recipe for Linux bots
- [ ] macOS install script (`scripts/macos/setup-exit.sh`)
- [ ] Security checklist (auth, no open LAN proxy, kill switch)

## v1 — thin `home-exit` agent

- [x] Small Go binary: authenticated SOCKS5 + HTTP CONNECT (`cmd/home-exit`)
- [x] Bind to localhost by default (`127.0.0.1:1080` / `:8080`)
- [x] User management with bcrypt hashes (`users add` / `users list`)
- [ ] Optional Cloudflare Tunnel sidecar compose file
- [ ] Windows PowerShell setup script

## v2 — niceties

- [ ] Sticky session helpers
- [ ] Health / “what’s my egress IP?” endpoint
- [ ] Multi-home / multi-region exits (pick which house)
