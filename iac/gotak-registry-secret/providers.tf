# Connect to the cluster via the named-cert hostname (sno.demoland.io), whose
# Sectigo wildcard cert is publicly trusted — so no cluster CA needs to be passed
# and TFC's remote runner can reach it over the internet. Auth is a scoped SA
# token (see bootstrap-sa.sh).
provider "kubernetes" {
  host  = var.k8s_host
  token = var.k8s_token
}
