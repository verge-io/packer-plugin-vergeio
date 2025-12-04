# VergeIO Builder

The VergeIO builder creates virtual machines on VergeIO virtualization platform. This builder provides complete VM lifecycle management including hardware configuration, storage setup, network configuration, and cloud-init provisioning.

The builder supports both Linux and Windows VMs with advanced features like UEFI boot, secure boot, static IP configuration, and graceful shutdown handling.

## Configuration Reference

**Required:**

- `vergeio_endpoint` (string) - The VergeIO cluster endpoint URL (e.g., `https://your-cluster.example.com`)
- `vergeio_username` (string) - Username for VergeIO cluster authentication
- `vergeio_password` (string) - Password for VergeIO cluster authentication

**Optional:**

### Connection Configuration

- `vergeio_port` (int) - VergeIO cluster port. Defaults to `443`
- `vergeio_insecure` (bool) - Skip TLS certificate verification. Defaults to `false`

### VM Hardware Configuration

- `name` (string) - VM name. Required for VM identification
- `description` (string) - VM description
- `cpu_cores` (int) - Number of CPU cores to assign to the VM
- `ram` (int) - RAM in MB to assign to the VM
- `cpu_type` (string) - CPU type/model for the VM
- `machine_type` (string) - Machine type (e.g., `pc-q35-8.2`)
- `uefi` (bool) - Enable UEFI boot. Defaults to `false`
- `secure_boot` (bool) - Enable UEFI secure boot. Requires `uefi = true`
- `guest_agent` (bool) - Enable guest agent for IP discovery. Defaults to `false`

### Storage Configuration

- `vm_disks` (list) - List of disk configurations for the VM:
  - `name` (string) - Disk name
  - `disksize` (int) - Disk size in GB
  - `interface` (string) - Disk interface (e.g., `virtio`, `ide`)
  - `media` (string) - Media type (`disk`, `import`, `cdrom`)
  - `media_source` (int) - Source media ID for imports
  - `preferred_tier` (string) - Storage tier preference

### Network Configuration

- `vm_nics` (list) - List of network interface configurations:
  - `name` (string) - NIC name
  - `vnet` (int) - Virtual network ID to attach the NIC to
  - `driver` (string) - NIC driver (e.g., `virtio`)
  - `enabled` (bool) - Enable the NIC. Defaults to `true`

### Cloud-Init Configuration

- `cloud_init_files` (list) - Cloud-init configuration files:
  - `name` (string) - File name (e.g., `user-data`, `meta-data`, `network-config`)
  - `contents` (string) - Inline file contents (mutually exclusive with `files`)
  - `files` (list of strings) - External file paths to load and concatenate

### Power and Timeout Configuration

- `power_on_timeout` (string) - Maximum time to wait for VM to power on. Defaults to `2m`
- `boot_timeout` (string) - Legacy timeout setting (not used - kept for compatibility)
- `shutdown_command` (string) - Command to run for graceful VM shutdown (e.g., `sudo shutdown -P now`)
- `shutdown_timeout` (string) - Maximum time to wait for shutdown command. Defaults to `5m`

## Example Usage

### Basic Linux VM

```hcl
source "vergeio" "linux-vm" {
  # VergeIO connection
  vergeio_endpoint = "https://your-cluster.example.com"
  vergeio_username = "your-username"
  vergeio_password = "your-password"

  # VM configuration
  name        = "packer-linux-vm"
  description = "Linux VM built with Packer"
  cpu_cores   = 2
  ram         = 2048
  uefi        = true

  # Storage
  vm_disks {
    name      = "system-disk"
    disksize  = 20
    interface = "virtio"
    media     = "import"
    media_source = 123  # Your base image ID
  }

  # Network
  vm_nics {
    name   = "nic1"
    vnet   = 1
    driver = "virtio"
  }

  # Cloud-init for user configuration
  cloud_init_files {
    name     = "user-data"
    contents = <<-EOF
      #cloud-config
      users:
        - name: packer
          sudo: ALL=(ALL) NOPASSWD:ALL
          ssh_authorized_keys:
            - ssh-rsa YOUR_SSH_KEY
    EOF
  }

  cloud_init_files {
    name  = "network-config"
    files = ["cloud-init/network-config.yml"]
  }

  # SSH configuration
  communicator = "ssh"
  ssh_username = "packer"
  ssh_password = "your-password"

  # Graceful shutdown
  shutdown_command = "sudo shutdown -P now"
  shutdown_timeout = "5m"
}

build {
  sources = ["source.vergeio.linux-vm"]

  provisioner "shell" {
    inline = [
      "sudo apt-get update",
      "sudo apt-get install -y nginx",
      "sudo systemctl enable nginx"
    ]
  }
}
```

### Windows VM with External Cloud-Init Files

```hcl
source "vergeio" "windows-vm" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password

  name        = "packer-windows-vm"
  cpu_cores   = 4
  ram         = 4096
  uefi        = true
  secure_boot = true

  vm_disks {
    name         = "system-disk"
    disksize     = 50
    interface    = "virtio"
    media        = "import"
    media_source = 456  # Windows base image ID
  }

  vm_nics {
    name   = "nic1"
    vnet   = 1
    driver = "virtio"
  }

  # Load cloud-init from external files
  cloud_init_files {
    name  = "user-data"
    files = ["cloud-init/windows-user-data.yml"]
  }

  cloud_init_files {
    name  = "meta-data"
    files = ["cloud-init/meta-data.yml"]
  }

  # WinRM configuration
  communicator   = "winrm"
  winrm_username = "Administrator"
  winrm_password = "YourPassword123!"

  shutdown_command = "shutdown /s /t 0"
}
```

### Advanced VM with Multiple File Cloud-Init

```hcl
source "vergeio" "advanced-vm" {
  vergeio_endpoint = "https://cluster.example.com"
  vergeio_username = "admin"
  vergeio_password = "password"

  name = "advanced-packer-vm"
  cpu_cores = 8
  ram = 8192
  machine_type = "pc-q35-8.2"
  uefi = true
  guest_agent = true

  vm_disks {
    name         = "system"
    disksize     = 50
    interface    = "virtio"
    media        = "import"
    media_source = 789
  }

  vm_disks {
    name      = "data"
    disksize  = 100
    interface = "virtio"
    media     = "disk"
  }

  vm_nics {
    name   = "management"
    vnet   = 10
    driver = "virtio"
  }

  # Modular cloud-init configuration
  cloud_init_files {
    name  = "user-data"
    files = [
      "cloud-init/base-config.yml",
      "cloud-init/packages.yml",
      "cloud-init/users.yml"
    ]
  }

  cloud_init_files {
    name     = "network-config"
    contents = templatefile("cloud-init/network.tpl", {
      static_ip = "192.168.1.100"
      gateway   = "192.168.1.1"
    })
  }

  communicator = "ssh"
  ssh_username = "packer"
  ssh_private_key_file = "~/.ssh/id_rsa"

  shutdown_command = "sudo shutdown -P now"
  power_on_timeout = "3m"
  shutdown_timeout = "10m"
}
```

## Features

- **Complete VM Lifecycle**: Creation, provisioning, and cleanup
- **Cloud-Init Integration**: Full support for user-data, meta-data, and network-config
- **External File Loading**: Load cloud-init from external files with concatenation support
- **Static IP Support**: Automatic IP extraction from cloud-init network configuration
- **Graceful Shutdown**: 4-phase shutdown process with power state verification
- **Multiple OS Support**: Linux and Windows VMs with appropriate defaults
- **Storage Management**: Disk imports, resize handling, and multiple disk support
- **Network Flexibility**: Multiple NICs with VLAN support
- **Timeout Configuration**: Configurable timeouts for all build phases
- **Error Recovery**: Comprehensive cleanup and rollback on failures

## Notes

- VMs are created in powered-off state and powered on during the build process
- Disk imports are automatically waited for before VM power-on
- Static IP addresses take priority over guest agent IP discovery
- The builder supports both SSH and WinRM communicators
- Cloud-init files support both inline contents and external file loading
- Graceful shutdown falls back to forced power-off if SSH/WinRM shutdown fails
