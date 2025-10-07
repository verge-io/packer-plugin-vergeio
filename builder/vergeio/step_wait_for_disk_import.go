package vergeio

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	client "github.com/vergeio/packer-plugin-vergeio/client"
)

// StepWaitForDiskImport waits for any disks with media="import" to complete importing
// before proceeding with VM power-on. This prevents the "Cannot power on a VM while
// drives are importing" error.
type StepWaitForDiskImport struct {
	Config *Config
}

func (s *StepWaitForDiskImport) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	// Check if there are any import disks to wait for
	importDiskKeysInterface := state.Get("import_disk_keys")
	if importDiskKeysInterface == nil {
		ui.Say("No import disks to wait for - proceeding with power-on")
		return multistep.ActionContinue
	}

	importDiskKeys, ok := importDiskKeysInterface.([]string)
	if !ok || len(importDiskKeys) == 0 {
		ui.Say("No import disks to wait for - proceeding with power-on")
		return multistep.ActionContinue
	}

	ui.Say(fmt.Sprintf("Waiting for %d disk(s) with media='import' to complete importing before power-on", len(importDiskKeys)))

	// Create VergeIO client
	vergeClient := client.NewClient(s.Config.Endpoint, s.Config.Username, s.Config.Password, s.Config.Insecure)
	driveAPI := client.NewDriveApi(vergeClient)

	// Wait for import completion with a reasonable retry limit
	// 10 retries * 5 seconds = 50 seconds max wait time per disk
	maxRetries := 20 // Increased to 100 seconds max wait time per disk

	err := driveAPI.WaitForDiskImportCompletion(ctx, importDiskKeys, maxRetries)
	if err != nil {
		ui.Error(fmt.Sprintf("Error waiting for disk import completion: %s", err))
		state.Put("error", fmt.Errorf("disk import failed: %w", err))
		return multistep.ActionHalt
	}

	ui.Say("All disk imports completed successfully")

	// Get the disk configurations to check sizes
	importDiskConfigsInterface := state.Get("import_disk_configs")
	if importDiskConfigsInterface != nil {
		importDiskConfigs, ok := importDiskConfigsInterface.([]client.VMDiskResourceModel)
		if ok && len(importDiskConfigs) > 0 {
			ui.Say(fmt.Sprintf("Checking and resizing %d imported disk(s) if needed", len(importDiskConfigs)))

			// Check and resize imported disks if their size doesn't match the requested size
			err = driveAPI.CheckAndResizeImportedDisks(ctx, importDiskConfigs, importDiskKeys)
			if err != nil {
				ui.Error(fmt.Sprintf("Error checking/resizing imported disks: %s", err))
				state.Put("error", fmt.Errorf("disk resize after import failed: %w", err))
				return multistep.ActionHalt
			}

			ui.Say("All imported disk size checks and resizing completed successfully")
		}
	}

	ui.Say("Ready for VM power-on")
	return multistep.ActionContinue
}

func (s *StepWaitForDiskImport) Cleanup(state multistep.StateBag) {
	// No cleanup needed for this step
	// VM cleanup is handled by StepVMCreate if needed
}
