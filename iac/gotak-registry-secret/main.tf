# ghcr.io image pull secret for the gotak namespace.
# The PAT comes in as a sensitive variable; the dockerconfigjson is assembled here.
# (The value lands in TFC state, which is Vault-encrypted at rest and access-scoped.)
resource "kubernetes_secret" "ghcr" {
  metadata {
    name      = var.secret_name
    namespace = var.namespace
    labels = {
      "app.kubernetes.io/name"      = "gotak"
      "app.kubernetes.io/component" = "registry-pull"
    }
  }

  type = "kubernetes.io/dockerconfigjson"

  data = {
    ".dockerconfigjson" = jsonencode({
      auths = {
        (var.registry_server) = {
          username = var.ghcr_username
          password = var.ghcr_pat
          auth     = base64encode("${var.ghcr_username}:${var.ghcr_pat}")
        }
      }
    })
  }
}

output "secret_name" {
  description = "Reference this from the deployments' imagePullSecrets."
  value       = kubernetes_secret.ghcr.metadata[0].name
}
