// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc mapstructure-to-hcl2 -type VMConfig,VMOutput
package vergeio

import (
	"context"
	"fmt"
	"log"

	client "github.com/vergeio/packer-plugin-vergeio/client"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/zclconf/go-cty/cty"
)

type VMConfig struct {
	// VergeIO connection configuration
	Username string `mapstructure:"vergeio_username" required:"true"`
	Password string `mapstructure:"vergeio_password" required:"true"`
	Endpoint string `mapstructure:"vergeio_endpoint" required:"true"`
	Port     int    `mapstructure:"vergeio_port" required:"false"`
	Insecure bool   `mapstructure:"vergeio_insecure" required:"false"`

	// Filter options for VM query
	FilterName   string `mapstructure:"filter_name" required:"false"`
	FilterId     int    `mapstructure:"filter_id" required:"false"`
	IsSnapshot   bool   `mapstructure:"is_snapshot" required:"false"`
}

type VMDataSource struct {
	config VMConfig
}

type VMInfo struct {
	ID          int32          `mapstructure:"id"`
	Name        string         `mapstructure:"name"`
	Key         int32          `mapstructure:"key"`
	IsSnapshot  bool           `mapstructure:"is_snapshot"`
	CPUType     string         `mapstructure:"cpu_type"`
	MachineType string         `mapstructure:"machine_type"`
	OSFamily    string         `mapstructure:"os_family"`
	UEFI        bool           `mapstructure:"uefi"`
	Drives      []VMDriveInfo  `mapstructure:"drives"`
	Nics        []VMNicInfo    `mapstructure:"nics"`
}

type VMDriveInfo struct {
	Key           int32                   `mapstructure:"key"`
	Name          string                  `mapstructure:"name"`
	Interface     string                  `mapstructure:"interface"`
	Media         string                  `mapstructure:"media"`
	Description   string                  `mapstructure:"description"`
	PreferredTier string                  `mapstructure:"preferred_tier"`
	MediaSource   *VMDriveMediaSourceInfo `mapstructure:"media_source"`
}

type VMDriveMediaSourceInfo struct {
	Key            int32 `mapstructure:"key"`
	UsedBytes      int64 `mapstructure:"used_bytes"`
	AllocatedBytes int64 `mapstructure:"allocated_bytes"`
	Filesize       int64 `mapstructure:"filesize"`
}

type VMNicInfo struct {
	Key        int32  `mapstructure:"key"`
	Name       string `mapstructure:"name"`
	Interface  string `mapstructure:"interface"`
	Vnet       string `mapstructure:"vnet"`
	Status     string `mapstructure:"status"`
	Ipaddress  string `mapstructure:"ipaddress"`
	MacAddress string `mapstructure:"macaddress"`
}

type VMOutput struct {
	VMs []VMInfo `mapstructure:"vms"`
}

func (d *VMDataSource) ConfigSpec() hcldec.ObjectSpec {
	return d.config.FlatMapstructure().HCL2Spec()
}

func (d *VMDataSource) Configure(raws ...interface{}) error {
	err := config.Decode(&d.config, nil, raws...)
	if err != nil {
		return err
	}

	// Set defaults
	if d.config.Port == 0 {
		d.config.Port = 443
	}

	// Validate required fields
	if d.config.Username == "" {
		return fmt.Errorf("vergeio_username is required")
	}
	if d.config.Password == "" {
		return fmt.Errorf("vergeio_password is required")
	}
	if d.config.Endpoint == "" {
		return fmt.Errorf("vergeio_endpoint is required")
	}

	log.Printf("[VergeIO VM DataSource]: Configured to connect to %s with user %s", 
		d.config.Endpoint, d.config.Username)
	log.Printf("[VergeIO VM DataSource]: Filter settings - name='%s', id=%d, is_snapshot=%t", 
		d.config.FilterName, d.config.FilterId, d.config.IsSnapshot)

	return nil
}

func (d *VMDataSource) OutputSpec() hcldec.ObjectSpec {
	return (&VMOutput{}).FlatMapstructure().HCL2Spec()
}

func (d *VMDataSource) Execute() (cty.Value, error) {
	log.Printf("[VergeIO VM DataSource]: Starting VM data source execution")

	// Create VergeIO client using the configured credentials
	vergeClient := client.NewClient(d.config.Endpoint, d.config.Username, d.config.Password, d.config.Insecure)
	vmAPI := client.NewVMApi(vergeClient)

	// Query VMs from VergeIO API
	vms, err := vmAPI.GetVMs(context.Background(), d.config.FilterName, d.config.FilterId, d.config.IsSnapshot)
	if err != nil {
		return cty.NilVal, fmt.Errorf("failed to get VMs from VergeIO API: %w", err)
	}

	log.Printf("[VergeIO VM DataSource]: Found %d VMs from VergeIO API", len(vms))

	// Convert to output format
	var vmInfos []VMInfo
	for _, vm := range vms {
		vmInfo := VMInfo{
			ID:          vm.ID,
			Name:        vm.Name,
			Key:         vm.Key,
			IsSnapshot:  vm.IsSnapshot,
			CPUType:     vm.CPUType,
			MachineType: vm.MachineType,
			OSFamily:    vm.OSFamily,
			UEFI:        vm.UEFI,
		}

		// Process drives
		if vm.Drives != nil {
			for _, drive := range vm.Drives {
				driveInfo := VMDriveInfo{
					Key:           drive.Key,
					Name:          drive.Name,
					Interface:     drive.Interface,
					Media:         drive.Media,
					Description:   drive.Description,
					PreferredTier: drive.PreferredTier,
				}
				
				if drive.MediaSource != nil {
					driveInfo.MediaSource = &VMDriveMediaSourceInfo{
						Key:            drive.MediaSource.Key,
						UsedBytes:      drive.MediaSource.UsedBytes,
						AllocatedBytes: drive.MediaSource.AllocatedBytes,
						Filesize:       drive.MediaSource.Filesize,
					}
				}
				
				vmInfo.Drives = append(vmInfo.Drives, driveInfo)
			}
		}

		// Process nics
		if vm.Nics != nil {
			for _, nic := range vm.Nics {
				nicInfo := VMNicInfo{
					Key:        nic.Key,
					Name:       nic.Name,
					Interface:  nic.Interface,
					Vnet:       nic.Vnet,
					Status:     nic.Status,
					Ipaddress:  nic.Ipaddress,
					MacAddress: nic.MacAddress,
				}
				vmInfo.Nics = append(vmInfo.Nics, nicInfo)
			}
		}

		vmInfos = append(vmInfos, vmInfo)
		log.Printf("[VergeIO VM DataSource]: VM - ID: %d, Name: %s, Key: %d, IsSnapshot: %t, Drives: %d, NICs: %d", 
			vm.ID, vm.Name, vm.Key, vm.IsSnapshot, len(vmInfo.Drives), len(vmInfo.Nics))
	}

	output := VMOutput{
		VMs: vmInfos,
	}

	log.Printf("[VergeIO VM DataSource]: Successfully processed %d VMs from VergeIO", len(vmInfos))
	return hcl2helper.HCL2ValueFromConfig(output, d.OutputSpec()), nil
}