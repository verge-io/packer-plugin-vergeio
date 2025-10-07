// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MIT

package vergeio

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

const (
	DiskEndpoint = APIEndpoint + "/machine_drives"
)

func NewDriveApi(c *Client) *DriveApi {
	return &DriveApi{
		name:   "Drive Api",
		client: c,
	}
}

type DriveApi struct {
	name   string
	client *Client
}

func (da *DriveApi) Name() string {
	return da.name
}

type VMDriveAPIDataSourceModel struct {
	Key           int                                `json:"$key,omitempty"`
	Name          string                             `json:"name,omitempty"`
	Interface     string                             `json:"interface,omitempty"`
	Media         string                             `json:"media,omitempty"`
	Description   string                             `json:"description,omitempty"`
	PreferredTier string                             `json:"preferred_tier,omitempty"`
	MediaSource   *VMDriveMediaSourceDataSourceModel `json:"media_source,omitempty"`
}

type VMDiskResourceModel struct {
	Key                 int    `json:"$key,omitempty"`
	Machine             int    `json:"machine,omitempty"`
	Name                string `json:"name,omitempty"`
	Description         string `json:"description,omitempty"`
	Interface           string `json:"interface,omitempty"`
	Media               string `json:"media,omitempty"`
	MediaSource         int    `json:"media_source,omitempty"`
	PreferredTier       string `json:"preferred_tier,omitempty"`
	DiskSize            int64  `json:"disksize,omitempty"`
	Enabled             bool   `json:"enabled,omitempty"`
	ReadOnly            bool   `json:"readonly,omitempty"`
	Serial              string `json:"serial,omitempty"`
	Asset               string `json:"asset,omitempty"`
	OrderId             int    `json:"orderid,omitempty"`
	PreserveDriveFormat bool   `json:"preserve_drive_format,omitempty"`
}

type VMDriveMediaSourceDataSourceModel struct {
	Key            int `json:"$key,omitempty"`
	UsedBytes      int `json:"used_bytes,omitempty"`
	AllocatedBytes int `json:"allocated_bytes,omitempty"`
	Filesize       int `json:"filesize,omitempty"`
}

type diskResponse struct {
	Key      string `json:"$key,omitempty"`
	Response string `json:"response,omitempty"`
	Error    string `json:"err,omitempty"`
}

// DiskPowerStatus represents the power/import status of a disk
type DiskPowerStatus struct {
	PowerState string `json:"powerstate,omitempty"`
}

func (da *DriveApi) CreateVMDisk(ctx context.Context, apiData *VMDiskResourceModel) error {
	// Encode the API data
	encodedBuffer := new(bytes.Buffer)
	if err := json.NewEncoder(encodedBuffer).Encode(apiData); err != nil {
		return errors.New("invalid format received for disk Item")
	}

	// Call the API and check the response
	apiResp, err := da.client.Post(DiskEndpoint, encodedBuffer)
	if err != nil {
		return err
	}
	if apiResp == nil {
		return errors.New("missing response from the API")
	}
	if apiResp.StatusCode != 201 {
		return fmt.Errorf("missing response from the API %d", apiResp.StatusCode)
	}

	// Decode the API response
	var diskAPIResp diskResponse
	if err := json.NewDecoder(apiResp.Body).Decode(&diskAPIResp); err != nil {
		return fmt.Errorf("invalid format received for creating a disk %v", err)
	}

	log.Printf("Created a disk with Id %v", diskAPIResp.Key)

	return nil
}

// CreateVMDiskWithKey creates a VM disk and returns the disk key for tracking
func (da *DriveApi) CreateVMDiskWithKey(ctx context.Context, apiData *VMDiskResourceModel) (string, error) {
	// Encode the API data
	encodedBuffer := new(bytes.Buffer)
	if err := json.NewEncoder(encodedBuffer).Encode(apiData); err != nil {
		return "", errors.New("invalid format received for disk Item")
	}

	// Call the API and check the response
	apiResp, err := da.client.Post(DiskEndpoint, encodedBuffer)
	if err != nil {
		return "", err
	}
	if apiResp == nil {
		return "", errors.New("missing response from the API")
	}
	if apiResp.StatusCode != 201 {
		return "", fmt.Errorf("missing response from the API %d", apiResp.StatusCode)
	}

	// Decode the API response
	var diskAPIResp diskResponse
	if err := json.NewDecoder(apiResp.Body).Decode(&diskAPIResp); err != nil {
		return "", fmt.Errorf("invalid format received for creating a disk %v", err)
	}

	log.Printf("Created a disk with Id %v", diskAPIResp.Key)
	return diskAPIResp.Key, nil
}

// CheckDiskImportStatus checks the import status of a disk by its key
func (da *DriveApi) CheckDiskImportStatus(ctx context.Context, diskKey string) (string, error) {
	log.Printf("[VergeIO]: Checking import status for disk key: %s", diskKey)

	// Call the disk endpoint to get the status
	apiResp, err := da.client.Get(fmt.Sprintf("%s/%s", DiskEndpoint, diskKey), &Options{
		Fields: "status#status as powerState",
	})

	if err != nil {
		return "", fmt.Errorf("failed to get disk status: %w", err)
	}

	if apiResp == nil {
		return "", errors.New("missing response from VergeIO API")
	}

	if apiResp.StatusCode != 200 {
		return "", fmt.Errorf("VergeIO API returned status code %d", apiResp.StatusCode)
	}

	// Decode the response
	var diskStatus DiskPowerStatus
	if err := json.NewDecoder(apiResp.Body).Decode(&diskStatus); err != nil {
		return "", fmt.Errorf("failed to decode disk status response: %w", err)
	}

	log.Printf("[VergeIO]: Disk %s status: %s", diskKey, diskStatus.PowerState)
	return diskStatus.PowerState, nil
}

// WaitForDiskImportCompletion waits for all disks with media="import" to complete importing
func (da *DriveApi) WaitForDiskImportCompletion(ctx context.Context, diskKeys []string, maxRetries int) error {
	if len(diskKeys) == 0 {
		log.Printf("[VergeIO]: No disks to wait for import completion")
		return nil
	}

	log.Printf("[VergeIO]: Waiting for import completion of %d disk(s)", len(diskKeys))

	// Initial delay to allow API to process the import request
	time.Sleep(5 * time.Second)

	for _, diskKey := range diskKeys {
		log.Printf("[VergeIO]: Checking import status for disk: %s", diskKey)
		
		retries := 0
		for retries < maxRetries {
			status, err := da.CheckDiskImportStatus(ctx, diskKey)
			if err != nil {
				return fmt.Errorf("failed to check disk import status: %w", err)
			}

			log.Printf("[VergeIO]: Disk %s import status: %s (attempt %d/%d)", diskKey, status, retries+1, maxRetries)

			// Check if import is complete (status is not "importing")
			if strings.ToLower(status) != "importing" {
				log.Printf("[VergeIO]: Disk %s import completed with status: %s", diskKey, status)
				break
			}

			// Still importing, wait and retry
			retries++
			if retries >= maxRetries {
				return fmt.Errorf("disk %s failed to complete import after %d retries, last status: %s", diskKey, maxRetries, status)
			}

			log.Printf("[VergeIO]: Disk %s still importing, waiting 5 seconds before retry %d/%d", diskKey, retries+1, maxRetries)
			time.Sleep(5 * time.Second)
		}
	}

	log.Printf("[VergeIO]: All disk imports completed successfully")
	return nil
}

// ReadDisk reads disk information from the API to get current size and status
func (da *DriveApi) ReadDisk(ctx context.Context, diskKey string) (*VMDiskResourceModel, error) {
	log.Printf("[VergeIO]: Reading disk information for key: %s", diskKey)

	// Call the disk endpoint to get the disk information
	apiResp, err := da.client.Get(fmt.Sprintf("%s/%s", DiskEndpoint, diskKey), &Options{
		Fields: "machine,name,disksize,interface,media,description,enabled,serial,media_source,preferred_tier,readonly,preserve_drive_format,asset,orderid",
	})

	if err != nil {
		return nil, fmt.Errorf("failed to read disk: %w", err)
	}

	if apiResp == nil {
		return nil, errors.New("missing response from VergeIO API")
	}

	if apiResp.StatusCode != 200 {
		return nil, fmt.Errorf("VergeIO API returned status code %d", apiResp.StatusCode)
	}

	// Decode the response
	var diskData VMDiskResourceModel
	if err := json.NewDecoder(apiResp.Body).Decode(&diskData); err != nil {
		return nil, fmt.Errorf("failed to decode disk response: %w", err)
	}

	log.Printf("[VergeIO]: Successfully read disk %s, current size: %d bytes", diskKey, diskData.DiskSize)
	return &diskData, nil
}

// UpdateDiskSize updates the disk size when import disk size differs from requested size
func (da *DriveApi) UpdateDiskSize(ctx context.Context, diskKey string, requestedSizeGB int64) error {
	log.Printf("[VergeIO]: Updating disk size for key %s to %d GB", diskKey, requestedSizeGB)

	// Prepare the API data packet with only the size field to update
	updateData := map[string]interface{}{
		"disksize": requestedSizeGB * 1024 * 1024 * 1024, // Convert GB to bytes
	}

	// Encode the API data
	encodedBuffer := new(bytes.Buffer)
	if err := json.NewEncoder(encodedBuffer).Encode(updateData); err != nil {
		return fmt.Errorf("failed to encode disk update data: %w", err)
	}

	// Call the API PUT endpoint to update the disk
	apiResp, err := da.client.Put(fmt.Sprintf("%s/%s", DiskEndpoint, diskKey), encodedBuffer)
	if err != nil {
		return fmt.Errorf("failed to update disk size: %w", err)
	}

	if apiResp == nil {
		return errors.New("missing response from VergeIO API")
	}

	if apiResp.StatusCode != 200 {
		return fmt.Errorf("VergeIO API returned status code %d for disk update", apiResp.StatusCode)
	}

	log.Printf("[VergeIO]: Successfully updated disk %s size to %d GB", diskKey, requestedSizeGB)
	return nil
}

// CheckAndResizeImportedDisks checks if imported disks need to be resized and updates them
func (da *DriveApi) CheckAndResizeImportedDisks(ctx context.Context, diskConfigs []VMDiskResourceModel, diskKeys []string) error {
	if len(diskConfigs) == 0 || len(diskKeys) == 0 {
		log.Printf("[VergeIO]: No disks to check for resizing")
		return nil
	}

	log.Printf("[VergeIO]: Checking %d imported disk(s) for size mismatches", len(diskKeys))

	// Create a map of disk names to requested sizes for quick lookup
	diskSizeMap := make(map[string]int64)
	for _, config := range diskConfigs {
		if config.Media == "import" {
			diskSizeMap[config.Name] = config.DiskSize
		}
	}

	// Check each imported disk
	for i, diskKey := range diskKeys {
		if i >= len(diskConfigs) {
			continue
		}

		config := diskConfigs[i]
		requestedSizeGB := config.DiskSize

		log.Printf("[VergeIO]: Checking disk '%s' (key: %s) - requested size: %d GB", config.Name, diskKey, requestedSizeGB)

		// Read current disk information
		currentDisk, err := da.ReadDisk(ctx, diskKey)
		if err != nil {
			return fmt.Errorf("failed to read disk %s after import: %w", config.Name, err)
		}

		// Convert current size from bytes to GB for comparison
		currentSizeGB := currentDisk.DiskSize / (1024 * 1024 * 1024)
		
		log.Printf("[VergeIO]: Disk '%s' current size: %d GB, requested size: %d GB", config.Name, currentSizeGB, requestedSizeGB)

		// Check if sizes match
		if currentSizeGB != requestedSizeGB {
			log.Printf("[VergeIO]: Size mismatch detected for disk '%s' - resizing from %d GB to %d GB", 
				config.Name, currentSizeGB, requestedSizeGB)

			// Update the disk size
			if err := da.UpdateDiskSize(ctx, diskKey, requestedSizeGB); err != nil {
				return fmt.Errorf("failed to resize disk %s: %w", config.Name, err)
			}

			log.Printf("[VergeIO]: Successfully resized disk '%s' to %d GB", config.Name, requestedSizeGB)
		} else {
			log.Printf("[VergeIO]: Disk '%s' size matches requested size, no resize needed", config.Name)
		}
	}

	log.Printf("[VergeIO]: Completed disk size checking and resizing")
	return nil
}