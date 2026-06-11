output "cluster_id" {
  description = "ROSA cluster ID."
  value       = module.rosa.cluster_id
}

output "cluster_name" {
  description = "ROSA cluster name."
  value       = var.cluster_name
}

output "api_url" {
  description = "OpenShift API server URL."
  value       = module.rosa.api_url
}

output "console_url" {
  description = "OpenShift web console URL."
  value       = module.rosa.console_url
}

output "domain" {
  description = "Cluster DNS domain."
  value       = module.rosa.domain
}

output "state" {
  description = "Cluster state."
  value       = module.rosa.state
}
