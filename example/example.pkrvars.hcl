# Example variables file for VergeIO Packer plugin
# Copy this to local.pkrvars.hcl and customize for your environment
# Usage: packer build -var-file="local.pkrvars.hcl" .

# VergeIO cluster connection settings
vergeio_endpoint = "your-vergeio-cluster.example.com"
vergeio_username = "your-username"
vergeio_password = "your-password"
vergeio_insecure = true
vergeio_port     = 443

# Network configuration settings
static_ip_address = "192.168.1.100/24"
gateway_ip       = "192.168.1.1"
dns_servers      = ["8.8.8.8", "1.1.1.1"]