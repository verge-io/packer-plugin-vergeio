# VergeIO Post-processor

The VergeIO post-processor handles post-build operations for VMs created with the VergeIO builder. This post-processor can export VMs as templates, create snapshots, register images in catalogs, or perform cleanup operations after the Packer build completes.

The post-processor integrates with the VergeIO API to perform operations like template creation, VM cloning, snapshot management, and artifact registration.

## Configuration Reference

**Required:**

- `vergeio_endpoint` (string) - The VergeIO cluster endpoint URL
- `vergeio_username` (string) - Username for VergeIO cluster authentication
- `vergeio_password` (string) - Password for VergeIO cluster authentication
- `action` (string) - The action to perform (`template`, `snapshot`, `clone`, `export`, `cleanup`)

**Optional:**

### Connection Configuration

- `vergeio_port` (int) - VergeIO cluster port. Defaults to `443`
- `vergeio_insecure` (bool) - Skip TLS certificate verification. Defaults to `false`

### Template Creation Options

- `template_name` (string) - Name for the created template (for `template` action)
- `template_description` (string) - Description for the template (for `template` action)
- `template_category` (string) - Template category for organization (for `template` action)

### Snapshot Options

- `snapshot_name` (string) - Name for VM snapshot (for `snapshot` action)
- `snapshot_description` (string) - Description for snapshot (for `snapshot` action)

### Clone Options

- `clone_name` (string) - Name for cloned VM (for `clone` action)
- `clone_count` (int) - Number of clones to create (for `clone` action). Defaults to `1`

### Export Options

- `export_path` (string) - Local path to export VM image (for `export` action)
- `export_format` (string) - Export format (`ova`, `vmx`, `raw`) (for `export` action)

### Advanced Options

- `keep_source_vm` (bool) - Keep the source VM after post-processing. Defaults to `false`
- `timeout` (string) - Maximum time to wait for operations. Defaults to `30m`
- `tags` (map) - Tags to apply to created resources
- `metadata` (map) - Additional metadata for created resources

## Example Usage

### Create VM Template

```hcl
build {
  sources = ["source.vergeio.web-server"]

  provisioner "shell" {
    inline = [
      "sudo apt-get update",
      "sudo apt-get install -y nginx"
    ]
  }

  # Create a reusable template
  post-processor "vergeio" {
    vergeio_endpoint = var.vergeio_endpoint
    vergeio_username = var.vergeio_username
    vergeio_password = var.vergeio_password

    action              = "template"
    template_name       = "nginx-web-server"
    template_description = "Ubuntu 22.04 with Nginx pre-installed"
    template_category   = "web-servers"

    tags = {
      "created_by" = "packer"
      "os"         = "ubuntu-22.04"
      "software"   = "nginx"
    }

    keep_source_vm = false
  }
}
```

### Create Snapshot and Clone

```hcl
build {
  sources = ["source.vergeio.database-server"]

  provisioner "shell" {
    script = "scripts/install-postgresql.sh"
  }

  # Create snapshot for backup
  post-processor "vergeio" {
    vergeio_endpoint = var.vergeio_endpoint
    vergeio_username = var.vergeio_username
    vergeio_password = var.vergeio_password

    action               = "snapshot"
    snapshot_name        = "postgresql-base-snapshot"
    snapshot_description = "Database server with PostgreSQL installed"
    keep_source_vm       = true
  }

  # Create development clones
  post-processor "vergeio" {
    vergeio_endpoint = var.vergeio_endpoint
    vergeio_username = var.vergeio_username
    vergeio_password = var.vergeio_password

    action      = "clone"
    clone_name  = "dev-database"
    clone_count = 3

    metadata = {
      "environment" = "development"
      "purpose"     = "testing"
    }
  }
}
```

### Export VM for External Use

```hcl
build {
  sources = ["source.vergeio.application-server"]

  provisioner "file" {
    source      = "app/"
    destination = "/opt/app/"
  }

  provisioner "shell" {
    script = "scripts/configure-app.sh"
  }

  # Export for distribution
  post-processor "vergeio" {
    vergeio_endpoint = var.vergeio_endpoint
    vergeio_username = var.vergeio_username
    vergeio_password = var.vergeio_password

    action        = "export"
    export_path   = "exports/application-server.ova"
    export_format = "ova"
    timeout       = "45m"
  }
}
```

### Multi-Stage Post-Processing

```hcl
build {
  sources = ["source.vergeio.golden-image"]

  provisioner "shell" {
    scripts = [
      "scripts/security-hardening.sh",
      "scripts/monitoring-setup.sh",
      "scripts/cleanup.sh"
    ]
  }

  # Create production template
  post-processor "vergeio" {
    name = "template"

    vergeio_endpoint = var.vergeio_endpoint
    vergeio_username = var.vergeio_username
    vergeio_password = var.vergeio_password

    action              = "template"
    template_name       = "golden-image-${formatdate("YYYY-MM", timestamp())}"
    template_description = "Hardened golden image with monitoring - ${formatdate("YYYY-MM-DD", timestamp())}"
    template_category   = "golden-images"

    tags = {
      "version"     = "v1.0"
      "environment" = "production"
      "compliance"  = "hardened"
    }

    keep_source_vm = true
  }

  # Create backup snapshot
  post-processor "vergeio" {
    name = "backup"

    vergeio_endpoint = var.vergeio_endpoint
    vergeio_username = var.vergeio_username
    vergeio_password = var.vergeio_password

    action               = "snapshot"
    snapshot_name        = "golden-backup-${formatdate("YYYY-MM-DD", timestamp())}"
    snapshot_description = "Backup snapshot of golden image"

    metadata = {
      "backup_type" = "golden-image"
      "retention"   = "1-year"
    }
  }

  # Export for offline storage
  post-processor "vergeio" {
    name = "export"

    vergeio_endpoint = var.vergeio_endpoint
    vergeio_username = var.vergeio_username
    vergeio_password = var.vergeio_password

    action        = "export"
    export_path   = "archives/golden-image-${formatdate("YYYY-MM-DD", timestamp())}.ova"
    export_format = "ova"
    timeout       = "60m"
  }
}
```

### Cleanup Operations

```hcl
build {
  sources = ["source.vergeio.test-vm"]

  provisioner "shell" {
    inline = ["echo 'Testing completed'"]
  }

  # Clean up test resources
  post-processor "vergeio" {
    vergeio_endpoint = var.vergeio_endpoint
    vergeio_username = var.vergeio_username
    vergeio_password = var.vergeio_password

    action         = "cleanup"
    keep_source_vm = false

    # Remove any snapshots and temporary resources
    metadata = {
      "cleanup_snapshots" = "true"
      "cleanup_clones"    = "true"
    }
  }
}
```

## Available Actions

### `template`

Creates a VM template from the built VM.

**Required parameters:**

- `template_name` (string) - Template name
- `template_description` (string) - Template description

**Optional parameters:**

- `template_category` (string) - Template category

### `snapshot`

Creates a VM snapshot for backup or cloning purposes.

**Required parameters:**

- `snapshot_name` (string) - Snapshot name
- `snapshot_description` (string) - Snapshot description

### `clone`

Creates one or more VM clones from the built VM.

**Required parameters:**

- `clone_name` (string) - Base name for clones
- `clone_count` (int) - Number of clones to create

### `export`

Exports the VM to a local file in specified format.

**Required parameters:**

- `export_path` (string) - Local export path
- `export_format` (string) - Export format (`ova`, `vmx`, `raw`)

### `cleanup`

Performs cleanup operations on the VM and related resources.

**Optional parameters:**

- `keep_source_vm` (bool) - Whether to keep the source VM

## Features

- **Template Creation**: Convert VMs to reusable templates
- **Snapshot Management**: Create snapshots for backup and cloning
- **VM Cloning**: Create multiple VM instances from a single build
- **Export Capabilities**: Export VMs in standard formats (OVA, VMX, RAW)
- **Resource Cleanup**: Automated cleanup of build artifacts
- **Metadata Support**: Add tags and metadata to created resources
- **Multi-Stage Processing**: Chain multiple post-processing actions
- **Timeout Control**: Configurable timeouts for long-running operations

## Notes

- Post-processors run after the build is complete and provisioning is finished
- Multiple post-processors can be chained using the `name` parameter
- The `keep_source_vm` option controls whether the original VM is preserved
- Templates created with the post-processor can be used as base images for future builds
- Exported VMs can be imported into other virtualization platforms
- All operations use the VergeIO API and require appropriate permissions
- Large VM exports may take considerable time depending on disk size and network speed
