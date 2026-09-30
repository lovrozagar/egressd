# egressd — build plan

## Goal

Clone → start on any machine with a real ISP or mobile connection → get a shareable authenticated proxy endpoint. Clients (browsers, apps, bots) connect with credentials. Egress uses the host’s real ISP IP (residential / consumer ASN — home, office, or mobile).

Not: commercial residential marketplace, Cloudflare/Hetzner “fake residential”, or bot-only coupling.

## Stack (v1)

| Layer | Choice | Why |
|---|---|---|
| Language | **Go** | One static binary for Linux / macOS / Windows; easy SOCKS5 + HTTP CONNECT |
| Local proxy | **SOCKS5 + HTTP CONNECT** on one port (or twin ports) | Universal client support |
| Auth | **Username + password** (per-client secrets) | Simple; works with every SOCKS/HTTP client |
| Public reachability | **Cloudflare Tunnel** (`cloudflared`) as default | No router port-forward; free; TCP via `cloudflared` private routing **or** published hostname patterns we document |
| Alt reachability | **Tailscale** (optional profile) | Zero open ports; great if both sides run Tailscale |
| Config | `egressd.yaml` + env | Clone-and-edit |
| UX | `egressd up` CLI | Prints connect URL + example curl |

**Explicit non-goals for v1:** rotating IP pools, bandwidth marketplace, GUI, mobile app store builds.

### Cloudflare note

Cloudflare Tunnel is the **pipe to the egress host**, not the exit ASN. Traffic still leaves via that host’s ISP. We will **not** claim CF region pinning = residential.

If Cloudflare TCP expose is awkward for raw SOCKS, fallback path: tunnel SSH or use Tailscale profile; document both.

## Architecture

```
Client --(auth)--> public endpoint (CF Tunnel or Tailscale IP)
                         |
                         v
              egressd agent (Go) on Mac/Linux/Windows
                         |
                         v
                   ISP / Wi‑Fi / LTE egress IP
```

## Repo layout (target)

```
cmd/egressd/          # CLI: up, users add/list, status
internal/proxy/         # SOCKS5 + HTTP CONNECT + auth
internal/config/
internal/users/         # bcrypt-hashed client credentials
egressd.example.yaml
deploy/cloudflared/     # example tunnel config (milestone 3)
scripts/{macos,linux,windows}/
docs/plan.md            # this file
README.md               # clone-and-start
```

## Milestones

1. [x] **Go agent MVP** — local SOCKS5+HTTP with user/pass; bind localhost; `egressd up`
2. [x] **User management** — `egressd users add` writes hashed secrets
3. [ ] **Tunnel profile** — cloudflared compose/docs so endpoint is reachable off-LAN
4. [ ] **Cross-compile** — release binaries for darwin/linux/windows amd64+arm64
5. [ ] **README polish** — true clone-and-start; security warnings (clone-and-start docs started in README)

## Security bar

- Default bind: localhost / Tailscale IP only until tunnel is configured
- No anonymous access
- Rate-limit failed auth
- Clear warning: sharing creds shares your egress IP & legal exposure

## Success check

From a second network: `curl -x socks5h://user:pass@ENDPOINT:PORT https://ifconfig.me` returns the host’s **real ISP** public IP.
