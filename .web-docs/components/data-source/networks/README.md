# VergeIO Networks Data Source

The VergeIO Networks data source allows you to discover VergeIO virtual networks by name or type, eliminating the need for hardcoded network IDs and making Packer configurations portable between different VergeIO clusters.

This data source queries the VergeIO API to retrieve network information that can be used in VM network interface configurations.

## Configuration Reference

**Required:**

- `vergeio_endpoint` (string) - The VergeIO cluster endpoint URL (e.g., `https://cluster.example.com`)
- `vergeio_username` (string) - Username for VergeIO cluster authentication  
- `vergeio_password` (string) - Password for VergeIO cluster authentication

**Optional:**

### Connection Configuration

- `vergeio_port` (int) - VergeIO cluster port. Defaults to `443`
- `vergeio_insecure` (bool) - Skip TLS certificate verification. Defaults to `false`

### Filter Options

- `filter_name` (string) - Filter networks by exact name match
- `filter_type` (string) - Filter networks by type (e.g., `switch`, `vlan`)

## Output Attributes

- `networks` (list) - List of networks matching the filter criteria. Each network contains:
  - `id` (int) - The network ID for use in VM NIC configurations
  - `name` (string) - The network name
  - `description` (string) - The network description

## Example Usage

### Basic Network Discovery

```hcl
data "vergeio" "external_network" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  
  filter_name = "External"
}

source "vergeio" "web-server" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username  
  vergeio_password = var.vergeio_password
  
  name = "web-server-vm"
  cpu_cores = 2
  ram = 2048
  
  # Use discovered network ID instead of hardcoded value
  vm_nics {
    name  = "primary_nic"
    vnet  = data.vergeio.external_network.networks[0].id
    driver = "virtio"
  }
}

build {
  sources = ["source.vergeio.web-server"]
}
```

### Multi-Tier Network Architecture

```hcl
# Discover networks for each tier
data "vergeio" "web_network" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  
  filter_name = "Web-Tier"
}

data "vergeio" "app_network" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  
  filter_name = "App-Tier"
}

data "vergeio" "db_network" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  
  filter_name = "DB-Tier"
}

# Web server with dual NICs
source "vergeio" "web-server" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  
  name = "web-server"
  
  vm_nics {
    name = "web_nic"
    vnet = data.vergeio.web_network.networks[0].id
    driver = "virtio"
  }
  
  vm_nics {
    name = "app_nic" 
    vnet = data.vergeio.app_network.networks[0].id
    driver = "virtio"
  }
}

# Database server on dedicated network
source "vergeio" "database" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  
  name = "database-server"
  
  vm_nics {
    name = "db_nic"
    vnet = data.vergeio.db_network.networks[0].id
    driver = "virtio"
  }
}

build {
  sources = ["source.vergeio.web-server", "source.vergeio.database"]
}
```

### Network Discovery and Selection

```hcl
# Discover all available networks
data "vergeio" "all_networks" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  
  # No filters - returns all networks
}

# Filter networks by type
data "vergeio" "switch_networks" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  
  filter_type = "switch"
}

locals {
  # Find management network from discovered networks
  management_network = [
    for network in data.vergeio.all_networks.networks :
    network if network.name == "Management"
  ][0]
  
  # Get production networks
  prod_networks = [
    for network in data.vergeio.all_networks.networks :
    network if can(regex("^prod-", lower(network.name)))
  ]
}

source "vergeio" "management-vm" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  
  name = "management-vm"
  
  vm_nics {
    name = "mgmt_nic"
    vnet = local.management_network.id
    driver = "virtio"
  }
}

# Output discovered networks for reference
output "available_networks" {
  value = {
    all_networks    = data.vergeio.all_networks.networks
    switch_networks = data.vergeio.switch_networks.networks
    prod_networks   = local.prod_networks
  }
}
```

### Environment-Portable Configuration

```hcl
variable "environment" {
  type        = string
  description = "Environment name (dev, staging, prod)"
}

# Discover environment-specific network
data "vergeio" "app_network" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  
  filter_name = "${var.environment}-application"
}

source "vergeio" "app-server" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  
  name = "${var.environment}-app-server"
  
  vm_nics {
    name = "app_nic"
    vnet = data.vergeio.app_network.networks[0].id
    driver = "virtio"
  }
}

build {
  sources = ["source.vergeio.app-server"]
}
```

### Network Validation

```hcl
data "vergeio" "required_network" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  
  filter_name = var.network_name
}

# Validate that the required network exists
locals {
  validate_network = length(data.vergeio.required_network.networks) > 0 ? data.vergeio.required_network.networks[0] : null
}

source "vergeio" "validated-vm" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  
  name = "validated-vm"
  
  vm_nics {
    name = "primary_nic"
    vnet = local.validate_network != null ? local.validate_network.id : null
    driver = "virtio"
  }
}
```

## Features

- **Dynamic Discovery**: Query networks at build time for up-to-date information
- **Portable Configurations**: Use network names instead of IDs for environment portability  
- **Flexible Filtering**: Filter by exact name match or network type
- **Validation**: Runtime validation ensures networks exist before VM creation
- **Multi-Network Support**: Discover and use multiple networks in a single configuration

## Common Use Cases

### Environment Promotion
Configurations work seamlessly across development, staging, and production environments when networks follow consistent naming conventions.

### Multi-Tier Applications  
Discover separate networks for web, application, and database tiers for proper network segmentation.

### Network Documentation
Query all available networks to understand the network topology and available options.

### Disaster Recovery
Network discovery enables consistent configurations across primary and DR sites.

## Notes

- Network discovery happens at build time, ensuring current network information
- Filter by name performs exact string matching (case-sensitive)
- When no filters are specified, all networks in the cluster are returned
- The data source requires read permissions for networks in the VergeIO cluster
- Network IDs returned by the data source can be used directly in `vm_nics` configurations
- If a filter returns no results, the `networks` list will be empty