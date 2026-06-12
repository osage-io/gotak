# Providers for the GoTAK secure-runtime demo stack.
# Bob generates the resource detail into the layer files (00-40); this pins the providers.

terraform {
  required_version = ">= 1.6"
  required_providers {
    aws        = { source = "hashicorp/aws", version = "~> 5.0" }
    rhcs       = { source = "terraform-redhat/rhcs", version = "~> 1.6" } # Red Hat OpenShift Service on AWS (ROSA)
    kubernetes = { source = "hashicorp/kubernetes", version = "~> 2.30" }
    helm       = { source = "hashicorp/helm", version = "~> 2.13" }
    vault      = { source = "hashicorp/vault", version = "~> 4.0" }
    consul     = { source = "hashicorp/consul", version = "~> 2.20" }
    boundary   = { source = "hashicorp/boundary", version = "~> 1.1" }
    keycloak   = { source = "mrparkers/keycloak", version = "~> 4.4" }
  }
}

provider "aws" {
  region = var.aws_region
}

# The kubernetes/helm providers point at the OpenShift cluster created in 10-cluster.tf.
# After the cluster exists, wire these to its endpoint + token (kept abstract here).
provider "kubernetes" {
  host                   = try(module.cluster.api_url, "")
  token                  = try(module.cluster.token, "")
  cluster_ca_certificate = try(module.cluster.ca_cert, "")
}

provider "helm" {
  kubernetes {
    host                   = try(module.cluster.api_url, "")
    token                  = try(module.cluster.token, "")
    cluster_ca_certificate = try(module.cluster.ca_cert, "")
  }
}

# Security-layer providers (30-security.tf) point at the in-cluster services once 20-platform is up.
provider "vault" {
  address = var.vault_addr
}

provider "boundary" {
  addr = var.boundary_addr
}

provider "keycloak" {
  url       = var.keycloak_url
  client_id = "admin-cli"
}
