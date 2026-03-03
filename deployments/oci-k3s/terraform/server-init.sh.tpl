#!/bin/bash
# server-init.sh.tpl

# Disable firewall (Oracle Linux default)
systemctl stop firewalld
systemctl disable firewalld

# Install k3s server
curl -sfL https://get.k3s.io | K3S_TOKEN=${k3s_token} sh -s - server \
  --cluster-init
