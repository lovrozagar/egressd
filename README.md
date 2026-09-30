# egressd

Self-hosted **authenticated residential / ISP egress** proxy (SOCKS5 + HTTP CONNECT).

Run it on any machine with a real ISP or mobile connection — home fiber, office, LTE hotspot — and expose that IP securely to bots, browsers, and apps you trust. This is **not** a commercial residential proxy marketplace and **not** a Cloudflare/Hetzner VPS VPN. Those look like datacenters. egressd makes *your* ISP IP available to clients you authorize.

## Status

**MVP (milestones 1–2):** local Go CLI with authenticated SOCKS5 + HTTP CONNECT, user management with bcrypt-hashed passwords. Tunnel / Tailscale profiles come next — see [`docs/plan.md`](docs/plan.md) and [`docs/roadmap.md`](docs/roadmap.md).

## Requirements

- Go 1.22+

## Quick start

```bash
git clone https://github.com/lovrozagar/egressd.git
cd egressd
cp egressd.example.yaml egressd.yaml

# Create a client (password printed once; only bcrypt hash is stored)
go run ./cmd/egressd users add alice

# Show config summary
go run ./cmd/egressd status

# Start proxy (foreground)
go run ./cmd/egressd up
```

Or build a binary:

```bash
go build -o bin/egressd ./cmd/egressd
./bin/egressd users add alice
./bin/egressd up
```

### Connect examples

With the password from `users add`:

```bash
curl -x socks5h://alice:PASS@127.0.0.1:1080 https://ifconfig.me
curl -x http://alice:PASS@127.0.0.1:8080 https://ifconfig.me
```

Anonymous / wrong credentials are rejected. Auth is always required.

## Config

Search order:

1. `-config` flag or `EGRESSD_CONFIG` env
2. `./egressd.yaml`
3. `~/.config/egressd/config.yaml`

Defaults (also in `egressd.example.yaml`):

| Key | Default |
|---|---|
| `listen.socks` | `127.0.0.1:1080` |
| `listen.http` | `127.0.0.1:8080` |
| `users.file` | `users.json` (next to the config file) |

`egressd.yaml` and `users.json` are gitignored — do not commit secrets. Hashes only land in `users.json`.

## CLI

| Command | Description |
|---|---|
| `egressd up` | Start SOCKS5 + HTTP CONNECT (foreground) |
| `egressd status` | Print listen addrs / users count (`running` is N/A in v1) |
| `egressd users add <name>` | Create user; print random password once |
| `egressd users list` | List usernames (no secrets) |

## Security

- Default bind is **localhost only** until you add a tunnel (Tailscale / Cloudflare Tunnel — planned).
- No anonymous access.
- Sharing credentials shares your egress IP and legal exposure — treat them like SSH keys.
- See [`docs/why-not-cloud.md`](docs/why-not-cloud.md).

## Architecture

```
[ Bot / browser ] --SOCKS5/HTTP + auth--> [ egressd agent ]
                                              |
                                              v
                                      residential / ISP / mobile IP
```

## License

MIT
