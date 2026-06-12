# demoland.io DNS — Namecheap records for the gotak demo

The OpenShift cluster stays on its install-time base domain
(`gotak.daniel-fedick.aws.sbx.hashicorpdemo.com`) — that can't move without a
reinstall. These `demoland.io` subdomains sit *in front* of it.

Add under **Namecheap → Domain List → demoland.io → Manage → Advanced DNS**:

| Type  | Host       | Value                                            | For |
|-------|------------|--------------------------------------------------|-----|
| A     | `boundary` | `<boundary-node public IP>`                      | Boundary bastion (no cert needed) |
| CNAME | `gotak`    | `<router canonical host>`                        | gotak app (API Gateway) |
| CNAME | `vault`    | `<router canonical host>`                        | Vault UI/API |
| CNAME | `consul`   | `<router canonical host>`                        | Consul UI |

Get the **router canonical host** (same for all three CNAMEs) after the custom
Routes exist:

```bash
oc -n gotak get route gotak-demoland \
  -o jsonpath='{.status.ingress[0].routerCanonicalHostname}{"\n"}'
```

Notes:
- Namecheap can't CNAME the apex (`demoland.io`) — that's why these are all
  subdomains. If a CNAME to the router host misbehaves, use **A** records to the
  router's public IPs instead (resolve them with
  `host console-openshift-console.apps.gotak.daniel-fedick.aws.sbx.hashicorpdemo.com`).
- TLS is **bring-your-own**: a `*.demoland.io` wildcard cert covers `gotak`,
  `vault`, and `consul`. `custom-domain-routes.sh` attaches it to the Routes.
- `boundary` needs no cert — the demo Boundary config uses `tls_disable`. Set
  `BOUNDARY_PUBLIC_ADDR=boundary.demoland.io` when running `install-boundary.sh`.
