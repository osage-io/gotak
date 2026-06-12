# Layer 2 — platform services on the cluster (Helm). Bob generates the release values.
# These are the zero-trust control plane GoTAK will be enrolled into.

# resource "helm_release" "vault" {
#   name = "vault"; namespace = "vault"; create_namespace = true
#   repository = "https://helm.releases.hashicorp.com"; chart = "vault"
#   # HA + integrated storage; injector on; OpenShift = true
# }

# resource "helm_release" "consul" {
#   name = "consul"; namespace = "consul"; create_namespace = true
#   repository = "https://helm.releases.hashicorp.com"; chart = "consul"
#   # connectInject.enabled = true; ui.enabled = true; openshift.enabled = true
# }

# resource "helm_release" "boundary" {
#   name = "boundary"; namespace = "boundary"; create_namespace = true
#   repository = "https://helm.releases.hashicorp.com"; chart = "boundary"
# }

# resource "helm_release" "keycloak" {
#   name = "keycloak"; namespace = "keycloak"; create_namespace = true
#   repository = "oci://registry-1.docker.io/bitnamicharts"; chart = "keycloak"
# }
