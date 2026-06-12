# Layer 3 — the zero-trust wiring. THIS is the "secures the app automatically" layer.
# Bob generates these against the live platform services so GoTAK is born enrolled.

# --- Keycloak: the agent identity ---
# resource "keycloak_realm" "gotak" { realm = "gotak" }
# resource "keycloak_openid_client" "agent" {
#   realm_id = keycloak_realm.gotak.id
#   client_id = "gotak-agent"
#   access_type = "CONFIDENTIAL"
#   service_accounts_enabled = true   # client-credentials for agent registration
# }

# --- Vault: federate to Keycloak + issue scoped dynamic creds ---
# resource "vault_jwt_auth_backend" "oidc" {
#   path = "oidc"; type = "oidc"
#   oidc_discovery_url = "${var.keycloak_url}/realms/gotak"
# }
# resource "vault_policy" "agent" { name = "gotak-agent"; policy = "..." }
# resource "vault_database_secret_backend_role" "gotak" { ... }  # dynamic Postgres creds

# --- Boundary: identity-based brokering to the query target ---
# resource "boundary_scope" "gotak" { ... }
# resource "boundary_auth_method_oidc" "keycloak" { ... }   # same realm
# resource "boundary_target" "query_target" { ... }

# --- Consul: mesh intentions (who may talk to whom) ---
# resource "consul_config_entry" "gotak_intentions" { kind = "service-intentions"; ... }
