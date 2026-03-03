variable "tenancy_ocid" {
  description = "The tenancy OCID"
  type        = string
}

variable "user_ocid" {
  description = "The user OCID"
  type        = string
}

variable "fingerprint" {
  description = "The fingerprint of the API key"
  type        = string
}

variable "private_key_path" {
  description = "The path to the API private key"
  type        = string
}

variable "region" {
  description = "The OCI region"
  type        = string
}

variable "compartment_ocid" {
  description = "The compartment OCID"
  type        = string
}

variable "ssh_public_key" {
  description = "The SSH public key for the instances"
  type        = string
}

variable "ssh_private_key_path" {
  description = "The path to the SSH private key used to connect to the instances"
  type        = string
}

variable "instance_count" {
  description = "Number of instances to create (e.g., 2 or 3 small VMs)"
  type        = number
  default     = 2
}

variable "instance_ocpus" {
  description = "Number of OCPUs per instance"
  type        = number
  default     = 2
}

variable "instance_memory_in_gbs" {
  description = "Amount of memory per instance in GBs"
  type        = number
  default     = 12
}

variable "k3s_token" {
  description = "Secret token for K3s cluster join"
  type        = string
  sensitive   = true
}
