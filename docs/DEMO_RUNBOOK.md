# GoTAK Demo Runbook — Kafka Comms + Vault Transit Encryption

End-to-end demo: open GoTAK and Redpanda Console, show **plaintext** chat
messages landing in a Kafka topic, then turn on **Vault transit encryption**
for a channel and show the same topic now carrying `vault:v1:…` ciphertext.

---

## Where the Vault credentials live

Both files are **gitignored** (mode `600`) under `openshift/platform/`:

| File | Contents | Use |
|------|----------|-----|
| `.vault-init.json` | Vault **root token** + **5 unseal keys** (Shamir, threshold 3) | Admin / re-unseal if `vault-0` restarts sealed |
| `.vault-gotak-token` | **gotak-scoped token** (`default`+`gotak` policies, full `transit/*` rights, valid to 2026-07-12) | **This is what you paste into the app's Vault modal** |

> The app never ships a Vault token in its JavaScript. You paste the
> gotak token into the Communications "Configure Vault Encryption" modal;
> it is stored only in *that browser's* `localStorage`
> (`gotak_vault_token`) and used for encrypt-on-send.

If Vault comes up **sealed** (`vault.demoland.io/v1/sys/health` → 503):
```bash
export KUBECONFIG=iac/sno/cluster-auth/kubeconfig
for k in $(jq -r '.unseal_keys_b64[0,1,2]' openshift/platform/.vault-init.json); do
  oc exec -n gotak vault-0 -- vault operator unseal "$k"
done
```

---

## Pre-flight (30 seconds)

```bash
curl -sk https://gotak.demoland.io/            -o /dev/null -w 'gotak  %{http_code}\n'
curl -sk https://vault.demoland.io/v1/sys/health -o /dev/null -w 'vault  %{http_code}\n'   # 200 = unsealed
dig +short boundary.demoland.io                                                            # note the IP
```
All three healthy → you're ready. (Boundary's node IP changes on every sandbox
stop/start — if it moved, update the Namecheap A record before the demo.)

---

## The demo

### 1 — Open the two surfaces

- **GoTAK** → <https://gotak.demoland.io> — log in as `admin` (password in 1Password).
- **Redpanda Console** is **Boundary-only** (no public route). Connect through Boundary:
  ```bash
  boundary connect -target-name redpanda-console -target-scope-name gotak
  # → prints a local 127.0.0.1:<port>; open http://127.0.0.1:<port> in a browser
  ```
  (Or use Boundary Desktop → target **redpanda-console** → Connect.)
  Boundary brokers to the console NodePort `10.0.1.52:30851`.

### 2 — Show raw PLAINTEXT in Kafka

1. In GoTAK, open the **Communications** tab and select a channel.
2. Type a message and send it. No encryption is configured yet → plaintext.
3. In Redpanda Console → **Topics** → `gotak.comms.<RoomID>` → **Messages**.
   The topic auto-creates on first send. You'll see your message **in the clear**.

> Topics live on an `emptyDir`, so they reset whenever the `kafka` pod
> restarts. A fresh broker shows `topic_count: 0` until the first message.

### 3 — Turn ON encryption for the channel

1. In the channel, open **Configure Vault Encryption**.
2. Fill in:
   - **Vault address**: `https://vault.demoland.io`
   - **Token**: paste the gotak token from `.vault-gotak-token`
     (`hvs.CAESIPVTRejJ…`)
   - **Key name**: leave the default (per-channel transit key).
3. Click enable. This creates the transit key in Vault **and** remembers the
   token in your browser for encrypt-on-send.

### 4 — Show the SAME topic is now ENCRYPTED

1. Send another message in that channel.
2. In Redpanda Console → same `gotak.comms.<RoomID>` topic → the new record's
   value is now **`vault:v1:<base64-ciphertext>`** instead of readable text.

Side-by-side, the topic shows the earlier plaintext record and the new
ciphertext record — the encryption boundary is visible in the broker itself.

---

## Reference — access map

| Surface | URL / target | Exposure |
|---------|--------------|----------|
| GoTAK app | `https://gotak.demoland.io` | Public (edge TLS, Sectigo) |
| Vault | `https://vault.demoland.io` | Public (browser-direct client) |
| Consul UI | Boundary target `consul-ui` → `10.0.1.52:30850` | Boundary-only |
| Redpanda Console | Boundary target `redpanda-console` → `10.0.1.52:30851` | Boundary-only |
| Postgres | Boundary target `postgres` → `10.0.1.52:30432` | Boundary-only |
| Boundary | `boundary.demoland.io:9200`, login `gotak` | Public controller |

- Chat publish path: `gotak-server` → Kafka topic `gotak.comms.<RoomID>`
  (best-effort; `KAFKA_BROKERS=kafka:9092`). When a channel is encrypted, the
  message text is ciphertext *before* it reaches Kafka.
- Boundary login password: `/etc/boundary/login-pass.txt` on the controller
  node; recovery/root creds: `/etc/boundary/init-output.txt`.
