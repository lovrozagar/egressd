# Why not Cloudflare / Hetzner / “VPN with region pin”?

Sites like X fingerprint **ASN + IP reputation**.

- **Hetzner / AWS / GCP / DO** → datacenter ASN → often blocked for login
- **Cloudflare WARP / Workers / colo pin** → Cloudflare ASN → still not your home ISP
- **Consumer VPN Croatia servers** → known VPN ranges → often blocked the same way

A **residential / ISP** exit means traffic leaves through a consumer-grade connection (home fiber/cable, office ISP, or mobile). That is what egressd targets — self-hosted on hardware you control, not a rented datacenter ASN.
