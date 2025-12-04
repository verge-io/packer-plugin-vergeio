# VergeIO Packer Plugin Example
# This example demonstrates the key features of the VergeIO Packer plugin

# Define VergeIO connection variables
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

variable "vergeio_insecure" {
  type        = bool
  description = "Skip TLS certificate verification"
}

variable "vergeio_port" {
  type        = number
  description = "VergeIO cluster port"
}

# Network configuration variables
variable "static_ip_address" {
  type        = string
  description = "Static IP address with CIDR notation"
  default     = "192.168.1.100/24"
}

variable "gateway_ip" {
  type        = string
  description = "Gateway IP address"
  default     = "192.168.1.1"
}

variable "dns_servers" {
  type        = list(string)
  description = "List of DNS server addresses"
  default     = ["8.8.8.8", "1.1.1.1"]
}

# Required plugins block
packer {
  required_plugins {
    vergeio = {
      version = ">=v0.1.0"
      source  = "github.com/verge-io/vergeio"
    }
  }
}

# Network data source - discover network by name
data "vergeio-networks" "external_network" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  vergeio_insecure = var.vergeio_insecure
  vergeio_port     = var.vergeio_port

  filter_name = "farooq-test"
}

# VM data source - discover VM by name (example usage)
data "vergeio-vms" "template_vm" {
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  vergeio_insecure = var.vergeio_insecure
  vergeio_port     = var.vergeio_port

  filter_name = "farooqtemp"  # Change this to your template VM name
}

# VergeIO VM source configuration
source "vergeio" "example" {
  # VergeIO cluster connection
  vergeio_endpoint = var.vergeio_endpoint
  vergeio_username = var.vergeio_username
  vergeio_password = var.vergeio_password
  vergeio_insecure = var.vergeio_insecure
  vergeio_port     = var.vergeio_port

  # VM hardware configuration
  name         = "packer-example-vm"
  description  = "Example VM built with Packer VergeIO plugin"
  enabled      = true
  os_family    = "linux"
  cpu_cores    = 4
  machine_type = "q35"
  ram          = 4096
  power_state  = true
  guest_agent  = true
  console      = "serial"

  # VM storage configuration
  vm_disks {
    # name           = "System Disk"
    # description    = "Primary system disk"
    # disksize       = 30
    # interface      = "virtio-scsi"
    # preferred_tier = 1
    # orderid        = 0
    # media          = "import"
    # media_source   = 15  # Change this to your source disk/image ID
    name           = "os_drive"
    description    = "Clone of disk 12"
    media          = "clone"
    enabled        = true
    media_source   = data.vergeio-vms.template_vm.vms[0].drives[0].media_source.key
    preferred_tier = 3
  }

  # VM network configuration using data source
  vm_nics {
    name             = "primary_nic"
    vnet             = data.vergeio-networks.external_network.networks[0].id
    description      = "Primary network interface"
    interface        = "virtio"
    assign_ipaddress = true
    enabled          = true
  }

  # Cloud-init configuration
  cloud_init_data_source = "nocloud"

  # Load user configuration from external file
  cloud_init_files {
    name  = "user-data"
    files = ["cloud-init/user-data.yml"]
  }

  # Instance metadata
  cloud_init_files {
    name     = "meta-data"
    contents = <<-EOF
      instance-id: packer-example-vm
      local-hostname: packer-example-vm
    EOF
  }

  # Network configuration with variables
  cloud_init_files {
    name     = "network-config"
    contents = <<-EOF
      version: 2
      ethernets:  
        primary:
          match:
            name: "en*"
          addresses:
            - ${var.static_ip_address}
          gateway4: ${var.gateway_ip}
          nameservers:
            addresses: ${jsonencode(var.dns_servers)}
          renderer: networkd
    EOF
  }

  # SSH communicator configuration
  communicator = "ssh"
  ssh_username = "packer"
  ssh_password = "your-vm-password"
  ssh_timeout  = "20m"

  # Power-on timeout configuration (optional)
  power_on_timeout = "3m"
  boot_timeout     = "5m"

  # Graceful shutdown configuration
  shutdown_command = "sudo shutdown -P now"
  shutdown_timeout = "5m"
}

# Build configuration
build {
  sources = ["source.vergeio.example"]

  # Basic connectivity and system verification
  provisioner "shell" {
    inline = [
      "echo '=== Packer SSH Connection Successful ==='",
      "echo 'Connected as user:' $(whoami)",
      "echo 'Hostname:' $(hostname)",
      "echo 'OS Info:' $(cat /etc/os-release | head -2)",
      "echo 'Network:' $(ip route get 8.8.8.8 | head -1)",
      "echo '=== VM Provisioning Complete ==='",
    ]
  }

  # Example: Install additional packages
  provisioner "shell" {
    inline = [
      "echo 'Installing additional packages...'",
      "sudo apt-get update -y || sudo yum update -y",
      "sudo apt-get install -y htop curl wget || sudo yum install -y htop curl wget",
      "echo 'Package installation complete'"
    ]
  }

  # Generate build manifest (optional)
  post-processor "manifest" {
    output = "manifest.json"
  }
}