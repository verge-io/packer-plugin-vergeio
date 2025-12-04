// This step creates a VM in VergeIO
package vergeio

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	client "github.com/verge-io/packer-plugin-vergeio/client"
)

// This is a definition of a builder step and should implement multistep.Step
type StepVMCreate struct {
	ClusterConfig ClusterConfig
	VmConfig      VmConfig
}

func (s *StepVMCreate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Running StepVMCreate")

	cc := state.Get("cluster_config").(ClusterConfig)
	vm := state.Get("vm_config").(VmConfig)

	// Create a new client instance
	c := client.NewClient(cc.Endpoint, cc.Username, cc.Password, cc.Insecure)
	vmAPI := client.NewVMApi(c)
	driveAPI := client.NewDriveApi(c)
	nicAPI := client.NewNicApi(c)

	// Prepare the API data packet from the plan
	apiData := client.VMAPIResourceModel{
		Machine:              vm.Machine,
		Name:                 vm.Name,
		Cluster:              vm.Cluster,
		Description:          vm.Description,
		Enabled:              vm.Enabled,
		MachineType:          vm.MachineType,
		AllowHotplug:         vm.AllowHotplug,
		DisablePowercycle:    vm.DisablePowercycle,
		CPUCores:             vm.CPUCores,
		CPUType:              vm.CPUType,
		RAM:                  vm.RAM,
		Console:              vm.Console,
		Display:              vm.Display,
		Video:                vm.Video,
		Sound:                vm.Sound,
		OSFamily:             vm.OSFamily,
		OSDescription:        vm.OSDescription,
		RTCBase:              vm.RTCBase,
		BootOrder:            vm.BootOrder,
		ConsolePassEnabled:   vm.ConsolePassEnabled,
		ConsolePass:          vm.ConsolePass,
		USBTablet:            vm.USBTablet,
		UEFI:                 vm.UEFI,
		SecureBoot:           vm.SecureBoot,
		SerialPort:           vm.SerialPort,
		BootDelay:            vm.BootDelay,
		PreferredNode:        vm.PreferredNode,
		SnapshotProfile:      vm.SnapshotProfile,
		CloudInitDataSource:  vm.CloudInitDataSource,
		GuestAgent:           vm.GuestAgent,
		HAGroup:              vm.HAGroup,
		Advanced:             vm.Advanced,
		NestedVirtualization: vm.NestedVirtualization,
		DisableHypervisor:    vm.DisableHypervisor,
		VmDisks:              []interface{}{},
	}

	// Add the cloud init files
	if vm.CloudInitFiles != nil {
		for _, cloudInitFile := range vm.CloudInitFiles {
			apiData.CloudInitFiles = append(apiData.CloudInitFiles, client.CloudInitFileAPI{
				Name:     cloudInitFile.Name,
				Contents: cloudInitFile.Contents,
			})
		}
	}

	// post the data to the API
	err := vmAPI.CreateVM(ctx, &apiData) // vmAPI.Post(client.VMActionEndpoint, bytes.NewBuffer(bytesPayload))
	if err != nil {
		ui.Error(fmt.Sprintf("Error making POST request to %s: %s", client.VMEndpoint, err))
		state.Put("error", fmt.Errorf("error making POST request to %s: %w", client.VMActionEndpoint, err))
		return multistep.ActionHalt
	}

	// Get the machine ID from the created VM (populated by CreateVM -> readVM)
	machineID := apiData.Machine
	if machineID == 0 {
		ui.Error("Failed to retrieve machine ID from created VM")
		state.Put("error", fmt.Errorf("failed to retrieve machine ID from created VM"))
		return multistep.ActionHalt
	}
	ui.Say(fmt.Sprintf("VM created successfully with Machine ID: %d", machineID))

	// Store the machine ID and VM ID in state for other steps to use
	state.Put("machine_id", machineID)
	state.Put("vm_id", apiData.Id) // Store VM ID for cleanup purposes

	// Create disks if any are defined
	var importDiskKeys []string                        // Track disks that need import completion waiting
	var importDiskConfigs []client.VMDiskResourceModel // Track disk configurations for size checking
	if vm.VmDiskConfigs != nil {
		for _, disk := range vm.VmDiskConfigs {
			diskData := client.VMDiskResourceModel{
				Machine:             machineID, // Use the actual machine ID from the created VM
				Name:                disk.Name,
				Description:         disk.Description,
				Interface:           disk.Interface,
				Media:               disk.Media,
				MediaSource:         disk.MediaSource,
				PreferredTier:       disk.PreferredTier,
				DiskSize:            disk.DiskSize,
				Enabled:             disk.Enabled,
				ReadOnly:            disk.ReadOnly,
				Serial:              disk.Serial,
				Asset:               disk.Asset,
				OrderId:             disk.OrderId,
				PreserveDriveFormat: disk.PreserveDriveFormat,
			}

			ui.Say(fmt.Sprintf("Creating disk '%s' for VM '%s' (Machine ID: %d)", disk.Name, vm.Name, machineID))
			diskKey, err := driveAPI.CreateVMDiskWithKey(ctx, &diskData)
			if err != nil {
				ui.Error(fmt.Sprintf("Error creating disk '%s': %s", disk.Name, err))
				ui.Error(fmt.Sprintf("Disk creation failed - VM '%s' will be cleaned up automatically", vm.Name))
				state.Put("error", fmt.Errorf("error creating disk '%s': %w", disk.Name, err))
				// Mark VM for cleanup - the Cleanup method will handle deletion
				state.Put("vm_creation_failed", true)
				return multistep.ActionHalt
			}
			ui.Say(fmt.Sprintf("Successfully created disk '%s'", disk.Name))

			// Track disks that need import completion waiting
			if disk.Media == "import" {
				log.Printf("[VergeIO]: Disk '%s' with media='import' will need import completion waiting (key: %s)", disk.Name, diskKey)
				importDiskKeys = append(importDiskKeys, diskKey)
				importDiskConfigs = append(importDiskConfigs, diskData)
			}
		}
		ui.Say(fmt.Sprintf("Successfully created %d disk(s) for VM '%s'", len(vm.VmDiskConfigs), vm.Name))
	}

	// Store import disk keys and configs in state for StepWaitForDiskImport to use
	if len(importDiskKeys) > 0 {
		log.Printf("[VergeIO]: Storing %d import disk keys and configs in state for import completion waiting and size checking", len(importDiskKeys))
		state.Put("import_disk_keys", importDiskKeys)
		state.Put("import_disk_configs", importDiskConfigs)
	}

	// Create NICs if any are defined
	if vm.VmNicConfigs != nil {
		for _, nic := range vm.VmNicConfigs {
			nicData := client.VMNicResourceModel{
				Machine:         machineID, // Use the actual machine ID from the created VM
				Name:            nic.Name,
				Description:     nic.Description,
				Interface:       nic.Interface,
				Driver:          nic.Driver,
				Model:           nic.Model,
				VNET:            nic.VNET,
				MAC:             nic.MAC,
				IPAddress:       nic.IPAddress,
				AssignIPAddress: nic.AssignIPAddress,
				Enabled:         nic.Enabled,
			}

			ui.Say(fmt.Sprintf("Creating NIC '%s' for VM '%s' (Machine ID: %d)", nic.Name, vm.Name, machineID))
			err := nicAPI.CreateVMNic(ctx, &nicData)
			if err != nil {
				ui.Error(fmt.Sprintf("Error creating NIC '%s': %s", nic.Name, err))
				ui.Error(fmt.Sprintf("NIC creation failed - VM '%s' will be cleaned up automatically", vm.Name))
				state.Put("error", fmt.Errorf("error creating NIC '%s': %w", nic.Name, err))
				// Mark VM for cleanup - the Cleanup method will handle deletion
				state.Put("vm_creation_failed", true)
				return multistep.ActionHalt
			}
			ui.Say(fmt.Sprintf("Successfully created NIC '%s'", nic.Name))
		}
		ui.Say(fmt.Sprintf("Successfully created %d NIC(s) for VM '%s'", len(vm.VmNicConfigs), vm.Name))
	}

	// VM and all components created successfully
	ui.Say(fmt.Sprintf("VM '%s' and all components created successfully!", vm.Name))
	return multistep.ActionContinue
}

func (s *StepVMCreate) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packersdk.Ui)

	// Check if we need to clean up a VM
	vmId, vmIdExists := state.GetOk("vm_id")
	creationFailed, failureExists := state.GetOk("vm_creation_failed")

	// Only cleanup if VM was created but something went wrong
	if vmIdExists && failureExists && creationFailed.(bool) {
		ui.Say(fmt.Sprintf("Cleaning up failed VM creation - deleting VM ID: %s", vmId.(string)))

		// Get cluster config to create API client
		cc, ccExists := state.GetOk("cluster_config")
		if !ccExists {
			ui.Error("Cannot cleanup VM: cluster configuration not found in state")
			return
		}

		clusterConfig := cc.(ClusterConfig)
		c := client.NewClient(clusterConfig.Endpoint, clusterConfig.Username, clusterConfig.Password, clusterConfig.Insecure)
		vmAPI := client.NewVMApi(c)

		// Attempt to delete the VM (this will also delete all associated disks)
		ctx := context.Background()
		err := vmAPI.DeleteVM(ctx, vmId.(string))
		if err != nil {
			ui.Error(fmt.Sprintf("Failed to cleanup VM %s: %s", vmId.(string), err))
			ui.Error("Manual cleanup may be required in VergeIO console")
		} else {
			ui.Say(fmt.Sprintf("Successfully cleaned up VM %s and all associated resources", vmId.(string)))
		}
	} else {
		ui.Say("No cleanup required for StepVMCreate")
	}
}
