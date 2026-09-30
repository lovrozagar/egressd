# Existing tools we lean on

- [Tailscale exit nodes](https://tailscale.com/kb/1103/exit-nodes)
- [ItalyPaleAle/tailsocks](https://github.com/ItalyPaleAle/tailsocks) — local SOCKS via a Tailscale exit
- [pkarpovich/vpn-exit-node](https://github.com/pkarpovich/vpn-exit-node) — compose appliance for Tailscale exit (+ optional SOCKS)
- OpenSSH `ssh -D` — classic dynamic SOCKS
- [wireproxy](https://github.com/pufferffish/wireproxy) — userspace WireGuard → SOCKS (useful *after* you have a home WireGuard peer)
