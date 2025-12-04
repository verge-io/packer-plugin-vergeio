# VergeIO Provisioner

The VergeIO provisioner provides VergeIO-specific provisioning capabilities during the Packer build process. This provisioner can perform VergeIO platform-specific operations and configurations that are not possible with standard Packer provisioners.

The provisioner integrates with the VergeIO API to perform actions such as VM configuration adjustments, network settings modifications, or other platform-specific customizations during the provisioning phase.

## Configuration Reference

**Required:**

- `vergeio_endpoint` (string) - The VergeIO cluster endpoint URL
- `vergeio_username` (string) - Username for VergeIO cluster authentication
- `vergeio_password` (string) - Password for VergeIO cluster authentication
- `action` (string) - The action to perform (`configure_vm`, `update_metadata`, `snapshot`)

**Optional:**

### Connection Configuration

- `vergeio_port` (int) - VergeIO cluster port. Defaults to `443`
- `vergeio_insecure` (bool) - Skip TLS certificate verification. Defaults to `false`

### Action-Specific Configuration

- `vm_settings` (map) - VM configuration settings to apply (for `configure_vm` action)
- `metadata` (map) - Metadata to set or update (for `update_metadata` action)
- `snapshot_name` (string) - Name for VM snapshot (for `snapshot` action)
- `snapshot_description` (string) - Description for VM snapshot (for `snapshot` action)

### Advanced Options

- `timeout` (string) - Maximum time to wait for the action to complete. Defaults to `5m`
- `retry_attempts` (int) - Number of retry attempts on failure. Defaults to `3`
- `retry_delay` (string) - Delay between retry attempts. Defaults to `10s`

## Example Usage

### Update VM Configuration

```hcl
build {
  sources = ["source.vergeio.example"]

  provisioner "shell" {
    inline = ["echo 'Installing software...'"]
  }

  # Update VM settings after provisioning
  provisioner "vergeio" {
    vergeio_endpoint = var.vergeio_endpoint
    vergeio_username = var.vergeio_username
    vergeio_password = var.vergeio_password

    action = "configure_vm"
    vm_settings = {
      description = "VM provisioned with Packer - ${formatdate("YYYY-MM-DD", timestamp())}"
      cpu_cores   = 4
      ram         = 4096
    }
  }
}
```

### Create VM Snapshot

```hcl
build {
  sources = ["source.vergeio.example"]

  provisioner "shell" {
    inline = [
      "sudo apt-get update",
      "sudo apt-get install -y docker.io"
    ]
  }

  # Create snapshot after software installation
  provisioner "vergeio" {
    vergeio_endpoint = var.vergeio_endpoint
    vergeio_username = var.vergeio_username
    vergeio_password = var.vergeio_password

    action           = "snapshot"
    snapshot_name    = "post-docker-install"
    snapshot_description = "VM snapshot after Docker installation"
    timeout          = "10m"
  }
}
```

### Update VM Metadata

```hcl
build {
  sources = ["source.vergeio.example"]

  provisioner "shell" {
    script = "scripts/install-application.sh"
  }

  # Update VM metadata with build information
  provisioner "vergeio" {
    vergeio_endpoint = var.vergeio_endpoint
    vergeio_username = var.vergeio_username
    vergeio_password = var.vergeio_password

    action = "update_metadata"
    metadata = {
      "build_date"    = formatdate("YYYY-MM-DD hh:mm:ss", timestamp())
      "build_user"    = env("USER")
      "packer_version" = packer.version
      "application_version" = "1.0.0"
    }

    retry_attempts = 5
    retry_delay    = "15s"
  }
}
```

### Multi-Step Provisioning with VergeIO Actions

```hcl
build {
  sources = ["source.vergeio.web-server"]

  # Install base software
  provisioner "shell" {
    inline = [
      "sudo apt-get update",
      "sudo apt-get install -y nginx postgresql"
    ]
  }

  # Create checkpoint snapshot
  provisioner "vergeio" {
    vergeio_endpoint = var.vergeio_endpoint
    vergeio_username = var.vergeio_username
    vergeio_password = var.vergeio_password

    action              = "snapshot"
    snapshot_name       = "base-software-installed"
    snapshot_description = "Checkpoint after base software installation"
  }

  # Configure applications
  provisioner "file" {
    source      = "configs/"
    destination = "/tmp/configs/"
  }

  provisioner "shell" {
    script = "scripts/configure-services.sh"
  }

  # Update VM configuration for production
  provisioner "vergeio" {
    vergeio_endpoint = var.vergeio_endpoint
    vergeio_username = var.vergeio_username
    vergeio_password = var.vergeio_password

    action = "configure_vm"
    vm_settings = {
      description = "Production Web Server - Built ${formatdate("YYYY-MM-DD", timestamp())}"
      # Optimize for web workload
      cpu_type = "host"
      # Add production metadata
      advanced = jsonencode({
        workload_type = "web-server"
        environment   = "production"
      })
    }
  }

  # Final snapshot
  provisioner "vergeio" {
    vergeio_endpoint = var.vergeio_endpoint
    vergeio_username = var.vergeio_username
    vergeio_password = var.vergeio_password

    action              = "snapshot"
    snapshot_name       = "production-ready"
    snapshot_description = "Production-ready web server snapshot"
    timeout             = "15m"
  }
}
```

## Available Actions

### `configure_vm`

Updates VM hardware and configuration settings.

**Supported `vm_settings`:**

- `cpu_cores` (int) - Number of CPU cores
- `ram` (int) - RAM in MB
- `cpu_type` (string) - CPU type/model
- `description` (string) - VM description
- `advanced` (string) - Advanced settings JSON

### `update_metadata`

Sets or updates VM metadata for tracking and organization.

**Supported `metadata` keys:**

- Any string key-value pairs for metadata tracking
- Common examples: `build_date`, `build_user`, `version`, `environment`

### `snapshot`

Creates a VM snapshot with specified name and description.

**Required parameters:**

- `snapshot_name` (string) - Snapshot name
- `snapshot_description` (string) - Snapshot description

## Features

- **API Integration**: Direct integration with VergeIO API for platform-specific operations
- **Retry Logic**: Configurable retry attempts with delays for reliability
- **Multiple Actions**: Support for VM configuration, metadata updates, and snapshots
- **Timeout Control**: Configurable timeouts for all operations
- **Error Handling**: Comprehensive error reporting and handling

## Notes

- The provisioner requires a VM to already exist (created by VergeIO builder)
- All actions are performed using the VergeIO API during the provisioning phase
- Snapshots created during provisioning can be used for rollback or cloning
- VM configuration changes take effect immediately and may affect running processes
- Metadata updates are preserved across VM reboots and operations
- The provisioner supports the same authentication methods as the VergeIO builder
