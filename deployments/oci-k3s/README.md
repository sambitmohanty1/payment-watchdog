# Payment Watchdog - OCI Free Tier Deployment

This directory contains the necessary scripts and configuration to deploy the Payment Watchdog application to Oracle Cloud Infrastructure (OCI) Free Tier using a lightweight Kubernetes cluster (K3s).

## Architecture overview
- **Infrastructure (Terraform)**: Provisions 2 (configurable) ARM-based OCI virtual machines in a dedicated Virtual Cloud Network (VCN). It sets up Security Lists (firewalls) to allow web traffic (80/443), SSH (22), and Kubernetes API access (6443).
- **Cluster (K3s)**: Lightweight Kubernetes installed automatically via Cloud-Init on the newly created VMs.
- **Ingress & SSL (Traefik + cert-manager)**: Routes incoming traffic to the appropriate service (Web UI or API) and automatically issues Let's Encrypt SSL certificates.
- **Application (Kustomize)**: Leverages the existing Kubernetes manifests located in `api/deployments/kubernetes/` but overlays them with an OCI-specific `ingress.yaml` to handle domain routing and Traefik configuration.

## Prerequisites
1. **OCI Account**: You need an active Oracle Cloud account.
2. **OCI CLI Setup**: Generate an API key pair from the OCI Console (User Settings -> API Keys).
3. **Terraform**: Install Terraform or OpenTofu on your local machine.
4. **Kubectl**: Ensure `kubectl` is installed locally to manage the cluster.
5. **Domain Name**: A domain name whose DNS A-records can be pointed to the new OCI server's public IP.

## Setup Instructions

### 1. Configure the Environment File
Copy the example environment file and fill it with your OCI and application details.
```bash
cp .env.example .env
nano .env
```
Ensure you provide correct paths to your SSH public key and OCI private key, as well as secure passwords for the database and Redis instances.

### 2. Deploy
Run the automated deployment script. This script will:
1. Initialize and apply the Terraform configuration.
2. Wait for the server to finish installing K3s.
3. Securely fetch the `kubeconfig` file from the remote server so you can run `kubectl` commands locally.
4. Use `kustomize` and `envsubst` to apply the application manifests directly to the cluster.

```bash
./deploy.sh
```

### 3. Configure DNS
Once the script completes, it will output the Public IP of your control plane node.
Update your domain's DNS settings to point an `A` record for your `DOMAIN` (e.g., `watchdog.yourdomain.com`) to this Public IP address.

### 4. Verify the Deployment
You can monitor the progress of the application startup using the securely downloaded kubeconfig:
```bash
export KUBECONFIG=~/.kube/oci_k3s.yaml
kubectl get pods -n payment-watchdog
```

Once all pods are running and DNS propagates, your application will be securely available at `https://<your-domain>`.

## Notes
- **Multi-architecture Docker Images**: The OCI Free Tier instances use ARM (`aarch64`) architecture. The project's GitHub Actions workflow has been updated to automatically build and push `linux/arm64` Docker images to the registry, so the Kubernetes nodes will correctly pull compatible images.
- **Database Storage**: PostgreSQL and Redis run inside the K3s cluster. Data is persisted using the server's local storage via standard PersistentVolumeClaims.
