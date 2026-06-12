# Layer 4 — deploy GoTAK onto the secured cluster.
# Reuses the manifests in ../../openshift (kustomize), wired to the platform services.

# resource "kubernetes_manifest" "gotak" { ... }   # or a thin helm chart wrapping ../../openshift
#
# GoTAK config (server.yaml) is rendered to point at:
#   - Vault (OIDC token validation + dynamic DB creds)
#   - Consul (service registration / Connect)
#   - Boundary (brokered access to the query target)
#   - Keycloak realm "gotak" (agent OIDC)
#
# Result: GoTAK comes up already inside the zero-trust mesh.
