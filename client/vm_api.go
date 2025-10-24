// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MIT

package vergeio

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"strings"
	"time"
)

const (
	VMEndpoint       = APIEndpoint + "/vms"
	VMActionEndpoint = APIEndpoint + "/vm_actions"
)

func NewVMApi(c *Client) *VMApi {
	return &VMApi{
		name:   "VM Api",
		client: c,
	}
}

type VMApi struct {
	name   string
	client *Client
}

func (va *VMApi) Name() string {
	return va.name
}

func getValidOSFamilies() []string {
	return []string{
		"linux",
		"windows",
		"freebsd",
		"other",
	}
}

func getValidMachineTypes() []string {
	return []string{"pc",
		"pc-i440fx-2.7",
		"pc-i440fx-2.8",
		"pc-i440fx-2.9",
		"pc-i440fx-2.10",
		"pc-i440fx-2.11",
		"pc-i440fx-2.12",
		"pc-i440fx-3.0",
		"pc-i440fx-3.1",
		"pc-i440fx-4.0",
		"pc-i440fx-4.1",
		"pc-i440fx-4.2",
		"pc-i440fx-5.0",
		"pc-i440fx-5.1",
		"pc-i440fx-5.2",
		"pc-i440fx-6.0",
		"pc-i440fx-6.1",
		"pc-i440fx-6.2",
		"pc-i440fx-7.0",
		"pc-i440fx-7.1",
		"pc-i440fx-7.2",
		"pc-i440fx-8.0",
		"pc-i440fx-8.1",
		"pc-i440fx-8.2",
		"pc-i440fx-9.0",
		"q35",
		"pc-q35-2.7",
		"pc-q35-2.8",
		"pc-q35-2.9",
		"pc-q35-2.10",
		"pc-q35-2.11",
		"pc-q35-2.12",
		"pc-q35-3.0",
		"pc-q35-3.1",
		"pc-q35-4.0",
		"pc-q35-4.1",
		"pc-q35-4.2",
		"pc-q35-5.0",
		"pc-q35-5.1",
		"pc-q35-5.2",
		"pc-q35-6.0",
		"pc-q35-6.1",
		"pc-q35-6.2",
		"pc-q35-7.0",
		"pc-q35-7.1",
		"pc-q35-7.2",
		"pc-q35-8.0",
		"pc-q35-8.1",
		"pc-q35-8.2",
		"pc-q35-9.0",
		"yottabyte",
	}
}

type VMAPIDataSourceModel struct {
	Id          int32  `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Key         int32  `json:"$key,omitempty"`
	IsSnapshot  bool   `json:"is_snapshot,omitempty"`
	CPUType     string `json:"cpu_type,omitempty"`
	MachineType string `json:"machine_type,omitempty"`
	OSFamily    string `json:"os_family,omitempty"`
	UEFI        bool   `json:"uefi,omitempty"`
	Machine     struct {
		Drives []*VMDriveAPIDataSourceModel `json:"drives,omitempty"`
		Nics   []*VMNICAPIDataSourceModel   `json:"nics,omitempty"`
	} `json:"machine,omitempty"`
}

type CloudInitFileAPI struct {
	Name     string `json:"name"`
	Contents string `json:"contents"`
}

type VMAPIGuestAgentModel struct {
	Machine struct {
		Status struct {
			AgentGuestInfo *VMAPIAgentGuestInfoModel `json:"agent_guest_info,omitempty"`
		} `json:"status,omitempty"`
	} `json:"machine,omitempty"`
}

type VMAPIAgentGuestInfoModel struct {
	Network []*VMAPIGuestAgentNetworkModel `json:"network,omitempty"`
}

type VMAPIGuestAgentNetworkModel struct {
	Name        string                        `json:"name,omitempty"`
	IPAddresses []*VMAPIGuestAgentIPAddresses `json:"ip-addresses,omitempty"`
}

type VMAPIGuestAgentIPAddresses struct {
	IPAddressType string `json:"ip-address-type,omitempty"`
	IPAddress     string `json:"ip-address,omitempty"`
}

type VMAPIResourceModel struct {
	Id                   string             `json:"id,omitempty"`
	Machine              int                `json:"machine,omitempty"`
	Name                 string             `json:"name,omitempty"`
	Cluster              string             `json:"cluster,omitempty"`
	Description          string             `json:"description,omitempty"`
	Enabled              bool               `json:"enabled,omitempty"`
	MachineType          string             `json:"machine_type,omitempty"`
	AllowHotplug         bool               `json:"allow_hotplug,omitempty"`
	DisablePowercycle    bool               `json:"disable_powercycle,omitempty"`
	CPUCores             int                `json:"cpu_cores,omitempty"`
	CPUType              string             `json:"cpu_type,omitempty"`
	RAM                  int                `json:"ram,omitempty"`
	Console              string             `json:"console,omitempty"`
	Display              string             `json:"display,omitempty"`
	Video                string             `json:"video,omitempty"`
	Sound                string             `json:"sound,omitempty"`
	OSFamily             string             `json:"os_family,omitempty"`
	OSDescription        string             `json:"os_description,omitempty"`
	RTCBase              string             `json:"rtc_base,omitempty"`
	BootOrder            string             `json:"boot_order,omitempty"`
	ConsolePassEnabled   bool               `json:"console_pass_enabled,omitempty"`
	ConsolePass          string             `json:"console_pass,omitempty"`
	USBTablet            bool               `json:"usb_tablet,omitempty"`
	UEFI                 bool               `json:"uefi,omitempty"`
	SecureBoot           bool               `json:"secure_boot,omitempty"`
	SerialPort           bool               `json:"serial_port,omitempty"`
	BootDelay            int                `json:"boot_delay,omitempty"`
	PreferredNode        string             `json:"preferred_node,omitempty"`
	SnapshotProfile      string             `json:"snapshot_profile,omitempty"`
	CloudInitDataSource  string             `json:"cloudinit_datasource,omitempty"`
	CloudInitFiles       []CloudInitFileAPI `json:"cloudinit_files,omitempty"`
	PowerState           bool               `json:"powerstate,omitempty"`
	GuestAgent           bool               `json:"guest_agent,omitempty"`
	HAGroup              string             `json:"ha_group,omitempty"`
	Advanced             string             `json:"advanced,omitempty"`
	NestedVirtualization bool               `json:"nested_virtualization"`
	DisableHypervisor    bool               `json:"disable_hypervisor"`
	VmDisks              []interface{}      `json:"vm_disks,omitempty"`
}

type VMAction struct {
	VM     int            `json:"vm,omitempty"`
	Action string         `json:"action,omitempty"`
	Params VMActionParams `json:"params,omitempty"`
}

type VMPowerState struct {
	PowerState *bool `json:"powerstate,omitempty"`
}

type VMActionParams struct {
	Device string `json:"device,omitempty"`
	Unplug bool   `json:"unplug,omitempty"`
}

type NewResponse struct {
	Key      string             `json:"$key,omitempty"`
	Response NewResponseMachine `json:"response,omitempty"`
}

type NewResponseMachine struct {
	Machine string `json:"machine,omitempty"`
}

func (va *VMApi) CreateVM(_ context.Context, apiData *VMAPIResourceModel) error {
	log.Printf("[Vergeio]: Creating VM with data: %+v", apiData)

	encodedBuffer := new(bytes.Buffer)
	if err := json.NewEncoder(encodedBuffer).Encode(apiData); err != nil {
		return errors.New("invalid format received for VM Item")
	}

	apiResp, err := va.client.Post(VMEndpoint, encodedBuffer)
	if err != nil {
		return err
	}
	if apiResp == nil {
		return errors.New("missing response from the API")
	}
	if apiResp.StatusCode != 201 {
		return fmt.Errorf("missing response from API %d", apiResp.StatusCode)
	}

	var vmAPIResp NewResponse
	if err := json.NewDecoder(apiResp.Body).Decode(&vmAPIResp); err != nil {
		return fmt.Errorf("invalid format received for VM Item: %v", err)
	}

	log.Printf("VM Key after creation %v", vmAPIResp.Response)

	apiData.Id = vmAPIResp.Key

	log.Printf("VM Id after creation %v", apiData.Id)

	if readError := va.readVM(apiData); readError != nil {
		return errors.New("Error reading the VM: " + readError.Error())
	}

	return nil
}

func (va *VMApi) DeleteVM(ctx context.Context, vmId string) error {
	log.Printf("[Vergeio]: Deleting VM with ID: %s", vmId)

	apiResp, err := va.client.Delete(fmt.Sprintf("%s/%s", VMEndpoint, url.PathEscape(vmId)))

	if err != nil {
		return fmt.Errorf("error deleting VM %s: %w", vmId, err)
	}
	if apiResp == nil {
		return fmt.Errorf("no response received when deleting VM %s", vmId)
	}
	if apiResp.StatusCode != 200 && apiResp.StatusCode != 204 {
		return fmt.Errorf("failed to delete VM %s, status code: %d", vmId, apiResp.StatusCode)
	}

	log.Printf("[Vergeio]: Successfully deleted VM with ID: %s (and all associated disks)", vmId)
	return nil
}

func (va *VMApi) IsVMRunning(ctx context.Context, vmId string) (*bool, error) {
	log.Printf("Checking power state for VM ID: %s", vmId)

	apiResp, err := va.client.Get(fmt.Sprintf("%s/%s",
		VMEndpoint,
		url.PathEscape(vmId)),
		&Options{
			Fields: "machine#status#running as powerstate"})

	if err != nil {
		return nil, err
	}
	if apiResp == nil {
		return nil, errors.New("missing response from the API")
	}
	if apiResp.StatusCode != 200 {
		return nil, fmt.Errorf("missing response from API %d", apiResp.StatusCode)
	}

	log.Printf("Power state API response status: %d", apiResp.StatusCode)

	var vmAPIResp VMPowerState
	if err := json.NewDecoder(apiResp.Body).Decode(&vmAPIResp); err != nil {
		return nil, fmt.Errorf("invalid format received for VM power state: %v", err)
	}

	log.Printf("VM power state read from API: %v", vmAPIResp.PowerState)

	return vmAPIResp.PowerState, nil
}

func (va *VMApi) PowerOnVM(vmKey string) error {
	log.Printf("Calling the Power On VM API for VM Key %s", vmKey)
	err := va.changeVMPowerState(vmKey, "poweron")
	if err != nil {
		return err
	}

	time.Sleep(10 * time.Second)

	return nil
}

func (va *VMApi) PowerOffVM(vmKey string) error {
	log.Printf("Calling the Power Off VM API for VM Key %s", vmKey)
	err := va.changeVMPowerState(vmKey, "kill")
	if err != nil {
		return err
	}

	time.Sleep(5 * time.Second)

	return nil
}

func (va *VMApi) changeVMPowerState(vmKey string, desiredState string) error {
	log.Printf("Change the power state for VM Key %s to %s", vmKey, desiredState)

	actionData := map[string]interface{}{
		"vm":     vmKey,
		"action": desiredState,
	}

	bytedata, err := json.Marshal(actionData)
	if err != nil {
		return err
	}

	req, err := va.client.Post(VMActionEndpoint, bytes.NewBuffer(bytedata))
	if err != nil {
		return err
	}
	if req.StatusCode != 201 {
		return fmt.Errorf("failed to change the VM power state: status code %v", req.StatusCode)
	}

	return nil
}

func (va *VMApi) readVM(data *VMAPIResourceModel) error {
	log.Printf("[Vergeio]: Reading the vm data")

	apiResp, err := va.client.Get(fmt.Sprintf("%s/%s",
		VMEndpoint,
		url.PathEscape(data.Id),
	), &Options{Fields: "id,machine,name,cluster,description,enabled,machine_type,allow_hotplug,disable_powercycle,cpu_cores,cpu_type,ram,console,display,video,sound,os_family,os_description,rtc_base,boot_order,console_pass_enabled,console_pass,usb_tablet,uefi,secure_boot,serial_port,boot_delay,preferred_node,snapshot_profile,cloudinit_datasource,ha_group,guest_agent,advanced,nested_virtualization,disable_hypervisor,machine#status#running as powerstate"})

	if err != nil {
		return err
	}
	if apiResp == nil {
		return errors.New("missing response from the API")
	}
	if apiResp.StatusCode != 200 {
		return fmt.Errorf("missing response from API %d", apiResp.StatusCode)
	}

	log.Printf("[Vergeio]: Read the VM %v", apiResp.Body)

	if err := json.NewDecoder(apiResp.Body).Decode(&data); err != nil {
		return fmt.Errorf("invalid format received for VM Item: %v", err)
	}

	log.Println("[Vergeio]: Data was successfully converted to a resource")

	return nil
}

func (va *VMApi) GetGuestAgentIPs(ctx context.Context, vmId string) ([]string, error) {
	log.Printf("[VergeIO]: Reading guest agent network information for VM ID: %s", vmId)

	apiResp, err := va.client.Get(fmt.Sprintf("%s/%s", VMEndpoint, vmId), &Options{
		Fields: "dashboard",
	})

	if err != nil {
		log.Printf("[VergeIO]: Error calling VergeIO API for guest agent info: %v", err)
		return nil, fmt.Errorf("failed to get guest agent info from VergeIO API: %w", err)
	}

	if apiResp == nil {
		return nil, fmt.Errorf("received nil response from VergeIO API")
	}

	if apiResp.StatusCode != 200 {
		return nil, fmt.Errorf("VergeIO API returned status code %d when requesting guest agent info", apiResp.StatusCode)
	}

	log.Printf("[VergeIO]: Successfully received guest agent response from API")

	if apiResp.Body == nil {
		log.Printf("[VergeIO]: No response body - guest agent may not be running yet")
		return nil, fmt.Errorf("no guest agent data available")
	}

	body, err := io.ReadAll(apiResp.Body)
	if err != nil {
		log.Printf("[VergeIO]: Failed to read response body: %v", err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	log.Printf("[VergeIO]: Raw guest agent response: %s", string(body))

	var gaResp VMAPIGuestAgentModel
	if err := json.Unmarshal(body, &gaResp); err != nil {
		log.Printf("[VergeIO]: Failed to decode guest agent JSON response: %v", err)
		return []string{}, nil
	}

	if gaResp.Machine.Status.AgentGuestInfo == nil {
		log.Printf("[VergeIO]: Guest agent is not reporting network information yet")
		return []string{}, nil
	}

	log.Printf("[VergeIO]: Guest agent is active and reporting network information")

	var ipAddresses []string
	guestInfo := gaResp.Machine.Status.AgentGuestInfo

	for _, network := range guestInfo.Network {
		log.Printf("[VergeIO]: Processing network interface: %s", network.Name)

		for _, ip := range network.IPAddresses {
			if ip.IPAddressType == "ipv4" {
				if ip.IPAddress != "" {
					if isLoopbackIP(ip.IPAddress) {
						log.Printf("[VergeIO]: Skipping loopback address: %s on interface %s", ip.IPAddress, network.Name)
						continue
					}
					log.Printf("[VergeIO]: Found IPv4 address: %s on interface %s", ip.IPAddress, network.Name)
					ipAddresses = append(ipAddresses, ip.IPAddress)
				}
			} else {
				log.Printf("[VergeIO]: Skipping non-IPv4 address: %s (type: %s)", ip.IPAddress, ip.IPAddressType)
			}
		}
	}

	if len(ipAddresses) > 0 {
		log.Printf("[VergeIO]: Successfully discovered %d non-loopback IPv4 address(es): %v", len(ipAddresses), ipAddresses)
	} else {
		log.Printf("[VergeIO]: No non-loopback IPv4 addresses found in guest agent data")
	}

	return ipAddresses, nil
}

func (va *VMApi) GetGuestAgentIPsWithDebug(ctx context.Context, vmId string) ([]string, string, error) {
	apiResp, err := va.client.Get(fmt.Sprintf("%s/%s", VMEndpoint, vmId), &Options{
		Fields: "dashboard",
	})

	if err != nil {
		return nil, "", fmt.Errorf("failed to get guest agent info from VergeIO API: %w", err)
	}

	if apiResp == nil {
		return nil, "", fmt.Errorf("received nil response from VergeIO API")
	}

	if apiResp.StatusCode != 200 {
		return nil, "", fmt.Errorf("VergeIO API returned status code %d when requesting guest agent info", apiResp.StatusCode)
	}

	if apiResp.Body == nil {
		return nil, "", fmt.Errorf("no guest agent data available")
	}

	body, err := io.ReadAll(apiResp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response body: %w", err)
	}
	rawResponse := string(body)

	var gaResp VMAPIGuestAgentModel
	if err := json.Unmarshal(body, &gaResp); err != nil {
		return []string{}, rawResponse, nil
	}

	if gaResp.Machine.Status.AgentGuestInfo == nil {
		return []string{}, rawResponse, nil
	}

	var ipAddresses []string
	guestInfo := gaResp.Machine.Status.AgentGuestInfo

	for _, network := range guestInfo.Network {
		for _, ip := range network.IPAddresses {
			if ip.IPAddressType == "ipv4" {
				if ip.IPAddress != "" && !isLoopbackIP(ip.IPAddress) {
					ipAddresses = append(ipAddresses, ip.IPAddress)
				}
			}
		}
	}

	return ipAddresses, rawResponse, nil
}

func (va *VMApi) WaitForGuestAgent(ctx context.Context, vmId string, timeout time.Duration) error {
	log.Printf("[VergeIO]: Waiting for guest agent to become available (timeout: %v)", timeout)

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			log.Printf("[VergeIO]: Timeout waiting for guest agent to become available")
			return fmt.Errorf("timeout waiting for guest agent (waited %v)", timeout)

		case <-ticker.C:
			log.Printf("[VergeIO]: Checking guest agent availability...")

			ips, err := va.GetGuestAgentIPs(ctx, vmId)

			if err == nil && len(ips) > 0 {
				log.Printf("[VergeIO]: Guest agent is now available and reporting IPs: %v", ips)
				return nil
			}

			if err != nil {
				log.Printf("[VergeIO]: Guest agent not yet available: %v", err)
			} else {
				log.Printf("[VergeIO]: Guest agent responding but no IPs reported yet")
			}
		}
	}
}

// VMInfo represents VM information for data source
type VMInfo struct {
	ID          int32          `json:"id,omitempty"`
	Name        string         `json:"name,omitempty"`
	Key         int32          `json:"$key,omitempty"`
	IsSnapshot  bool           `json:"is_snapshot,omitempty"`
	CPUType     string         `json:"cpu_type,omitempty"`
	MachineType string         `json:"machine_type,omitempty"`
	OSFamily    string         `json:"os_family,omitempty"`
	UEFI        bool           `json:"uefi,omitempty"`
	Drives      []*VMDriveInfo `json:"drives,omitempty"`
	Nics        []*VMNicInfo   `json:"nics,omitempty"`
}

type VMDriveInfo struct {
	Key           int32                   `json:"key,omitempty"`
	Name          string                  `json:"name,omitempty"`
	Interface     string                  `json:"interface,omitempty"`
	Media         string                  `json:"media,omitempty"`
	Description   string                  `json:"description,omitempty"`
	PreferredTier string                  `json:"preferred_tier,omitempty"`
	MediaSource   *VMDriveMediaSourceInfo `json:"media_source,omitempty"`
}

type VMDriveMediaSourceInfo struct {
	Key            int32 `json:"key,omitempty"`
	UsedBytes      int64 `json:"used_bytes,omitempty"`
	AllocatedBytes int64 `json:"allocated_bytes,omitempty"`
	Filesize       int64 `json:"filesize,omitempty"`
}

type VMNicInfo struct {
	Key        int32  `json:"key,omitempty"`
	Name       string `json:"name,omitempty"`
	Interface  string `json:"interface,omitempty"`
	Vnet       string `json:"vnet,omitempty"`
	Status     string `json:"status,omitempty"`
	Ipaddress  string `json:"ipaddress,omitempty"`
	MacAddress string `json:"macaddress,omitempty"`
}

// GetVMs queries VMs and returns matching VMs with drives and nics data
func (va *VMApi) GetVMs(ctx context.Context, filterName string, filterId int, isSnapshot bool) ([]VMInfo, error) {
	log.Printf("[VergeIO]: Querying VMs with filters - Name: %s, Id: %d, IsSnapshot: %t", filterName, filterId, isSnapshot)

	// Build filter options - use fields similar to Terraform implementation
	opts := &Options{
		Fields: "machine#$key as id, dashboard", // This matches the Terraform query
	}

	// Add filters if specified
	var filters []string
	if filterName != "" {
		filters = append(filters, fmt.Sprintf("name eq '%s'", filterName))
	}
	if filterId > 0 {
		filters = append(filters, fmt.Sprintf("id eq %d", filterId))
	}

	if len(filters) > 0 {
		opts.Filter = strings.Join(filters, " and ")
	}

	// Query the API
	apiResp, err := va.client.Get(VMEndpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to query VMs: %w", err)
	}

	if apiResp == nil {
		return nil, errors.New("missing response from VergeIO API")
	}

	if apiResp.StatusCode != 200 {
		return nil, fmt.Errorf("VergeIO API returned status code %d", apiResp.StatusCode)
	}

	// Decode the response using the data source model
	var vmAPIResp []VMAPIDataSourceModel
	if err := json.NewDecoder(apiResp.Body).Decode(&vmAPIResp); err != nil {
		return nil, fmt.Errorf("failed to decode VMs response: %w", err)
	}

	// Convert API response to VMInfo format
	var vms []VMInfo
	for _, vmAPIRespItem := range vmAPIResp {
		// Apply snapshot filter if specified
		if filterName != "" || filterId > 0 {
			// For specific name/id queries, include regardless of snapshot status for now
		} else if isSnapshot != vmAPIRespItem.IsSnapshot {
			// Skip if snapshot filter doesn't match
			continue
		}

		vm := VMInfo{
			ID:          vmAPIRespItem.Id,
			Name:        vmAPIRespItem.Name,
			Key:         vmAPIRespItem.Key,
			IsSnapshot:  vmAPIRespItem.IsSnapshot,
			CPUType:     vmAPIRespItem.CPUType,
			MachineType: vmAPIRespItem.MachineType,
			OSFamily:    vmAPIRespItem.OSFamily,
			UEFI:        vmAPIRespItem.UEFI,
		}

		// Process drives
		if vmAPIRespItem.Machine.Drives != nil {
			for _, vmDrive := range vmAPIRespItem.Machine.Drives {
				drive := &VMDriveInfo{
					Key:           int32(vmDrive.Key),
					Name:          vmDrive.Name,
					Interface:     vmDrive.Interface,
					Media:         vmDrive.Media,
					Description:   vmDrive.Description,
					PreferredTier: vmDrive.PreferredTier,
				}

				if vmDrive.MediaSource != nil {
					drive.MediaSource = &VMDriveMediaSourceInfo{
						Key:            int32(vmDrive.MediaSource.Key),
						UsedBytes:      int64(vmDrive.MediaSource.UsedBytes),
						AllocatedBytes: int64(vmDrive.MediaSource.AllocatedBytes),
						Filesize:       int64(vmDrive.MediaSource.Filesize),
					}
				}

				vm.Drives = append(vm.Drives, drive)
			}
		}

		// Process nics
		if vmAPIRespItem.Machine.Nics != nil {
			for _, vmNic := range vmAPIRespItem.Machine.Nics {
				nic := &VMNicInfo{
					Key:        int32(vmNic.Key),
					Name:       vmNic.Name,
					Interface:  vmNic.Interface,
					Vnet:       vmNic.Vnet,
					Status:     vmNic.Status,
					Ipaddress:  vmNic.Ipaddress,
					MacAddress: vmNic.MacAddress,
				}
				vm.Nics = append(vm.Nics, nic)
			}
		}

		vms = append(vms, vm)
	}

	log.Printf("[VergeIO]: Found %d VM(s) matching the criteria", len(vms))
	return vms, nil
}

// isLoopbackIP checks if an IP address is a loopback address
// This includes 127.0.0.1, ::1, and any address in the 127.0.0.0/8 range
func isLoopbackIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	return ip.IsLoopback()
}
