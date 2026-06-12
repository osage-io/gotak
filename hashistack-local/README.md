# hashistack

Single-node local development stack for HashiCorp **Consul**, **Vault**, and
**Nomad**. Each service runs in `-dev` mode under your user account, with
PIDs and logs tracked in this directory so they can be cleanly stopped.

## Quick start

```bash
make hashi-up        # install (via brew if needed) + start all three
make hashi-status    # show health
make nomad-deploy    # deploy the GoTAK standalone stack to the local Nomad
make hashi-down      # stop everything and clean ephemeral data
```

## Layout

```
hashistack/
├── up.sh             # starts consul, vault, nomad (dev mode, single node)
├── down.sh           # stops all three and removes ./data
├── status.sh         # quick health check
├── nomad-deploy.sh   # renders + submits gotak-complete.nomad.hcl locally
├── data/             # ephemeral consul/nomad data (gitignored)
├── logs/             # consul.log, vault.log, nomad.log
├── run/              # *.pid files
└── .rendered/        # temp Nomad job specs with local node substitutions
```

## Endpoints

| Service | URL                           | Token |
|---------|-------------------------------|-------|
| Consul  | http://127.0.0.1:8500/ui      | n/a   |
| Vault   | http://127.0.0.1:8200/ui      | `root`|
| Nomad   | http://127.0.0.1:4646/ui      | n/a   |

Export these to use the CLIs against this stack:

```bash
export CONSUL_HTTP_ADDR=http://127.0.0.1:8500
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=root
export NOMAD_ADDR=http://127.0.0.1:4646
```

## Notes

- **Auto-install**: `up.sh` installs missing binaries via
  `brew install hashicorp/tap/{consul,vault,nomad}`.
- **Docker driver**: Nomad's docker driver needs Docker Desktop running.
- **Data**: `down.sh` deletes `./data` by default. Set `HASHI_KEEP_DATA=1`
  to keep Consul state across restarts.
- **Node name**: Nomad starts with `-node gotak-dev`. `nomad-deploy.sh`
  rewrites the canonical `hashinuc01` constraints in
  `nomad/deployments/standalone/gotak-complete.nomad.hcl` to that local
  node and points DB hostnames at `127.0.0.1`.
