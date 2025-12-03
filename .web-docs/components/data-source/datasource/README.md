# VergeIO Data Sources

VergeIO provides multiple data sources for discovering and querying resources within a VergeIO cluster. These data sources enable dynamic, portable configurations by resolving resource names to IDs at build time.

## Available Data Sources

### Networks Data Source

Query VergeIO virtual networks by name or type to obtain network IDs for VM NIC configurations.

**Use cases:**

- Portable configurations across environments
- Dynamic network discovery
- Multi-tier application deployments
- Environment-specific network selection

[→ Networks Data Source Documentation](/docs/datasources/vergeio-networks)

### VMs Data Source

Discover existing virtual machines for template discovery, configuration cloning, and inventory management.

**Use cases:**

- Finding base images and templates
- Cloning VM configurations
- Infrastructure documentation
- Snapshot analysis and management

[→ VMs Data Source Documentation](/docs/datasources/vergeio-vms)

## Common Configuration

All VergeIO data sources share common connection parameters:

```hcl
data "vergeio" "example" {
  # Required connection parameters
  vergeio_endpoint = var.vergeio_endpoint  # e.g., "https://cluster.example.com"
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password

  # Optional connection parameters
  vergeio_port     = 443    # Default HTTPS port
  vergeio_insecure = false  # Skip TLS verification

  # Data source specific filters...
}
```

## Best Practices

### Variable Management

Store connection details in variables for reusability:

```hcl
variable "vergeio_endpoint" {
  type        = string
  description = "VergeIO cluster endpoint"
}

variable "vergeio_username" {
  type        = string
  description = "VergeIO username"
}

variable "vergeio_password" {
  type        = string
  description = "VergeIO password"
  sensitive   = true
}
```

### Environment Portability

Use consistent naming conventions for portable configurations:

```hcl
data "vergeio" "app_network" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password

  # Works across dev/staging/prod with consistent naming
  filter_name = "${var.environment}-application"
}
```

### Error Handling

Validate that resources exist before using them:

```hcl
locals {
  network_found = length(data.vergeio.app_network.networks) > 0
  network_id    = local.network_found ? data.vergeio.app_network.networks[0].id : null
}
```

### Performance Optimization

Use specific filters to reduce query time and API load:

```hcl
# Good - specific filter
data "vergeio" "prod_network" {
  filter_name = "Production-Web"
}

# Less optimal - query all then filter locally
data "vergeio" "all_networks" {
  # No filters - returns everything
}
```

## Integration Examples

### Builder Integration

Use data sources to dynamically configure VM builders:

```hcl
data "vergeio" "web_network" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  filter_name = "Web-Tier"
}

source "vergeio" "web-server" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password

  vm_nics {
    vnet = data.vergeio.web_network.networks[0].id
    name = "web_nic"
    driver = "virtio"
  }
}
```

### Multi-Data Source Usage

Combine multiple data sources for complex configurations:

```hcl
# Discover network
data "vergeio" "app_network" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  filter_name = "Application"
}

# Find template VM
data "vergeio" "base_template" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  filter_name = "Ubuntu-22.04-Template"
}

source "vergeio" "app-server" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password

  # Use discovered network
  vm_nics {
    vnet = data.vergeio.app_network.networks[0].id
    name = "app_nic"
    driver = "virtio"
  }

  # Import from discovered template
  vm_disks {
    name         = "system"
    media        = "import"
    media_source = data.vergeio.base_template.vms[0].drives[0].media_source.key
    disksize     = 30
    interface    = "virtio"
  }
}
```

## Authentication & Permissions

### Required Permissions

Data sources require read access to query VergeIO resources:

- **Networks**: Read access to network objects
- **VMs**: Read access to VM objects and their configurations

### Credential Management

Use environment variables or credential files for security:

```bash
# Environment variables
export VERGEIO_ENDPOINT="https://cluster.example.com"
export VERGEIO_USERNAME="your-username"
export VERGEIO_PASSWORD="your-password"
```

```hcl
# Reference in configuration
data "vergeio" "example" {
  vergeio_endpoint = env("VERGEIO_ENDPOINT")
  vergeio_username = env("VERGEIO_USERNAME")
  vergeio_password = env("VERGEIO_PASSWORD")
}
```

## Troubleshooting

### Common Issues

**Empty Results**: Verify filter criteria and resource existence

```hcl
# Debug: Query without filters first
data "vergeio" "debug_all" {
  # No filters - see all available resources
}

output "debug_results" {
  value = data.vergeio.debug_all
}
```

**Connection Errors**: Check endpoint URL and network connectivity

- Ensure HTTPS URL format: `https://cluster.example.com`
- Verify port accessibility (default 443)
- Check for firewall or proxy issues

**Authentication Errors**: Verify credentials and permissions

- Test credentials with VergeIO web interface
- Ensure user has appropriate read permissions
- Check for account lockouts or password expiration

**Performance Issues**: Optimize queries with specific filters

- Use exact name filters when possible
- Avoid querying large datasets unnecessarily
- Consider caching results for repeated builds

## Notes

- Data source queries execute at build time, not template parse time
- Results reflect current state of the VergeIO cluster
- Network timeouts may occur with slow or overloaded clusters
- Cached results are not shared between different data source instances
- Changes to VergeIO resources between builds may affect reproducibility
