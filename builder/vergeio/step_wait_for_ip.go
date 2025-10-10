// This step waits for the VergeIO guest agent to report IP addresses
// This is essential for provisioning as Packer needs to know how to connect to the VM
package vergeio

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	client "github.com/vergeio/packer-plugin-vergeio/client"
)

// StepWaitForIP waits for the VM's guest agent to report IP addresses
// This step is critical for provisioning because:
// 1. It ensures the guest agent is running and functional
// 2. It discovers the IP address(es) that Packer can use to connect to the VM
// 3. It validates network connectivity is available before proceeding to SSH/WinRM
type StepWaitForIP struct {
	// WaitTimeout is the maximum time to wait for IP discovery
	// Default should be around 10 minutes to allow for VM boot + agent startup
	WaitTimeout time.Duration

	// SettleTimeout is how long to wait for the IP to remain stable
	// This prevents connection attempts during IP changes (DHCP renewals, etc.)
	SettleTimeout time.Duration

	// Config contains the builder configuration for validation
	Config *Config
}

// Run executes the IP discovery process
// This method implements the multistep.Step interface required by Packer
func (s *StepWaitForIP) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Waiting for VM to obtain IP address from guest agent...")

	// Get the cluster configuration from state (set by previous steps)
	cc := state.Get("cluster_config").(ClusterConfig)

	// Get the VM ID from state (set by StepVMCreate)
	vmId, vmIdExists := state.GetOk("vm_id")
	if !vmIdExists {
		ui.Error("VM ID not found in state - cannot wait for IP address")
		state.Put("error", fmt.Errorf("vm_id not available in build state"))
		return multistep.ActionHalt
	}

	vmIdStr := vmId.(string)
	ui.Message(fmt.Sprintf("Waiting for guest agent IP discovery for VM ID: %s", vmIdStr))

	// Set default timeouts if not configured
	waitTimeout := s.WaitTimeout
	if waitTimeout == 0 {
		waitTimeout = 10 * time.Minute // Default: 10 minutes for IP discovery
		ui.Message(fmt.Sprintf("Using default IP discovery timeout: %v", waitTimeout))
	}

	settleTimeout := s.SettleTimeout
	if settleTimeout == 0 {
		settleTimeout = 30 * time.Second // Default: 30 seconds for IP stability
		ui.Message(fmt.Sprintf("Using default IP settle timeout: %v", settleTimeout))
	}

	// Create a new VergeIO API client using the cluster configuration
	c := client.NewClient(cc.Endpoint, cc.Username, cc.Password, cc.Insecure)
	vmAPI := client.NewVMApi(c)

	ui.Message(fmt.Sprintf("Starting IP discovery process (timeout: %v, settle: %v)", waitTimeout, settleTimeout))

	// Phase 1: Wait for guest agent to become available and report IPs
	ui.Say("Phase 1: Waiting for guest agent to report network information...")

	// Create timeout context for the entire IP discovery process
	timeoutCtx, cancel := context.WithTimeout(ctx, waitTimeout)
	defer cancel()

	var discoveredIPs []string
	var lastDiscoveredIPs []string
	var stableStart time.Time

	// Poll for IP addresses every 15 seconds
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	// Initial check before starting the timer
	ui.Message("Performing initial guest agent check...")
	initialIPs, err := vmAPI.GetGuestAgentIPs(ctx, vmIdStr)
	if err == nil && len(initialIPs) > 0 {
		ui.Say(fmt.Sprintf("Guest agent immediately available with IPs: %v", initialIPs))
		discoveredIPs = initialIPs
		lastDiscoveredIPs = initialIPs
		stableStart = time.Now()
	} else {
		ui.Message(fmt.Sprintf("Initial check: %v", err))
	}

	// Main polling loop for IP discovery
	for len(discoveredIPs) == 0 {
		select {
		case <-timeoutCtx.Done():
			ui.Error(fmt.Sprintf("Timeout waiting for guest agent IP discovery (waited %v)", waitTimeout))
			ui.Error("This usually means:")
			ui.Error("  1. Guest agent is not installed in the VM")
			ui.Error("  2. Guest agent is not running or has failed to start")
			ui.Error("  3. VM networking is not properly configured")
			ui.Error("  4. VM has not fully booted yet")
			state.Put("error", fmt.Errorf("timeout waiting for guest agent IP discovery after %v", waitTimeout))
			return multistep.ActionHalt

		case <-ticker.C:
			ui.Message("Checking guest agent for IP addresses...")

			// Try to get current IP addresses from guest agent only
			currentIPs, err := vmAPI.GetGuestAgentIPs(ctx, vmIdStr)

			if err != nil {
				ui.Message(fmt.Sprintf("Guest agent not yet available: %v", err))
				continue
			}

			if len(currentIPs) == 0 {
				ui.Message("Guest agent responding but no IP addresses reported yet")
				continue
			}

			// We found IP addresses!
			ui.Say(fmt.Sprintf("Guest agent reported IP addresses: %v", currentIPs))
			discoveredIPs = currentIPs
			lastDiscoveredIPs = currentIPs
			stableStart = time.Now()
		}
	}

	// Phase 2: Wait for IP address to stabilize
	ui.Say(fmt.Sprintf("Phase 2: Waiting for IP address to stabilize (settle timeout: %v)...", settleTimeout))

	// Create settle timeout context
	settleCtx, settleCancel := context.WithTimeout(ctx, settleTimeout)
	defer settleCancel()

	settleTicker := time.NewTicker(5 * time.Second)
	defer settleTicker.Stop()

	for {
		select {
		case <-settleCtx.Done():
			// Settle timeout reached - IPs are considered stable
			ui.Say(fmt.Sprintf("IP address has been stable for %v - proceeding with: %v", settleTimeout, discoveredIPs))

		case <-settleTicker.C:
			// Check if IPs have changed during settle period
			currentIPs, err := vmAPI.GetGuestAgentIPs(ctx, vmIdStr)

			if err != nil {
				ui.Error(fmt.Sprintf("Lost guest agent connection during settle period: %v", err))
				ui.Error("IP discovery failed during settle period - guest agent connection lost")
				state.Put("error", fmt.Errorf("guest agent connection lost during settle period"))
				settleCancel() // Clean up context before returning
				return multistep.ActionHalt
			}

			// Compare current IPs with last discovered IPs
			if !ipSlicesEqual(currentIPs, lastDiscoveredIPs) {
				ui.Message(fmt.Sprintf("IP address changed during settle period: %v -> %v", lastDiscoveredIPs, currentIPs))
				ui.Message("Restarting settle timer...")
				discoveredIPs = currentIPs
				lastDiscoveredIPs = currentIPs
				stableStart = time.Now()

				// Restart settle timeout
				settleCancel()
				settleCtx, settleCancel = context.WithTimeout(ctx, settleTimeout)
				continue
			}

			// IPs are still the same - continue waiting for settle timeout
			elapsed := time.Since(stableStart)
			ui.Message(fmt.Sprintf("IP address stable for %v (need %v total)", elapsed, settleTimeout))
		}
		break // Exit the settle loop when timeout is reached
	}

	// Ensure settle context is cancelled when we exit the loop
	settleCancel()

	// Phase 3: Select and validate the IP address to use
	ui.Say("Phase 3: Selecting IP address for communication...")

	if len(discoveredIPs) == 0 {
		ui.Error("No IP addresses available after discovery process")
		state.Put("error", fmt.Errorf("no IP addresses discovered"))
		return multistep.ActionHalt
	}

	// Select the first IP address as the host for communication
	// TODO: In the future, we could add logic to prefer certain networks or IP ranges
	selectedIP := discoveredIPs[0]

	if len(discoveredIPs) > 1 {
		ui.Message(fmt.Sprintf("Multiple IP addresses available: %v", discoveredIPs))
		ui.Message(fmt.Sprintf("Using first IP address for communication: %s", selectedIP))
	} else {
		ui.Say(fmt.Sprintf("Using IP address for communication: %s", selectedIP))
	}

	// Store the selected IP address in state for the communicator
	// This is the key step that enables SSH/WinRM connectivity
	state.Put("host", selectedIP)
	state.Put("discovered_ips", discoveredIPs) // Store all IPs for reference

	ui.Say(fmt.Sprintf("IP discovery successful! VM is ready for provisioning at: %s", selectedIP))
	return multistep.ActionContinue

}

// Cleanup handles any cleanup needed if the step fails or is interrupted
// For IP discovery, there's typically no cleanup needed as we're just reading state
func (s *StepWaitForIP) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packersdk.Ui)
	ui.Message("StepWaitForIP cleanup: No cleanup required for IP discovery step")
}

// ipSlicesEqual compares two IP address slices for equality
// This helper function is used to detect when IP addresses change during the settle period
func ipSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// Create maps for comparison (order-independent)
	aMap := make(map[string]bool)
	bMap := make(map[string]bool)

	for _, ip := range a {
		aMap[ip] = true
	}

	for _, ip := range b {
		bMap[ip] = true
	}

	// Check if all IPs in a are in b
	for ip := range aMap {
		if !bMap[ip] {
			return false
		}
	}

	// Check if all IPs in b are in a
	for ip := range bMap {
		if !aMap[ip] {
			return false
		}
	}

	return true
}
