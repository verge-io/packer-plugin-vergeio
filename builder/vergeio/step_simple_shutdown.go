// This step powers off the VM using the VergeIO API
// This is a simple shutdown that just calls the power-off API
package vergeio

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	client "github.com/verge-io/packer-plugin-vergeio/client"
)

// StepSimpleShutdown powers off the VM using the VergeIO API
// This step is used when we want to simply power off the VM without SSH/WinRM connectivity
type StepSimpleShutdown struct{}

// Run executes the power-off process
func (s *StepSimpleShutdown) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Powering off VM...")

	// Get the cluster configuration from state
	cc := state.Get("cluster_config").(ClusterConfig)

	// Get the VM Key from state (set by StepVMCreate)
	vmId, vmIdExists := state.GetOk("vm_id")
	if !vmIdExists {
		ui.Error("VM Key not found in state - cannot power off VM")
		ui.Message("VM may still be running - manual intervention may be required")
		return multistep.ActionContinue // Don't fail the build for this
	}

	vmKeyStr := vmId.(string)
	ui.Message(fmt.Sprintf("Powering off VM with Key: %s", vmKeyStr))

	// Create a new VergeIO API client
	c := client.NewClient(cc.Endpoint, cc.Username, cc.Password, cc.Insecure)
	vmAPI := client.NewVMApi(c)

	// Call PowerOffVM to shut down the VM
	err := vmAPI.PowerOffVM(vmKeyStr)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to power off VM: %v", err))
		ui.Error("VM may still be running - manual intervention may be required")
		return multistep.ActionContinue // Don't fail the build for power-off issues
	}

	ui.Say("VM powered off successfully!")
	state.Put("vm_powered_off", true)

	return multistep.ActionContinue
}

// Cleanup handles any cleanup needed after the shutdown step
func (s *StepSimpleShutdown) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packersdk.Ui)

	poweredOff, exists := state.GetOk("vm_powered_off")

	if exists && poweredOff.(bool) {
		ui.Message("VM was powered off successfully - no cleanup needed")
	} else {
		ui.Message("VM power-off status unknown - check VM power state manually if needed")
	}

	ui.Message("StepSimpleShutdown cleanup completed")
}
