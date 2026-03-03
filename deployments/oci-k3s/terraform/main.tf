terraform {
  required_providers {
    oci = {
      source  = "oracle/oci"
      version = "~> 5.0"
    }
  }
}

provider "oci" {
  tenancy_ocid     = var.tenancy_ocid
  user_ocid        = var.user_ocid
  fingerprint      = var.fingerprint
  private_key_path = var.private_key_path
  region           = var.region
}

# Networking
resource "oci_core_vcn" "k3s_vcn" {
  compartment_id = var.compartment_ocid
  display_name   = "k3s-vcn"
  cidr_block     = "10.0.0.0/16"
}

resource "oci_core_internet_gateway" "k3s_ig" {
  compartment_id = var.compartment_ocid
  vcn_id         = oci_core_vcn.k3s_vcn.id
  display_name   = "k3s-ig"
  enabled        = true
}

resource "oci_core_route_table" "k3s_rt" {
  compartment_id = var.compartment_ocid
  vcn_id         = oci_core_vcn.k3s_vcn.id
  display_name   = "k3s-rt"

  route_rules {
    destination       = "0.0.0.0/0"
    destination_type  = "CIDR_BLOCK"
    network_entity_id = oci_core_internet_gateway.k3s_ig.id
  }
}

resource "oci_core_security_list" "k3s_sl" {
  compartment_id = var.compartment_ocid
  vcn_id         = oci_core_vcn.k3s_vcn.id
  display_name   = "k3s-sl"

  egress_security_rules {
    destination = "0.0.0.0/0"
    protocol    = "all"
  }

  ingress_security_rules {
    protocol = "all"
    source   = "10.0.0.0/16"
  }

  ingress_security_rules {
    protocol = "6"
    source   = "0.0.0.0/0"
    tcp_options {
      min = 22
      max = 22
    }
  }

  ingress_security_rules {
    protocol = "6"
    source   = "0.0.0.0/0"
    tcp_options {
      min = 80
      max = 80
    }
  }

  ingress_security_rules {
    protocol = "6"
    source   = "0.0.0.0/0"
    tcp_options {
      min = 443
      max = 443
    }
  }

  ingress_security_rules {
    protocol = "6"
    source   = "0.0.0.0/0"
    tcp_options {
      min = 6443
      max = 6443
    }
  }
}

resource "oci_core_subnet" "k3s_subnet" {
  compartment_id    = var.compartment_ocid
  vcn_id            = oci_core_vcn.k3s_vcn.id
  cidr_block        = "10.0.1.0/24"
  display_name      = "k3s-subnet"
  route_table_id    = oci_core_route_table.k3s_rt.id
  security_list_ids = [oci_core_security_list.k3s_sl.id]
}

# Compute Instances
data "oci_identity_availability_domains" "ads" {
  compartment_id = var.compartment_ocid
}

# Oracle Linux 8 aarch64
data "oci_core_images" "arm_image" {
  compartment_id           = var.compartment_ocid
  operating_system         = "Oracle Linux"
  operating_system_version = "8"
  shape                    = "VM.Standard.A1.Flex"
  sort_by                  = "TIMECREATED"
  sort_order               = "DESC"
}

resource "oci_core_instance" "k3s_nodes" {
  count               = var.instance_count
  availability_domain = data.oci_identity_availability_domains.ads.availability_domains[0].name
  compartment_id      = var.compartment_ocid
  shape               = "VM.Standard.A1.Flex"
  display_name        = count.index == 0 ? "k3s-server" : "k3s-agent-${count.index}"

  shape_config {
    ocpus         = var.instance_ocpus
    memory_in_gbs = var.instance_memory_in_gbs
  }

  create_vnic_details {
    subnet_id        = oci_core_subnet.k3s_subnet.id
    assign_public_ip = true
  }

  source_details {
    source_type = "image"
    source_id   = data.oci_core_images.arm_image.images[0].id
  }

  metadata = {
    ssh_authorized_keys = var.ssh_public_key
    user_data = base64encode(count.index == 0 ? templatefile("${path.module}/server-init.sh.tpl", {
      k3s_token = var.k3s_token
    }) : templatefile("${path.module}/agent-init.sh.tpl", {
      k3s_token  = var.k3s_token
      server_ip  = oci_core_instance.k3s_nodes[0].private_ip
    }))
  }
}
