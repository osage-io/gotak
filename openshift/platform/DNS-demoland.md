# demoland.io DNS — Namecheap records for the gotak demo

The OpenShift cluster stays on its install-time base domain
(`gotak.daniel-fedick.aws.sbx.hashicorpdemo.com`) — that can't move without a
reinstall. These `demoland.io` subdomains sit *in front* of it.

Add under **Namecheap → Domain List → demoland.io → Manage → Advanced DNS**.
Either record type works — TLS matches on the hostname, not on how DNS resolved,
so the wildcard cert is happy with either.

### Option A — A records (simplest; pin the current ingress IPs)

The apps router resolves directly to two IPs (no LB hostname in between). List both
for round-robin, or just one.

| Type | Host       | Value                       | For |
|------|------------|-----------------------------|-----|
| A    | `boundary` | `<boundary-node public IP>` | Boundary bastion (no cert needed) |
| A    | `gotak`    | `18.217.224.91`             | gotak app (API Gateway) |
| A    | `gotak`    | `18.116.145.144`            | (2nd ingress IP, optional) |
| A    | `vault`    | `18.217.224.91`             | Vault UI/API |
| A    | `vault`    | `18.116.145.144`            | (optional) |
| A    | `consul`   | `18.217.224.91`             | Consul UI |
| A    | `consul`   | `18.116.145.144`            | (optional) |

Re-check the current IPs any time with:
```bash
dig +short console-openshift-console.apps.gotak.daniel-fedick.aws.sbx.hashicorpdemo.com
```
Trade-off: if the cluster's ingress IPs ever change (reinstall, ingress
recreation), you update these A records by hand.

### Option B — CNAME records (auto-follow IP changes)

| Type  | Host       | Value                       |
|-------|------------|-----------------------------|
| A     | `boundary` | `<boundary-node public IP>` |
| CNAME | `gotak`    | `<router canonical host>`   |
| CNAME | `vault`    | `<router canonical host>`   |
| CNAME | `consul`   | `<router canonical host>`   |

Get the **router canonical host** (same for all three CNAMEs) after the custom
Routes exist:

```bash
oc -n gotak get route gotak-demoland \
  -o jsonpath='{.status.ingress[0].routerCanonicalHostname}{"\n"}'
```

Notes:
- Namecheap can't CNAME the apex (`demoland.io`) — that's why these are all
  subdomains. `boundary` is an A record either way (single node, no router).
- TLS is **bring-your-own**: a `*.demoland.io` wildcard cert covers `gotak`,
  `vault`, and `consul`. `custom-domain-routes.sh` attaches it to the Routes.
- `boundary` needs no cert — the demo Boundary config uses `tls_disable`. Set
  `BOUNDARY_PUBLIC_ADDR=boundary.demoland.io` when running `install-boundary.sh`.
