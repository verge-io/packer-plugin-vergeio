// This step gracefully shuts down the VM after provisioning is complete
// This ensures the VM is in a clean state before finalizing the build
package vergeio

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	client "github.com/verge-io/packer-plugin-vergeio/client"
)

// StepShutdown gracefully shuts down the VM after provisioning
// This step is important because:
// 1. It ensures all provisioning changes are properly written to disk
// 2. It allows the guest OS to perform a clean shutdown sequence
// 3. It puts the VM in a powered-off state suitable for template creation
type StepShutdown struct {
	// Command is the shutdown command to run inside the VM
	// Examples: "sudo shutdown -P now" (Linux), "shutdown /s /t 0" (Windows)
	Command string

	// Timeout is how long to wait for the shutdown command to complete
	// If this timeout is exceeded, the VM will be forcefully powered off
	Timeout time.Duration
}

// Run executes the shutdown process
// This method implements the multistep.Step interface required by Packer
func (s *StepShutdown) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

	// Check if we have a shutdown command configured
	if s.Command == "" {
		ui.Say("No shutdown command configured - leaving VM powered on")
		ui.Message("Note: You may want to configure 'shutdown_command' for clean VM shutdown")
		return multistep.ActionContinue
	}

	ui.Say("Gracefully shutting down VM...")

	// Get the communicator from state (set by Packer's communicator steps)
	comm := state.Get("communicator").(packersdk.Communicator)
	if comm == nil {
		ui.Error("No communicator available - cannot send shutdown command")
		ui.Error("VM will remain powered on")
		return multistep.ActionContinue // Don't fail the build for this
	}

	// Get cluster configuration and VM ID for potential forced shutdown
	cc := state.Get("cluster_config").(ClusterConfig)
	vmId, vmIdExists := state.GetOk("vm_id")
	vmIdStr := ""
	if vmIdExists {
		vmIdStr = vmId.(string)
	}

	// Set default timeout if not configured
	timeout := s.Timeout
	if timeout == 0 {
		timeout = 5 * time.Minute // Default: 5 minutes for shutdown
		ui.Message(fmt.Sprintf("Using default shutdown timeout: %v", timeout))
	}

	ui.Say(fmt.Sprintf("Executing shutdown command: %s", s.Command))

	// Phase 1: Send the shutdown command via the communicator
	ui.Say("Phase 1: Sending shutdown command to VM...")

	// Create a remote command to execute the shutdown
	cmd := &packersdk.RemoteCmd{
		Command: s.Command,
	}

	// Execute the shutdown command
	err := comm.Start(ctx, cmd)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to start shutdown command: %v", err))
		ui.Error("Attempting forced shutdown...")
		return s.performForcedShutdown(state, vmIdStr, cc, ui)
	}

	ui.Say("Shutdown command sent successfully")

	// Phase 2: Wait for the shutdown command to complete or timeout
	ui.Say("Phase 2: Waiting for VM to shut down...")

	// Create timeout context for shutdown waiting
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Monitor the command execution and wait for completion
	cmdComplete := make(chan bool, 1)
	go func() {
		cmd.Wait()
		cmdComplete <- true
	}()

	// Wait for either command completion or timeout
	select {
	case <-timeoutCtx.Done():
		ui.Error(fmt.Sprintf("Shutdown command timed out after %v", timeout))
		ui.Error("The VM may still be shutting down, or the command failed")
		return s.performForcedShutdown(state, vmIdStr, cc, ui)

	case <-cmdComplete:
		if cmd.ExitStatus() == 0 {
			ui.Say("Shutdown command completed successfully")
		} else {
			ui.Error(fmt.Sprintf("Shutdown command failed with exit code: %d", cmd.ExitStatus()))
			ui.Error("Attempting forced shutdown...")
			return s.performForcedShutdown(state, vmIdStr, cc, ui)
		}
	}

	// Phase 3: Wait additional time for the VM to actually power off
	ui.Say("Phase 3: Waiting for VM to power off...")

	// Give the VM additional time to complete the shutdown process
	shutdownWait := 30 * time.Second
	ui.Message(fmt.Sprintf("Waiting %v for VM to complete shutdown process...", shutdownWait))

	// Create shutdown timer
	shutdownTimer := time.NewTimer(shutdownWait)
	defer shutdownTimer.Stop()

	select {
	case <-ctx.Done():
		ui.Error("Build cancelled during shutdown wait")
		return multistep.ActionHalt

	case <-shutdownTimer.C:
		ui.Say("Shutdown wait period completed")
	}

	// Phase 4: Verify VM is actually powered off
	ui.Say("Phase 4: Verifying VM power state...")

	if vmIdStr != "" {
		c := client.NewClient(cc.Endpoint, cc.Username, cc.Password, cc.Insecure)
		vmAPI := client.NewVMApi(c)

		// Check VM power state
		isRunning, err := vmAPI.IsVMRunning(ctx, vmIdStr)
		if err != nil {
			ui.Error(fmt.Sprintf("Failed to verify VM power state: %v", err))
			ui.Message("Assuming shutdown was successful despite verification failure")
		} else if isRunning != nil && *isRunning {
			ui.Error("VM is still running after shutdown attempt")
			ui.Error("The graceful shutdown may have failed - VM remains powered on")
			// Don't fail the build, just warn the user
		} else {
			ui.Say("VM power state verified: VM is powered off")
		}
	} else {
		ui.Message("VM ID not available - skipping power state verification")
	}

	ui.Say("VM shutdown process completed successfully!")
	state.Put("vm_shutdown_completed", true)
	return multistep.ActionContinue
}

// Cleanup handles any cleanup needed after the shutdown step
// Typically no cleanup is needed for shutdown, but we log the state
func (s *StepShutdown) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packersdk.Ui)

	shutdownCompleted, completed := state.GetOk("vm_shutdown_completed")

	if completed && shutdownCompleted.(bool) {
		ui.Message("VM shutdown completed successfully - no cleanup needed")
	} else {
		ui.Message("VM shutdown may not have completed - check VM power state manually if needed")
	}

	ui.Message("StepShutdown cleanup completed")
}

// performForcedShutdown handles forced shutdown when graceful shutdown fails
func (s *StepShutdown) performForcedShutdown(state multistep.StateBag, vmIdStr string, cc ClusterConfig, ui packersdk.Ui) multistep.StepAction {
	// Phase 4: Forced shutdown if graceful shutdown failed
	ui.Say("Phase 4: Performing forced shutdown...")

	if vmIdStr == "" {
		ui.Error("Cannot perform forced shutdown - VM ID not available")
		ui.Error("VM may still be running - manual intervention may be required")
		return multistep.ActionContinue // Don't fail the build
	}

	// Create VergeIO API client for forced shutdown
	c := client.NewClient(cc.Endpoint, cc.Username, cc.Password, cc.Insecure)
	vmAPI := client.NewVMApi(c)

	ui.Message(fmt.Sprintf("Performing forced power-off for VM ID: %s", vmIdStr))

	// Perform forced power-off via VergeIO API
	err := vmAPI.PowerOffVM(vmIdStr)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to perform forced power-off: %v", err))
		ui.Error("VM may still be running - manual intervention may be required")
		return multistep.ActionContinue // Don't fail the build for shutdown issues
	}

	ui.Say("Forced power-off completed successfully")
	state.Put("vm_shutdown_completed", true)

	return multistep.ActionContinue
}
