# home-exit

Turn a **real home (or mobile) internet connection** into a private SOCKS5 / HTTP egress for bots, browsers, and apps.

This is **not** a commercial “residential proxy network” and **not** a Cloudflare/Hetzner VPS VPN. Those look like datacenters. This project makes *your* ISP IP available securely to machines you trust (e.g. a Grok Bot / CI box that gets blocked by X/Twitter).

## Does this already exist?

Mostly yes — as building blocks. We wrap and document the boring reliable stack instead of inventing new crypto.

| Approach | Pros | Cons |
|---|---|---|
| **Tailscale exit node** + SOCKS (`tailsocks`, Tailscale `--socks5-server`) | Cross‑platform, no open ports, battle‑tested | Needs Tailscale account |
| **SSH dynamic port forward** (`ssh -D`) | Zero install beyond OpenSSH | Needs reachability (port forward / tunnel) |
| **Outline / Shadowsocks / SoftEther** | Mature VPN products | Heavier; still need a *home* host |
| Paid residential proxies | Easy | Costs money; third party sees traffic |

Closest ready-made repos: [ItalyPaleAle/tailsocks](https://github.com/ItalyPaleAle/tailsocks), [pkarpovich/vpn-exit-node](https://github.com/pkarpovich/vpn-exit-node), Tailscale’s own exit nodes.

**home-exit** = opinionated glue: one home agent, auth’d proxy, Mac/Linux/Windows install scripts, and a client recipe for bots.

## Platforms

| Role | Linux | macOS | Windows |
|---|---|---|---|
| **Home exit host** (where the residential IP lives) | ✅ target | ✅ target | ✅ target |
| **Client** (bot / laptop using the exit) | ✅ | ✅ | ✅ |

Priority order for v0: **macOS + Linux** (your MacBook case), then Windows.

## Architecture (v0)

```
[ Bot / browser ] --SOCKS5/HTTP--> [ tunnel ] --> [ home-exit agent on Mac/PC ]
                                                      |
                                                      v
                                              home ISP / mobile IP
```

Recommended tunnel: **Tailscale** (no inbound ports). Fallback: **Cloudflare Tunnel** or SSH.

Cloudflare *Workers/WARP regions* do **not** give you a residential Croatia IP. Cloudflare Tunnel is only the *pipe* to your home machine; egress still exits your house.

## Status

Scaffolding. Roadmap in [`docs/roadmap.md`](docs/roadmap.md).

## Quick start (planned)

1. Install Tailscale on the home machine → enable **Exit node**.
2. On the client: run tailsocks / Tailscale SOCKS against that exit.
3. Point the browser or app at `socks5://127.0.0.1:…`.

See [`docs/why-not-cloud.md`](docs/why-not-cloud.md).

## License

MIT
