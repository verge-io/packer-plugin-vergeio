# VergeIO VMs Data Source

The VergeIO VMs data source allows you to discover existing virtual machines in a VergeIO cluster by name, ID, or other criteria. This is useful for finding base images, template VMs, or existing infrastructure to use as source material for new builds.

This data source queries the VergeIO API to retrieve VM information including hardware configuration, storage details, and network interfaces.

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

- `filter_name` (string) - Filter VMs by exact name match
- `filter_id` (int) - Filter VMs by specific ID
- `is_snapshot` (bool) - Filter to include only snapshots (`true`) or exclude snapshots (`false`)

## Output Attributes

- `vms` (list) - List of VMs matching the filter criteria. Each VM contains:
  - `id` (int) - The VM ID
  - `name` (string) - The VM name  
  - `key` (int) - The VM key for API operations
  - `is_snapshot` (bool) - Whether this VM is a snapshot
  - `cpu_type` (string) - CPU type/model
  - `machine_type` (string) - Machine type
  - `os_family` (string) - Operating system family
  - `uefi` (bool) - Whether UEFI boot is enabled
  - `drives` (list) - List of VM drives with details:
    - `key` (int) - Drive key
    - `name` (string) - Drive name
    - `interface` (string) - Drive interface (virtio, ide, etc.)
    - `media` (string) - Media type (disk, cdrom, etc.)
    - `description` (string) - Drive description
    - `preferred_tier` (string) - Storage tier preference
    - `media_source` (object) - Media source information if applicable
  - `nics` (list) - List of network interfaces with details:
    - `key` (int) - NIC key
    - `name` (string) - NIC name
    - `interface` (string) - NIC interface type
    - `driver` (string) - NIC driver
    - `vnet` (int) - Virtual network ID
    - `ipaddress` (string) - Assigned IP address (if any)
    - `macaddress` (string) - MAC address

## Example Usage

### Find Template VMs

```hcl
# Find all template VMs (non-snapshots)
data "vergeio-vms" "templates" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  
  is_snapshot = false
}

locals {
  # Find Ubuntu templates
  ubuntu_templates = [
    for vm in data.vergeio-vms.templates.vms :
    vm if can(regex("ubuntu", lower(vm.name)))
  ]
  
  # Find the latest Ubuntu template
  ubuntu_22_04 = [
    for vm in local.ubuntu_templates :
    vm if can(regex("22.04", vm.name))
  ][0]
}

source "vergeio" "web-server" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  
  name = "web-server-from-template"
  
  # Import the discovered template as base
  vm_disks {
    name         = "system-disk"
    disksize     = 50
    interface    = "virtio"
    media        = "import"
    media_source = local.ubuntu_22_04.drives[0].media_source.key
  }
}
```

### Clone Existing VM Configuration

```hcl
# Find a specific VM to clone its configuration
data "vergeio-vms" "source_vm" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  
  filter_name = "production-web-server"
}

locals {
  source_vm = data.vergeio-vms.source_vm.vms[0]
}

source "vergeio" "development-clone" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  
  # Clone the configuration of the source VM
  name         = "dev-${local.source_vm.name}"
  cpu_type     = local.source_vm.cpu_type
  machine_type = local.source_vm.machine_type
  uefi         = local.source_vm.uefi
  
  # Clone storage configuration
  dynamic "vm_disks" {
    for_each = local.source_vm.drives
    content {
      name         = "${vm_disks.value.name}-dev"
      interface    = vm_disks.value.interface
      media        = vm_disks.value.media
      media_source = vm_disks.value.media_source != null ? vm_disks.value.media_source.key : null
      disksize     = 20  # Smaller for development
    }
  }
  
  # Clone network configuration
  dynamic "vm_nics" {
    for_each = local.source_vm.nics
    content {
      name   = vm_nics.value.name
      vnet   = vm_nics.value.vnet
      driver = vm_nics.value.driver
    }
  }
}
```

### Inventory and Documentation

```hcl
# Get all VMs for inventory
data "vergeio-vms" "all_vms" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
}

# Get all snapshots for backup analysis
data "vergeio-vms" "snapshots" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  
  is_snapshot = true
}

locals {
  # Analyze VM inventory
  vm_stats = {
    total_vms = length(data.vergeio-vms.all_vms.vms)
    snapshots = length(data.vergeio-vms.snapshots.vms)
    
    by_os_family = {
      for os in distinct([for vm in data.vergeio-vms.all_vms.vms : vm.os_family]) :
      os => length([for vm in data.vergeio-vms.all_vms.vms : vm if vm.os_family == os])
    }
    
    uefi_enabled = length([
      for vm in data.vergeio-vms.all_vms.vms : vm if vm.uefi == true
    ])
  }
  
  # Find Windows VMs
  windows_vms = [
    for vm in data.vergeio-vms.all_vms.vms :
    vm if can(regex("windows", lower(vm.os_family)))
  ]
  
  # Find VMs with multiple NICs
  multi_nic_vms = [
    for vm in data.vergeio-vms.all_vms.vms :
    vm if length(vm.nics) > 1
  ]
}

# Output inventory for documentation
output "vm_inventory" {
  value = {
    statistics     = local.vm_stats
    windows_vms    = local.windows_vms
    multi_nic_vms  = local.multi_nic_vms
    all_vms        = data.vergeio-vms.all_vms.vms
  }
}
```

### Find Source Images by Criteria

```hcl
# Find VMs with specific characteristics for use as base images
data "vergeio-vms" "base_images" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  
  is_snapshot = false
}

locals {
  # Find UEFI-enabled VMs suitable for modern deployments
  uefi_vms = [
    for vm in data.vergeio-vms.base_images.vms :
    vm if vm.uefi == true
  ]
  
  # Find VMs with virtio storage for performance
  virtio_storage_vms = [
    for vm in data.vergeio-vms.base_images.vms :
    vm if length([
      for drive in vm.drives :
      drive if drive.interface == "virtio"
    ]) > 0
  ]
  
  # Find VMs with single NIC for simple deployments
  single_nic_vms = [
    for vm in data.vergeio-vms.base_images.vms :
    vm if length(vm.nics) == 1
  ]
  
  # Combine criteria to find optimal base images
  optimal_base_images = [
    for vm in local.uefi_vms :
    vm if contains([for v in local.virtio_storage_vms : v.id], vm.id) &&
          contains([for v in local.single_nic_vms : v.id], vm.id)
  ]
}

source "vergeio" "optimized-vm" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  
  name = "optimized-from-base"
  
  # Use the first optimal base image found
  vm_disks {
    name         = "system-disk"
    disksize     = 30
    interface    = "virtio"
    media        = "import"
    media_source = local.optimal_base_images[0].drives[0].media_source.key
  }
  
  # Copy NIC configuration from base
  vm_nics {
    name   = local.optimal_base_images[0].nics[0].name
    vnet   = local.optimal_base_images[0].nics[0].vnet
    driver = "virtio"
  }
}
```

### Validate VM Existence

```hcl
variable "source_vm_name" {
  type        = string
  description = "Name of VM to use as source"
}

# Check if the specified VM exists
data "vergeio-vms" "source_validation" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  
  filter_name = var.source_vm_name
}

locals {
  source_vm_exists = length(data.vergeio-vms.source_validation.vms) > 0
  source_vm = local.source_vm_exists ? data.vergeio-vms.source_validation.vms[0] : null
}

# Only proceed if source VM exists
source "vergeio" "conditional-build" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  
  name = "built-from-${var.source_vm_name}"
  
  vm_disks {
    name         = "system-disk"
    disksize     = 25
    interface    = "virtio"
    media        = "import"
    media_source = local.source_vm != null ? local.source_vm.drives[0].media_source.key : null
  }
  
  vm_nics {
    name   = "primary_nic"
    vnet   = local.source_vm != null ? local.source_vm.nics[0].vnet : 1
    driver = "virtio"
  }
}

# Output validation results
output "source_vm_validation" {
  value = {
    exists     = local.source_vm_exists
    source_vm  = local.source_vm
    message    = local.source_vm_exists ? "Source VM found" : "Source VM '${var.source_vm_name}' not found"
  }
}
```

## Features

- **VM Discovery**: Find existing VMs by name, ID, or characteristics
- **Template Discovery**: Locate suitable base images and templates
- **Configuration Cloning**: Copy hardware and network configuration from existing VMs
- **Inventory Management**: Query all VMs for documentation and analysis
- **Snapshot Analysis**: Discover and analyze VM snapshots
- **Validation**: Verify VM existence before using in builds

## Common Use Cases

### Template Management
Discover available VM templates and base images for consistent deployments across environments.

### Configuration Standardization  
Clone proven VM configurations to ensure consistency and reduce configuration errors.

### Infrastructure Documentation
Query VM inventory for documentation, compliance, and capacity planning.

### Disaster Recovery
Identify source VMs and snapshots for disaster recovery planning and testing.

## Notes

- VM discovery happens at build time, providing current VM information
- The data source returns detailed VM configuration including hardware and network details
- Storage media source keys can be used directly in `media_source` fields for disk imports
- Network VNET IDs can be used directly in VM NIC configurations
- Filtering by name performs exact string matching (case-sensitive)
- When no filters are specified, all VMs (including snapshots) are returned
- The `is_snapshot` filter helps distinguish between regular VMs and snapshot VMs
- Large clusters may have many VMs; use filters to improve query performance
- The data source requires read permissions for VMs in the VergeIO cluster