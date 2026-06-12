variable "k8s_host" {
  description = "Cluster API URL. Use the named-cert host so the public Sectigo cert validates without a CA."
  type        = string
  default     = "https://sno.demoland.io:6443"
}

variable "k8s_token" {
  description = "Token for a ServiceAccount allowed to manage secrets in the namespace (see bootstrap-sa.sh)."
  type        = string
  sensitive   = true
}

variable "namespace" {
  description = "Namespace to create the pull secret in."
  type        = string
  default     = "gotak"
}

variable "secret_name" {
  description = "Name of the docker-registry pull secret. Referenced by the deployments' imagePullSecrets."
  type        = string
  default     = "ghcr"
}

variable "registry_server" {
  description = "Container registry host."
  type        = string
  default     = "ghcr.io"
}

variable "ghcr_username" {
  description = "GitHub username/org that owns the packages."
  type        = string
}

variable "ghcr_pat" {
  description = "GitHub PAT with read:packages. Mark this SENSITIVE in TFC."
  type        = string
  sensitive   = true
}
