// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package vergeio

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

const BuilderId = "packer.vergeio"

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {

	ui.Message("[VergeIO]: Starting VergeIO Packer Builder...")
	ui.Message("[VergeIO]: This will create a VM, provision it, and prepare it for use")

	log.Printf("[VergeIO]: Builder configuration - Cluster: %s, VM: %s",
		b.config.ClusterConfig.Username, b.config.VmConfig.Name)

	// Define the complete build workflow with all provisioning steps
	steps := []multistep.Step{}

	// ==========================================
	// PHASE 1: VM CREATION AND SETUP
	// ==========================================

	// Step 1: Create the VM with all hardware, disks, and NICs
	// This step handles the complete VM creation process including error recovery
	steps = append(steps, &StepVMCreate{
		ClusterConfig: b.config.ClusterConfig,
		VmConfig:      b.config.VmConfig,
	})

	// Step 2: Wait for disk imports to complete (if any disks have media="import")
	// This prevents "Cannot power on a VM while drives are importing" errors
	steps = append(steps, &StepWaitForDiskImport{
		Config: &b.config,
	})

	// Step 3: Power on the VM so the guest OS can start
	// VMs are created in powered-off state, so this is essential for provisioning
	powerOnTimeout := b.config.PowerOnTimeout
	if powerOnTimeout == 0 {
		powerOnTimeout = 2 * time.Minute // Default: 2 minutes
	}

	bootTimeout := b.config.BootTimeout
	if bootTimeout == 0 {
		bootTimeout = 5 * time.Minute // Default: 5 minutes
	}

	steps = append(steps, &StepPowerOn{
		PowerOnTimeout: powerOnTimeout, // User-configured or default timeout
		BootTimeout:    bootTimeout,    // User-configured or default timeout
	})

	// ==========================================
	// PHASE 2: NETWORK DISCOVERY AND CONNECTIVITY
	// ==========================================

	// Step 4: Wait for guest agent to report IP addresses
	// TODO: TEMPORARILY DISABLED - Guest agent IP discovery not working
	// This step discovers the VM's IP address(es) needed for SSH/WinRM connectivity
	// steps = append(steps, &StepWaitForIP{
	// 	WaitTimeout:   10 * time.Minute, // Maximum time to wait for IP discovery
	// 	SettleTimeout: 30 * time.Second, // Time for IP to remain stable
	// 	Config:        &b.config,        // Pass config for potential network filtering
	// })

	// ==========================================
	// PHASE 3: PROVISIONING
	// ==========================================

	// Step 4: Connect to the VM via SSH/WinRM
	// This uses Packer's standard communicator step to establish connectivity
	steps = append(steps, &communicator.StepConnect{
		Config:    &b.config.Comm,
		Host:      b.getHostFunc(),
		SSHConfig: b.config.Comm.SSHConfigFunc(),
	})

	// Step 5: Run all configured provisioners
	// This is where shell scripts, file uploads, Ansible, etc. are executed
	steps = append(steps, &commonsteps.StepProvision{})

	// ==========================================
	// PHASE 4: CLEANUP AND FINALIZATION
	// ==========================================

	// Step 6: Gracefully shut down the VM via SSH/WinRM
	// This ensures the VM is in a clean state and all changes are persisted
	steps = append(steps, &StepShutdown{
		Command: b.config.ShutdownCommand, // User-configured shutdown command
		Timeout: b.config.ShutdownTimeout, // How long to wait for shutdown
	})

	// ==========================================
	// EXECUTION SETUP
	// ==========================================

	// Setup the state bag with initial configuration and context
	// The state bag is how information flows between build steps
	state := new(multistep.BasicStateBag)
	state.Put("cluster_config", b.config.ClusterConfig) // VergeIO connection info
	state.Put("vm_config", b.config.VmConfig)           // VM specification
	state.Put("hook", hook)                             // Packer hooks for events
	state.Put("ui", ui)                                 // User interface for output
	state.Put("config", &b.config)                      // Complete configuration

	// Initialize generated data storage for provisioners and post-processors
	state.Put("generated_data", map[string]interface{}{
		"vm_name":   b.config.VmConfig.Name,
		"os_family": b.config.VmConfig.OSFamily,
		"cpu_cores": b.config.VmConfig.CPUCores,
		"ram":       b.config.VmConfig.RAM,
	})

	ui.Message("[VergeIO]: Starting build workflow with the following phases:")
	ui.Message("  Phase 1: VM Creation (VM + Disks + NICs)")
	ui.Message("  Phase 2: Disk Import Completion + Power Management")
	ui.Message("  Phase 3: Network Discovery and Connectivity")
	ui.Message("  Phase 4: Provisioning via SSH/WinRM")
	ui.Message("  Phase 5: Cleanup and Shutdown")

	// ==========================================
	// EXECUTION
	// ==========================================

	// Execute the complete workflow
	// The runner will execute each step in sequence and handle failures
	b.runner = &multistep.BasicRunner{Steps: steps}
	b.runner.Run(ctx, state)

	// Check if any step failed and return the error
	if err, ok := state.GetOk("error"); ok {
		log.Printf("[VergeIO]: Build failed with error: %v", err)
		return nil, err.(error)
	}

	// ==========================================
	// SUCCESS - CREATE ARTIFACT
	// ==========================================

	ui.Message("[VergeIO]: Build completed successfully!")

	// Create the build artifact containing information about the created VM
	// This can be used by post-processors for further processing
	artifact := &Artifact{
		StateData: map[string]interface{}{
			"generated_data": state.Get("generated_data"),
			"vm_id":          state.Get("vm_id"),
			"machine_id":     state.Get("machine_id"),
			"discovered_ips": state.Get("discovered_ips"),
		},
	}

	log.Printf("[VergeIO]: Build artifact created successfully")
	return artifact, nil
}

// getHostFunc returns a function that retrieves the host for communication
// This function reads the "host" key from the state bag, which is set by StepWaitForIP
func (b *Builder) getHostFunc() func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		host := state.Get("host")
		if host == nil {
			return "", fmt.Errorf("no host found in state - IP discovery may have failed")
		}
		return host.(string), nil
	}
}
