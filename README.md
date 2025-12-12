# Packer Plugin VergeIO

A production-ready Packer plugin for creating and provisioning virtual machines on VergeIO virtualization platform. This plugin provides complete VM lifecycle management with advanced features for enterprise environments.

## Features

**Core VM Management:**

- **VM Creation**: Complete VM creation with hardware, storage, and network configuration
- **Intelligent Power Management**: Real power state verification with configurable timeouts
- **Disk Import Support**: Automatic waiting for disk imports to complete before VM power-on
- **Network Discovery**: Dynamic network resolution by name using data sources
- **Cloud-init Integration**: Full cloud-init support with inline contents and external file loading
- **Multi-platform**: Supports Linux, Windows, and other operating systems

**Advanced Capabilities:**

- **SSH/WinRM Connectivity**: Optimized boot process leveraging Packer's communicator for reliability
- **Static IP Configuration**: Automatic extraction from cloud-init network configuration
- **Graceful Shutdown**: 4-phase shutdown process with power state verification
- **Network Data Source**: Query VergeIO networks by name for portable configurations
- **Variable Management**: Centralized credential management with sensitive value protection
- **Error Recovery**: Comprehensive rollback and cleanup on failures

**Enterprise Ready:**

- **Production Tested**: Used in live VergeIO environments
- **Comprehensive Logging**: Detailed debug output for troubleshooting
- **API Integration**: Full VergeIO API v4 support with proper error handling

## Components

- **Builder** ([builder/vergeio](builder/vergeio)) - Creates VMs with complete hardware configuration
- **Network Data Source** ([datasource/vergeio](datasource/vergeio)) - Discovers network IDs by name
- **Documentation** ([.web-docs](.web-docs)) - Complete usage documentation
- **Examples** ([example](example)) - Working configuration examples

## Quick Start

## Installation

### Using Packer Init (Recommended)

Add the following to your Packer configuration and run `packer init`:

```hcl
packer {
  required_plugins {
    vergeio = {
      source  = "github.com/verge-io/vergeio"
      version = ">=0.1.1"
    }
  }
}
```

### Using Packer Plugins Install

```bash
packer plugins install github.com/verge-io/vergeio
```

### 2. Configuration

Create a variables file with your VergeIO credentials:

```bash
# Copy example variables
cp example/example.pkrvars.hcl example/local.pkrvars.hcl

# Edit with your VergeIO cluster details
vim example/local.pkrvars.hcl
```

### 3. Build Your First VM

```bash
# Validate configuration
packer validate -var-file="example/local.pkrvars.hcl" example/build.pkr.hcl

# Build VM
packer build -var-file="example/local.pkrvars.hcl" example/build.pkr.hcl
```

## Configuration Example

```hcl
# Variables for VergeIO connection
variable "vergeio_endpoint" {
  type        = string
  description = "VergeIO cluster endpoint"
}

variable "vergeio_username" {
  type        = string
  description = "VergeIO cluster username"
}

variable "vergeio_password" {
  type        = string
  description = "VergeIO cluster password"
  sensitive   = true
}

variable "static_ip_address" {
  type        = string
  description = "Static IP address with CIDR notation"
  default     = "192.168.1.2/24"
}

variable "gateway_ip" {
  type        = string
  description = "Gateway IP address"
  default     = "192.168.1.1"
}

# Network data source - discover network by name
data "vergeio-networks" "external_network" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password

  filter_name = "External"  # Find network by name
}

# VergeIO VM source
source "vergeio" "example" {
  # VergeIO connection
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  vergeio_insecure = true
  vergeio_port     = 443

  # VM configuration
  name         = "packer-vm"
  description  = "VM built with Packer"
  os_family    = "linux"
  cpu_cores    = 4
  ram          = 4096
  power_state  = true
  guest_agent  = true

  # Storage with import support
  vm_disks {
    name           = "System Disk"
    disksize       = 20
    interface      = "virtio-scsi"
    preferred_tier = 1
    media          = "import"      # Automatic import waiting
    media_source   = 15            # Source disk/image ID
  }

  # Network with dynamic discovery
  vm_nics {
    name             = "primary_nic"
    vnet             = data.vergeio-networks.external_network.networks[0].id
    interface        = "virtio"
    assign_ipaddress = true
    enabled          = true
  }

  # Cloud-init configuration with inline contents
  cloud_init_data_source = "nocloud"
  cloud_init_files {
    name     = "user-data"
    contents = <<-EOF
      #cloud-config
      users:
        - name: packer
          sudo: ALL=(ALL) NOPASSWD:ALL
          shell: /bin/bash
          lock_passwd: false
          passwd: $6$rounds=4096$salt$hash...

      packages:
        - curl
        - wget
        - htop
      EOF
  }

  cloud_init_files {
    name     = "meta-data"
    contents = <<-EOF
      instance-id: packer-vm-${uuidv4()}
      local-hostname: packer-vm
      EOF
  }

  # SSH connectivity
  communicator = "ssh"
  ssh_username = "packer"
  ssh_password = "your-password"
  ssh_timeout  = "20m"

  # Power-on timeout configuration (optional)
  # power_on_timeout = "3m"  # How long to wait for VM to power on (default: 2m)
  # boot_timeout = "7m"      # Kept for compatibility (not used - SSH communicator handles boot waiting)

  # Graceful shutdown
  shutdown_command = "sudo shutdown -P now"
  shutdown_timeout = "5m"
}

# Build with provisioning
build {
  sources = ["source.vergeio.example"]

  # Verify SSH connectivity and system state
  provisioner "shell" {
    inline = [
      "echo 'SSH connection successful!'",
      "whoami",
      "hostname",
      "ip addr show",
      "echo 'VM provisioning completed!'"
    ]
  }
}
```

## Key Features Explained

### Disk Import Management

The plugin automatically handles disk imports with `media="import"`:

```hcl
vm_disks {
  media        = "import"
  media_source = 15  # Source disk/image ID
}
```

The plugin will:

1. Create the disk with import configuration
2. Wait for import to complete (prevents power-on errors)
3. Proceed with VM power-on only after import success

### Network Data Source

Discover networks dynamically instead of hardcoding IDs:

```hcl
data "vergeio-networks" "external" {
  filter_name = "External"  # Network name
}

vm_nics {
  vnet = data.vergeio-networks.external.networks[0].id
}
```

### Cloud-Init File Loading

The plugin supports both inline contents and external file loading for cloud-init configuration:

#### Inline Contents (Traditional)

```hcl
cloud_init_files {
  name = "user-data"
  contents = <<-EOF
    #cloud-config
    hostname: my-vm
    users:
      - name: admin
        sudo: ALL=(ALL) NOPASSWD:ALL
    EOF
}
```

#### External File Loading

```hcl
# Single file
cloud_init_files {
  name  = "user-data"
  files = ["cloud-init/user-data.yml"]
}

# Multiple files (concatenated)
cloud_init_files {
  name  = "user-data"
  files = [
    "cloud-init/base-config.yml",
    "cloud-init/packages.yml",
    "cloud-init/services.yml"
  ]
}
```

#### File Loading Rules

- **Mutually Exclusive**: Use either `contents` OR `files`, never both
- **Optional**: Both `contents` and `files` are optional - entries without either are skipped
- **Multiple Files**: Files are concatenated with newlines in the order specified
- **Path Resolution**: Relative paths are resolved from the Packer template directory
- **Validation**: Files must exist and be readable during configuration validation

#### Example File Structure

```
project/
├── build.pkr.hcl
├── cloud-init/
│   ├── user-data.yml      # Base user configuration
│   ├── meta-data.yml      # Instance metadata
│   ├── network-config.yml # Network configuration
│   ├── packages.yml       # Additional packages
│   └── services.yml       # Service configuration
└── variables.pkrvars.hcl
```

#### Comprehensive Example

```hcl
cloud_init_data_source = "nocloud"

# Load base configuration from file
cloud_init_files {
  name  = "user-data"
  files = ["cloud-init/user-data.yml"]
}

# Load metadata from file
cloud_init_files {
  name  = "meta-data"
  files = ["cloud-init/meta-data.yml"]
}

# Combine multiple configuration files
cloud_init_files {
  name  = "user-data-extended"
  files = [
    "cloud-init/base-config.yml",
    "cloud-init/packages.yml",
    "cloud-init/services.yml"
  ]
}

# Optional cloud-init file (skipped if no contents/files)
cloud_init_files {
  name = "vendor-data"
  # No contents or files - this entry will be skipped
}
```

### Variable Management

Centralized credential management:

```hcl
# Define once
variable "vergeio_password" {
  type      = string
  sensitive = true
}

# Use everywhere
data "vergeio-networks" "net" {
  vergeio_password = var.vergeio_password
}

source "vergeio" "vm" {
  vergeio_password = var.vergeio_password
}
```

## Build Phases

The plugin executes builds in these phases:

1. **VM Creation**: VM + Disks + NICs creation
2. **Disk Import Completion**: Wait for any import disks to finish
3. **Power Management**: VM power-on with status verification
4. **Network Discovery**: IP address discovery (guest agent)
5. **Provisioning**: SSH/WinRM connectivity and provisioning
6. **Cleanup**: Graceful shutdown and finalization

## Platform Support

**Tested Operating Systems:**

- Linux distributions (Ubuntu, CentOS, Debian)
- Windows Server (with WinRM)

**Communication Methods:**

- SSH (Linux)
- WinRM (Windows)
- Custom communicators

## Troubleshooting

**Enable Debug Logging:**

```bash
PACKER_LOG=1 packer build -var-file="local.pkrvars.hcl" build.pkr.hcl
```

**Common Issues:**

- **"No such file or directory" for vnet**: Network ID doesn't exist - use network data source
- **"Cannot power on VM while drives importing"**: Fixed automatically with import waiting
- **SSH timeout**: Check cloud-init configuration and network settings
- **Guest agent not reporting IP**: Ensure guest agent is installed on source image

## Development

### Building from Source

```bash
# Clone repository
git clone https://github.com/verge-io/packer-plugin-vergeio.git
cd packer-plugin-vergeio

# Build plugin
go build -ldflags="-X github.com/verge-io/packer-plugin-vergeio/version.VersionPrerelease=dev" -o packer-plugin-vergeio

# Install locally
packer plugins install --path packer-plugin-vergeio github.com/verge-io/vergeio
```

### Running Tests

```bash
# Install plugin locally first
make dev

# Run acceptance tests
PACKER_ACC=1 go test -count 1 -v ./... -timeout=120m
```

### Linux/macOS Development

```bash
# Quick development build and install
make dev
```

### Windows Development

```powershell
$MODULE_NAME = (Get-Content go.mod | Where-Object { $_ -match "^module"  }) -replace 'module ',''
$FQN = $MODULE_NAME -replace 'packer-plugin-',''
go build -ldflags="-X $MODULE_NAME/version.VersionPrerelease=dev" -o packer-plugin-vergeio.exe
packer plugins install --path packer-plugin-vergeio.exe $FQN
```

## Requirements

- **Go**: >= 1.20
- **Packer**: >= v1.10.2
- **packer-plugin-sdk**: >= v0.6.1
- **VergeIO**: API v4 compatible cluster

## License

This project is licensed under the Mozilla Public License 2.0 - see the [LICENSE](LICENSE) file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## Support

For issues and questions:

- GitHub Issues: [packer-plugin-vergeio/issues](https://github.com/verge-io/packer-plugin-vergeio/issues)
- VergeIO Documentation: [VergeIO Docs](https://www.verge.io/resources/documents/)
- Packer Documentation: [Packer Docs](https://www.packer.io/docs/)
