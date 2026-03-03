#!/bin/bash
# agent-init.sh.tpl

# Disable firewall (Oracle Linux default)
systemctl stop firewalld
systemctl disable firewalld

# Install k3s agent
curl -sfL https://get.k3s.io | K3S_URL=https://${server_ip}:6443 K3S_TOKEN=${k3s_token} sh -
