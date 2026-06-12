# Packer Plugin for VergeIO

A comprehensive Packer plugin for creating and managing virtual machines on VergeIO virtualization platform. This plugin provides complete VM lifecycle management with advanced features for enterprise environments.

The VergeIO plugin enables automated VM provisioning with support for cloud-init configuration, network discovery, storage management, and graceful shutdown handling. It's designed for production use with robust error handling and configurable timeouts.

## Features

**Complete VM Management:**

- VM creation with hardware, storage, and network configuration
- Intelligent power management with real power state verification
- Disk import support with automatic waiting
- Cloud-init integration with external file loading
- Graceful shutdown with 4-phase process

**Advanced Capabilities:**

- Network discovery by name for portable configurations
- VM discovery for template and configuration management
- Static IP configuration with automatic extraction
- Multiple storage tiers and interface support
- SSH/WinRM connectivity with optimized boot process

**Enterprise Ready:**

- Comprehensive error handling and rollback
- Configurable timeouts for all operations
- Production tested with detailed logging
- Multi-platform support (Linux, Windows)
- Full VergeIO API v4 integration

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

### Building from Source (Recommended only for development/troubleshooting)

```bash
git clone https://github.com/verge-io/packer-plugin-vergeio
cd packer-plugin-vergeio
make dev
```

## Quick Start

```hcl
packer {
  required_plugins {
    vergeio = {
      source  = "github.com/verge-io/vergeio"
      version = ">=0.1.0"
    }
  }
}

source "vergeio" "example" {
  # VergeIO connection
  vergeio_endpoint = "https://your-cluster.example.com"
  vergeio_username = "your-username"
  vergeio_password = "your-password"

  # VM configuration
  name        = "packer-example-vm"
  cpu_cores   = 2
  ram         = 2048
  uefi        = true

  # Storage
  vm_disks {
    name         = "system-disk"
    disksize     = 20
    interface    = "virtio"
    media        = "import"
    media_source = 123  # Your base image ID
  }

  # Network
  vm_nics {
    name   = "nic1"
    vnet   = 1
    driver = "virtio"
  }

  # Cloud-init
  cloud_init_files {
    name     = "user-data"
    contents = "#cloud-config\nusers:\n  - name: packer\n    sudo: ALL=(ALL) NOPASSWD:ALL"
  }

  # SSH connectivity
  communicator = "ssh"
  ssh_username = "packer"
  ssh_password = "your-password"

  shutdown_command = "sudo shutdown -P now"
}

build {
  sources = ["source.vergeio.example"]

  provisioner "shell" {
    inline = [
      "sudo apt-get update",
      "sudo apt-get install -y nginx"
    ]
  }
}
```

## Components

### Builder

The VergeIO builder creates virtual machines on VergeIO infrastructure with complete lifecycle management.

**Key Features:**

- Hardware configuration (CPU, RAM, machine type)
- Storage management with multiple disks and import support
- Network configuration with multiple NICs
- Cloud-init integration with external file loading
- Power management with configurable timeouts
- Graceful shutdown with power state verification

[→ Builder Documentation](/docs/builders/vergeio)

### Data Sources

#### Networks Data Source

Discover VergeIO networks by name or type for portable configurations.

```hcl
data "vergeio" "app_network" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password

  filter_name = "Application-Network"
}
```

[→ Networks Data Source Documentation](/docs/datasources/vergeio-networks)

#### VMs Data Source

Discover existing VMs for template discovery and configuration cloning.

```hcl
data "vergeio" "base_templates" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password

  is_snapshot = false
}
```

[→ VMs Data Source Documentation](/docs/datasources/vergeio-vms)

### Provisioner

Perform VergeIO-specific operations during provisioning such as VM configuration updates, metadata management, and snapshot creation.

```hcl
provisioner "vergeio" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password

  action           = "snapshot"
  snapshot_name    = "post-install-snapshot"
  snapshot_description = "Snapshot after software installation"
}
```

[→ Provisioner Documentation](/docs/provisioners/vergeio)

### Post-processor

Handle post-build operations like template creation, VM export, and resource cleanup.

```hcl
post-processor "vergeio" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password

  action              = "template"
  template_name       = "nginx-web-server"
  template_description = "Ubuntu with Nginx pre-installed"
}
```

[→ Post-processor Documentation](/docs/post-processors/vergeio)

## Configuration Examples

### Basic Linux VM

```hcl
source "vergeio" "ubuntu" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password

  name = "ubuntu-server"
  cpu_cores = 2
  ram = 2048
  uefi = true

  vm_disks {
    name         = "system"
    disksize     = 20
    interface    = "virtio"
    media        = "import"
    media_source = 123
  }

  vm_nics {
    name   = "eth0"
    vnet   = 1
    driver = "virtio"
  }

  communicator = "ssh"
  ssh_username = "ubuntu"
  ssh_password = "ubuntu"

  shutdown_command = "sudo shutdown -P now"
}
```

### Windows VM with Cloud-Init

```hcl
source "vergeio" "windows" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password

  name = "windows-server"
  cpu_cores = 4
  ram = 4096
  uefi = true
  secure_boot = true

  vm_disks {
    name         = "system"
    disksize     = 60
    interface    = "virtio"
    media        = "import"
    media_source = 456
  }

  cloud_init_files {
    name  = "user-data"
    files = ["cloud-init/windows-user-data.yml"]
  }

  communicator   = "winrm"
  winrm_username = "Administrator"
  winrm_password = "Password123!"

  shutdown_command = "shutdown /s /t 0"
}
```

### Multi-Network VM with External Files

```hcl
data "vergeio" "web_network" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password

  filter_name = "Web-DMZ"
}

source "vergeio" "web-server" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password

  name = "multi-nic-server"
  cpu_cores = 4
  ram = 8192
  uefi = true

  vm_disks {
    name      = "system"
    disksize  = 50
    interface = "virtio"
    media     = "disk"
  }

  vm_nics {
    name   = "dmz_nic"
    vnet   = data.vergeio.web_network.networks[0].id
    driver = "virtio"
  }

  vm_nics {
    name   = "internal_nic"
    vnet   = 10
    driver = "virtio"
  }

  cloud_init_files {
    name  = "user-data"
    files = [
      "cloud-init/base.yml",
      "cloud-init/users.yml",
      "cloud-init/packages.yml"
    ]
  }

  cloud_init_files {
    name = "network-config"
    files = ["cloud-init/dual-nic-network.yml"]
  }

  communicator = "ssh"
  ssh_username = "packer"
  ssh_private_key_file = "~/.ssh/id_rsa"

  shutdown_command = "sudo shutdown -P now"
  power_on_timeout = "3m"
  shutdown_timeout = "10m"
}
```

## Requirements

- **VergeIO Cluster**: Access to VergeIO virtualization platform
- **Packer**: Version 1.10.2 or later
- **Go**: Version 1.23.2 or later (for building from source)
- **Network Access**: Connectivity to VergeIO cluster API (default port 443)
- **Credentials**: VergeIO username and password with appropriate permissions

## Supported Platforms

- **Host Platforms**: Linux, macOS, Windows
- **Guest Operating Systems**: Linux distributions, Windows Server/Desktop
- **Architectures**: x86_64, ARM64 

## Documentation

- [Builder Configuration Reference](/docs/builders/vergeio)
- [Networks Data Source](/docs/datasources/vergeio-networks)
- [VMs Data Source](/docs/datasources/vergeio-vms)
- [Provisioner Reference](/docs/provisioners/vergeio)
- [Post-processor Reference](/docs/post-processors/vergeio)

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Support

- **Issues**: [GitHub Issues](https://github.com/verge-io/packer-plugin-vergeio/issues)
- **Documentation**: [Official Documentation](https://developer.hashicorp.com/packer/integrations/vergeio/vergeio/latest)
- **Community**: [HashiCorp Community Forum](https://discuss.hashicorp.com)

## License

This project is licensed under the Mozilla Public License 2.0 - see the [LICENSE](LICENSE) file for details.
