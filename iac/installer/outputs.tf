output "instance_id" {
  description = "Installer node instance ID."
  value       = aws_instance.installer.id
}

output "public_ip" {
  description = "Public IP of the installer node."
  value       = aws_instance.installer.public_ip
}

output "ssh_command" {
  description = "SSH command (key: the dfed01 private key matching var.ssh_public_key)."
  value       = "ssh -i ~/.ssh/dfed01 ec2-user@${aws_instance.installer.public_ip}"
}

output "iam_role_arn" {
  description = "IAM role the node's instance profile uses (the installer's AWS identity)."
  value       = aws_iam_role.installer.arn
}

output "ami_id" {
  description = "AMI the node booted from."
  value       = aws_instance.installer.ami
}
