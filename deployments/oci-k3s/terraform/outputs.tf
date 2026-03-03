output "k3s_server_public_ip" {
  value       = oci_core_instance.k3s_nodes[0].public_ip
  description = "The public IP of the K3s server (Control Plane)"
}

output "k3s_agent_public_ips" {
  value       = [for i in oci_core_instance.k3s_nodes : i.public_ip if i.display_name != "k3s-server"]
  description = "The public IPs of the K3s agents"
}

output "ssh_command" {
  value       = "ssh -i ${var.ssh_private_key_path} opc@${oci_core_instance.k3s_nodes[0].public_ip}"
  description = "SSH command to connect to the K3s server"
}
