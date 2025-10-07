package vergeio

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
)

// Config represents the complete configuration for the VergeIO builder
// It combines Packer's standard configuration, VergeIO-specific settings,
// and provisioning-related configuration into a single structure
type Config struct {
	// PackerConfig contains standard Packer configuration fields like build name,
	// debug flags, and other core Packer functionality
	common.PackerConfig `mapstructure:",squash"`

	// Comm contains communicator configuration for SSH/WinRM connectivity
	// This enables Packer to connect to the VM for provisioning after it's created
	// Includes fields like ssh_username, ssh_password, ssh_port, ssh_timeout, etc.
	Comm communicator.Config `mapstructure:",squash"`

	// ClusterConfig contains VergeIO cluster connection information
	// This is embedded with squash so fields appear at the root level in HCL
	ClusterConfig `mapstructure:",squash"`

	// VmConfig contains VM specification including hardware, network, and storage
	// This is also embedded so VM fields appear at the root level in HCL
	VmConfig `mapstructure:",squash"`

	// ShutdownCommand is the command to run inside the VM to shut it down gracefully
	// Example: "sudo shutdown -P now" for Linux, "shutdown /s /t 0" for Windows
	ShutdownCommand string `mapstructure:"shutdown_command"`

	// ShutdownTimeout is how long to wait for the shutdown command to complete
	// If this timeout is exceeded, the VM will be forcefully powered off
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`

	// PowerOnTimeout is the maximum time to wait for the VM to power on
	// This should be relatively short as power-on is usually quick
	// Default: 2 minutes
	PowerOnTimeout time.Duration `mapstructure:"power_on_timeout"`

	// BootTimeout is how long to wait for the VM to fully boot after power-on
	// This needs to be longer to allow for OS boot sequence
	// Default: 5 minutes
	BootTimeout time.Duration `mapstructure:"boot_timeout"`
}

type Builder struct {
	config Config
	runner multistep.Runner
}

type ClusterConfig struct {
	Username string `mapstructure:"vergeio_username" required:"false"`
	Password string `mapstructure:"vergeio_password" required:"false"`
	Insecure bool   `mapstructure:"vergeio_insecure" required:"false"`
	Endpoint string `mapstructure:"vergeio_endpoint" required:"false"`
	Port     int    `mapstructure:"vergeio_port" required:"false"`
}

type VmConfig struct {
	// Id int `mapstructure:"id" required:"false" json:"id"`
	Machine              int             `mapstructure:"machine" required:"false" json:"machine"`
	Name                 string          `mapstructure:"name" required:"false" json:"name"`
	Cluster              string          `mapstructure:"cluster" required:"false"`
	Description          string          `mapstructure:"description" required:"false"`
	Enabled              bool            `mapstructure:"enabled" required:"false"`
	MachineType          string          `mapstructure:"machine_type" required:"false"`
	AllowHotplug         bool            `mapstructure:"allow_hotplug" required:"false"`
	DisablePowercycle    bool            `mapstructure:"disable_powercycle" required:"false"`
	CPUCores             int             `mapstructure:"cpu_cores" required:"false"`
	CPUType              string          `mapstructure:"cpu_type" required:"false"`
	RAM                  int             `mapstructure:"ram" required:"false"`
	Console              string          `mapstructure:"console" required:"false"`
	Display              string          `mapstructure:"display" required:"false"`
	Video                string          `mapstructure:"video" required:"false"`
	Sound                string          `mapstructure:"sound" required:"false"`
	OSFamily             string          `mapstructure:"os_family" required:"false"`
	OSDescription        string          `mapstructure:"os_description" required:"false"`
	RTCBase              string          `mapstructure:"rtc_base" required:"false"`
	BootOrder            string          `mapstructure:"boot_order" required:"false"`
	ConsolePassEnabled   bool            `mapstructure:"console_pass_enabled" required:"false"`
	ConsolePass          string          `mapstructure:"console_pass" required:"false"`
	USBTablet            bool            `mapstructure:"usb_tablet" required:"false"`
	UEFI                 bool            `mapstructure:"uefi" required:"false"`
	SecureBoot           bool            `mapstructure:"secure_boot" required:"false"`
	SerialPort           bool            `mapstructure:"serial_port" required:"false"`
	BootDelay            int             `mapstructure:"boot_delay" required:"false"`
	PreferredNode        string          `mapstructure:"preferred_node" required:"false"`
	SnapshotProfile      string          `mapstructure:"snapshot_profile" required:"false"`
	CloudInitDataSource  string          `mapstructure:"cloud_init_data_source" required:"false"`
	PowerState           bool            `mapstructure:"power_state" required:"false"`
	GuestAgent           bool            `mapstructure:"guest_agent" required:"false"`
	HAGroup              string          `mapstructure:"ha_group" required:"false"`
	Advanced             string          `mapstructure:"advanced" required:"false"`
	NestedVirtualization bool            `mapstructure:"nested_virtualization" required:"false"`
	DisableHypervisor    bool            `mapstructure:"disable_hypervisor" required:"false"`
	VmDiskConfigs        []VmDiskConfig  `mapstructure:"vm_disks" required:"false"`
	VmNicConfigs         []VmNicConfig   `mapstructure:"vm_nics" required:"false"`
	CloudInitFiles       []CloudInitFile `mapstructure:"cloud_init_files" required:"false"`
}

// CloudInitFile represents a cloud-init file with name and contents
// This mirrors the structure used in the Terraform VergeIO provider
// Supports both inline contents and external file paths (mutually exclusive)
type CloudInitFile struct {
	Name     string   `mapstructure:"name" required:"false"`
	Contents string   `mapstructure:"contents" required:"false"`
	Files    []string `mapstructure:"files" required:"false"`
}

type VmDiskConfig struct {
	Machine             int    `mapstructure:"machine" required:"false"`
	Name                string `mapstructure:"name" required:"false"`
	Description         string `mapstructure:"description" required:"false"`
	Interface           string `mapstructure:"interface" required:"false"`
	Media               string `mapstructure:"media" required:"false"`
	MediaSource         int    `mapstructure:"media_source" required:"false"`
	PreferredTier       string `mapstructure:"preferred_tier" required:"false"`
	DiskSize            int64  `mapstructure:"disksize" required:"false"`
	Enabled             bool   `mapstructure:"enabled" required:"false"`
	ReadOnly            bool   `mapstructure:"readonly" required:"false"`
	Serial              string `mapstructure:"serial" required:"false"`
	Asset               string `mapstructure:"asset" required:"false"`
	OrderId             int    `mapstructure:"orderid" required:"false"`
	PreserveDriveFormat bool   `mapstructure:"preserve_drive_format" required:"false"`
}

type VmNicConfig struct {
	Machine         int    `mapstructure:"machine" required:"false"`
	Name            string `mapstructure:"name" required:"false"`
	Description     string `mapstructure:"description" required:"false"`
	Interface       string `mapstructure:"interface" required:"false"`
	Driver          string `mapstructure:"driver" required:"false"`
	Model           string `mapstructure:"model" required:"false"`
	VNET            int    `mapstructure:"vnet" required:"false"`
	MAC             string `mapstructure:"macaddress" required:"false"`
	IPAddress       string `mapstructure:"ipaddress" required:"false"`
	AssignIPAddress bool   `mapstructure:"assign_ipaddress" required:"false"`
	Enabled         bool   `mapstructure:"enabled" required:"false"`
}

// Prepare validates and sets up the configuration for the VergeIO builder
// This method is called by Packer to validate the user's configuration and set defaults
func (b *Builder) Prepare(raws ...interface{}) (generatedVars []string, warnings []string, err error) {

	log.Printf("[Vergeio]: Starting Builder configuration preparation")
	log.Printf("[Vergeio]: Raw configuration input: %+v", raws)

	// Decode the user's HCL configuration into our Config struct
	// This converts the HCL input into Go struct fields
	err = config.Decode(&b.config, &config.DecodeOpts{
		PluginType:  "packer.builder.vergeio",
		Interpolate: true, // Allow variable interpolation in config
	}, raws...)

	if err != nil {
		log.Printf("[Vergeio]: Configuration decode failed: %+v", err)
		return nil, nil, err
	}

	log.Printf("[Vergeio]: Decoded configuration: %+v", b.config)
	log.Printf("[Vergeio]: VM configuration: %+v", b.config.VmConfig)
	log.Printf("[Vergeio]: Communicator configuration: %+v", b.config.Comm)

	// Accumulate any configuration errors and warnings
	var errs *packer.MultiError
	warnings = make([]string, 0)

	// === VergeIO Cluster Configuration Validation ===
	// These are required for connecting to the VergeIO API
	if b.config.Endpoint == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("vergeio_endpoint must be specified"))
	}
	if b.config.Username == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("vergeio_username must be specified"))
	}
	if b.config.Password == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("vergeio_password must be specified"))
	}

	// Set default port if not specified (HTTPS standard port)
	if b.config.Port == 0 {
		log.Printf("[Vergeio]: No port specified, defaulting to 443 (HTTPS)")
		b.config.Port = 443
	}

	// === Communicator Configuration Setup ===
	// The communicator is how Packer connects to the VM for provisioning

	// Set default communicator type to SSH if not specified
	if b.config.Comm.Type == "" {
		log.Printf("[Vergeio]: No communicator type specified, defaulting to SSH")
		b.config.Comm.Type = "ssh"
	}

	// Set default SSH port if using SSH communicator
	if b.config.Comm.Type == "ssh" && b.config.Comm.SSHPort == 0 {
		log.Printf("[Vergeio]: No SSH port specified, defaulting to 22")
		b.config.Comm.SSHPort = 22
	}

	// Set default SSH timeout if not specified
	if b.config.Comm.SSHTimeout == 0 {
		log.Printf("[Vergeio]: No SSH timeout specified, defaulting to 20 minutes")
		b.config.Comm.SSHTimeout = 20 * time.Minute
	}

	// Set default WinRM port if using WinRM communicator
	if b.config.Comm.Type == "winrm" && b.config.Comm.WinRMPort == 0 {
		if b.config.Comm.WinRMUseSSL {
			log.Printf("[Vergeio]: No WinRM port specified (SSL), defaulting to 5986")
			b.config.Comm.WinRMPort = 5986
		} else {
			log.Printf("[Vergeio]: No WinRM port specified (no SSL), defaulting to 5985")
			b.config.Comm.WinRMPort = 5985
		}
	}

	// Set default WinRM timeout if using WinRM
	if b.config.Comm.Type == "winrm" && b.config.Comm.WinRMTimeout == 0 {
		log.Printf("[Vergeio]: No WinRM timeout specified, defaulting to 20 minutes")
		b.config.Comm.WinRMTimeout = 20 * time.Minute
	}

	// === Shutdown Configuration Setup ===
	// Set default shutdown timeout if not specified
	if b.config.ShutdownTimeout == 0 {
		log.Printf("[Vergeio]: No shutdown timeout specified, defaulting to 5 minutes")
		b.config.ShutdownTimeout = 5 * time.Minute
	}

	// Validate that shutdown command is provided if we expect to run provisioners
	// (We'll add this validation later once we know the expected usage patterns)

	// === Network Configuration Validation ===
	// Ensure at least one NIC is configured for provisioning connectivity
	if len(b.config.VmNicConfigs) == 0 {
		warnings = append(warnings, "No vm_nics configured - provisioning may fail without network connectivity")
	}

	// === Communicator Validation ===
	// Validate that required communicator credentials are provided
	if b.config.Comm.Type == "ssh" {
		if b.config.Comm.SSHUsername == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("ssh_username is required when using SSH communicator"))
		}

		// Must have either password or private key for SSH authentication
		if b.config.Comm.SSHPassword == "" && b.config.Comm.SSHPrivateKeyFile == "" && !b.config.Comm.SSHAgentAuth {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("ssh_password, ssh_private_key_file, or ssh_agent_auth must be specified for SSH authentication"))
		}
	}

	if b.config.Comm.Type == "winrm" {
		if b.config.Comm.WinRMUser == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("winrm_username is required when using WinRM communicator"))
		}
		if b.config.Comm.WinRMPassword == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("winrm_password is required when using WinRM communicator"))
		}
	}

	// Check for any validation errors before continuing
	if errs != nil {
		log.Printf("[Vergeio]: Configuration validation failed with errors: %+v", errs)
		return nil, warnings, errs
	}

	// === Cloud-Init File Processing ===
	// Load contents from external files if specified
	err = b.processCloudInitFiles()
	if err != nil {
		log.Printf("[Vergeio]: Cloud-init file processing failed: %+v", err)
		errs = packer.MultiErrorAppend(errs, err)
		return nil, warnings, errs
	}

	log.Printf("[Vergeio]: Configuration validation completed successfully")
	log.Printf("[Vergeio]: Final configuration - Comm: %+v", b.config.Comm)
	log.Printf("[Vergeio]: Final configuration - Shutdown timeout: %v", b.config.ShutdownTimeout)

	// Return empty generated data for now
	// This could be extended to provide VM information to provisioners
	buildGeneratedData := []string{}

	return buildGeneratedData, warnings, nil
}

// processCloudInitFiles handles loading external cloud-init files and validates configuration
func (b *Builder) processCloudInitFiles() error {
	for i := range b.config.VmConfig.CloudInitFiles {
		cloudInitFile := &b.config.VmConfig.CloudInitFiles[i]

		// Check what's specified
		hasContents := cloudInitFile.Contents != ""
		hasFiles := len(cloudInitFile.Files) > 0

		// Validate that contents and files are mutually exclusive
		if hasContents && hasFiles {
			return fmt.Errorf("cloud_init_files[%d] (%s): 'contents' and 'files' are mutually exclusive", i, cloudInitFile.Name)
		}

		// If neither is specified, skip this entry (it's optional)
		if !hasContents && !hasFiles {
			log.Printf("[Vergeio]: Skipping cloud-init file '%s' (no contents or files specified)", cloudInitFile.Name)
			continue
		}

		// If files are specified, load and concatenate their contents
		if hasFiles {
			log.Printf("[Vergeio]: Loading cloud-init file '%s' from %d files: %v", cloudInitFile.Name, len(cloudInitFile.Files), cloudInitFile.Files)

			var allContents []string

			for j, filePath := range cloudInitFile.Files {
				// Handle relative paths by making them relative to the current working directory
				absolutePath := filePath
				if !filepath.IsAbs(filePath) {
					// Get current working directory
					wd, err := os.Getwd()
					if err != nil {
						return fmt.Errorf("failed to get working directory for cloud_init_files[%d] (%s) file[%d]: %w", i, cloudInitFile.Name, j, err)
					}
					absolutePath = filepath.Join(wd, filePath)
				}

				// Check if file exists
				if _, err := os.Stat(absolutePath); os.IsNotExist(err) {
					return fmt.Errorf("cloud_init_files[%d] (%s) file[%d]: file not found: %s", i, cloudInitFile.Name, j, absolutePath)
				}

				// Read file contents
				contents, err := os.ReadFile(absolutePath)
				if err != nil {
					return fmt.Errorf("cloud_init_files[%d] (%s) file[%d]: failed to read file %s: %w", i, cloudInitFile.Name, j, absolutePath, err)
				}

				allContents = append(allContents, string(contents))
				log.Printf("[Vergeio]: Loaded file %s (%d bytes)", filePath, len(contents))
			}

			// Concatenate all file contents with newlines
			cloudInitFile.Contents = ""
			for k, content := range allContents {
				if k > 0 {
					cloudInitFile.Contents += "\n"
				}
				cloudInitFile.Contents += content
			}

			// Clear the files array after loading (for security)
			cloudInitFile.Files = nil

			log.Printf("[Vergeio]: Successfully loaded cloud-init file '%s' (%d total bytes from %d files)", cloudInitFile.Name, len(cloudInitFile.Contents), len(allContents))
		}

		// Validate that contents is not empty after loading
		if cloudInitFile.Contents == "" {
			return fmt.Errorf("cloud_init_files[%d] (%s): contents cannot be empty after loading files", i, cloudInitFile.Name)
		}
	}

	return nil
}
