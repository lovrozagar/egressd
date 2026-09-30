# Why not Cloudflare / Hetzner / “VPN with region pin”?

Sites like X fingerprint **ASN + IP reputation**.

- **Hetzner / AWS / GCP / DO** → datacenter ASN → often blocked for login
- **Cloudflare WARP / Workers / colo pin** → Cloudflare ASN → still not your home ISP
- **Consumer VPN Croatia servers** → known VPN ranges → often blocked the same way

A **residential** exit means traffic leaves through an ISP customer connection (home fiber/cable or mobile). That is what this repo targets.
