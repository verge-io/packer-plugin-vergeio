// This step powers on the VM after creation and waits for it to be running
// This is essential because VMs are created in a powered-off state
package vergeio

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	client "github.com/vergeio/packer-plugin-vergeio/client"
)

// StepPowerOn powers on the VM and verifies it's running
// This step is necessary because VMs are created in a powered-off state by default
// SSH connectivity and boot waiting is handled by Packer's communicator step
type StepPowerOn struct {
	// PowerOnTimeout is the maximum time to wait for the VM to power on
	// This should be relatively short as power-on is usually quick
	PowerOnTimeout time.Duration

	// BootTimeout is kept for backward compatibility but not used
	// Boot waiting is handled by Packer's SSH communicator
	BootTimeout time.Duration
}

// extractIPFromCloudInit parses cloud-init network-config to extract the static IP address
func (s *StepPowerOn) extractIPFromCloudInit(config *Config) (string, error) {
	// Look for network-config cloud-init file
	for _, cloudInitFile := range config.VmConfig.CloudInitFiles {
		if cloudInitFile.Name == "network-config" {
			// Use regex to extract IP address from CIDR notation (e.g., "192.168.1.100/24" -> "192.168.1.100")
			ipRegex := regexp.MustCompile(`(?:^|\s+)(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\/\d{1,2}`)
			matches := ipRegex.FindStringSubmatch(cloudInitFile.Contents)
			if len(matches) >= 2 {
				return matches[1], nil
			}
		}
	}
	return "", fmt.Errorf("no static IP address found in cloud-init network-config")
}

// Run executes the power-on process
// This method implements the multistep.Step interface required by Packer
func (s *StepPowerOn) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Powering on VM...")

	// Get the cluster configuration from state (set by previous steps)
	cc := state.Get("cluster_config").(ClusterConfig)

	// Get the VM Key from state (set by StepVMCreate) for power operations
	vmId, vmIdExists := state.GetOk("vm_id")
	if !vmIdExists {
		ui.Error("VM Key not found in state - cannot power on VM")
		state.Put("error", fmt.Errorf("vm_id (key) not available in build state"))
		return multistep.ActionHalt
	}

	vmKeyStr := vmId.(string)
	ui.Message(fmt.Sprintf("Powering on VM with Key: %s", vmKeyStr))

	// Set default timeouts if not configured
	powerOnTimeout := s.PowerOnTimeout
	if powerOnTimeout == 0 {
		powerOnTimeout = 2 * time.Minute // Default: 2 minutes for power-on
		ui.Message(fmt.Sprintf("Using default power-on timeout: %v", powerOnTimeout))
	}

	bootTimeout := s.BootTimeout
	if bootTimeout == 0 {
		bootTimeout = 5 * time.Minute // Default: 5 minutes for full boot
		ui.Message(fmt.Sprintf("Using default boot timeout: %v", bootTimeout))
	}

	// Create a new VergeIO API client using the cluster configuration
	c := client.NewClient(cc.Endpoint, cc.Username, cc.Password, cc.Insecure)
	vmAPI := client.NewVMApi(c)

	// Power on the VM
	ui.Say("Sending power-on command to VM...")

	// Call PowerOnVM with the VM Key
	err := vmAPI.PowerOnVM(vmKeyStr)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to power on VM: %v", err))
		state.Put("error", fmt.Errorf("failed to power on VM: %w", err))
		return multistep.ActionHalt
	}

	ui.Say("Power-on command sent successfully")

	// Wait for VM to reach running state
	ui.Say("Waiting for VM to reach running state...")

	// Create timeout context for power-on verification
	timeoutCtx, cancel := context.WithTimeout(ctx, powerOnTimeout)
	defer cancel()

	// Poll every 5 seconds to check power state
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			ui.Error(fmt.Sprintf("Timeout waiting for VM to power on (waited %v)", powerOnTimeout))
			ui.Error("The VM may have hardware issues or insufficient resources")
			state.Put("error", fmt.Errorf("timeout waiting for VM to power on after %v", powerOnTimeout))
			// Mark for cleanup
			state.Put("vm_power_on_failed", true)
			return multistep.ActionHalt

		case <-ticker.C:
			ui.Message("Checking VM power state...")

			// Use VergeIO API to check if VM is actually running
			isRunning, err := vmAPI.IsVMRunning(ctx, vmKeyStr)
			if err != nil {
				ui.Error(fmt.Sprintf("Failed to check VM power state: %v", err))
				ui.Message("Continuing anyway - VM may still be starting up")
				continue
			}

			if isRunning != nil && *isRunning {
				ui.Say("VM is now powered on and running")
				goto bootWait
			} else {
				ui.Message("VM is not yet running, continuing to wait...")
				continue
			}
		}
	}

bootWait:

	// Mark that VM has been powered on for cleanup purposes
	state.Put("vm_powered_on", true)

	// Try to extract static IP from cloud-init network configuration (optional)
	config := state.Get("config").(*Config)
	staticIP, err := s.extractIPFromCloudInit(config)
	if err != nil {
		// Static IP extraction failed - this is OK if guest agent is enabled
		ui.Message("No static IP found in cloud-init network-config")
		if config.GuestAgent {
			ui.Message("Guest agent is enabled - IP discovery will be handled by next step")
		} else {
			ui.Error("No static IP configured and guest_agent is disabled")
			ui.Error("Must either:")
			ui.Error("  1. Configure static IP in cloud-init network-config, or")
			ui.Error("  2. Enable guest_agent = true for automatic IP discovery")
			state.Put("error", fmt.Errorf("no IP discovery method available - enable guest_agent or configure static IP"))
			return multistep.ActionHalt
		}
	} else {
		// Static IP found - use it
		ui.Say(fmt.Sprintf("Using static IP from cloud-init network-config: %s", staticIP))
		state.Put("host", staticIP)
		state.Put("discovered_ips", []string{staticIP})
		ui.Message("Static IP configured - guest agent IP discovery will be skipped")
	}

	ui.Say("VM power-on completed successfully!")
	if staticIP != "" {
		ui.Message(fmt.Sprintf("VM is now running at %s - SSH connectivity will be handled by Packer's communicator", staticIP))
	} else {
		ui.Message("VM is now running - IP discovery will be handled by next step")
	}

	return multistep.ActionContinue
}

// Cleanup handles cleanup if the power-on step fails or is interrupted
// This ensures we don't leave VMs in an unknown power state
func (s *StepPowerOn) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packersdk.Ui)

	// Check if power-on failed and we need to clean up
	powerOnFailed, failureExists := state.GetOk("vm_power_on_failed")
	poweredOn, poweredOnExists := state.GetOk("vm_powered_on")

	if failureExists && powerOnFailed.(bool) {
		ui.Say("Power-on failed - VM cleanup will be handled by StepVMCreate")
		// The VM creation step will handle deletion if needed
		return
	}

	if poweredOnExists && poweredOn.(bool) {
		ui.Message("VM was powered on successfully - no power-on cleanup needed")
		// We leave the VM powered on for provisioning
		// The shutdown step will handle graceful shutdown later
		return
	}

	ui.Message("StepPowerOn cleanup: No cleanup required")
}
