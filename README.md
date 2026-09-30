# home-exit

Turn a **real home (or mobile) internet connection** into a private SOCKS5 / HTTP CONNECT egress for bots, browsers, and apps.

This is **not** a commercial “residential proxy network” and **not** a Cloudflare/Hetzner VPS VPN. Those look like datacenters. This project makes *your* ISP IP available securely to machines you trust.

## Status

**MVP (milestones 1–2):** local Go CLI with authenticated SOCKS5 + HTTP CONNECT, user management with bcrypt-hashed passwords. Tunnel / Tailscale profiles come next — see [`docs/plan.md`](docs/plan.md) and [`docs/roadmap.md`](docs/roadmap.md).

## Requirements

- Go 1.22+

## Quick start

```bash
git clone https://github.com/lovrozagar/home-exit.git
cd home-exit
cp home-exit.example.yaml home-exit.yaml

# Create a client (password printed once; only bcrypt hash is stored)
go run ./cmd/home-exit users add alice

# Show config summary
go run ./cmd/home-exit status

# Start proxy (foreground)
go run ./cmd/home-exit up
```

Or build a binary:

```bash
go build -o bin/home-exit ./cmd/home-exit
./bin/home-exit users add alice
./bin/home-exit up
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

1. `-config` flag or `HOME_EXIT_CONFIG` env
2. `./home-exit.yaml`
3. `~/.config/home-exit/config.yaml`

Defaults (also in `home-exit.example.yaml`):

| Key | Default |
|---|---|
| `listen.socks` | `127.0.0.1:1080` |
| `listen.http` | `127.0.0.1:8080` |
| `users.file` | `users.json` (next to the config file) |

`home-exit.yaml` and `users.json` are gitignored — do not commit secrets. Hashes only land in `users.json`.

## CLI

| Command | Description |
|---|---|
| `home-exit up` | Start SOCKS5 + HTTP CONNECT (foreground) |
| `home-exit status` | Print listen addrs / users count (`running` is N/A in v1) |
| `home-exit users add <name>` | Create user; print random password once |
| `home-exit users list` | List usernames (no secrets) |

## Security

- Default bind is **localhost only** until you add a tunnel (Tailscale / Cloudflare Tunnel — planned).
- No anonymous access.
- Sharing credentials shares your home IP and legal exposure — treat them like SSH keys.
- See [`docs/why-not-cloud.md`](docs/why-not-cloud.md).

## Architecture

```
[ Bot / browser ] --SOCKS5/HTTP + auth--> [ home-exit agent ]
                                              |
                                              v
                                      home ISP / mobile IP
```

## License

MIT
