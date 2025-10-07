package vergeio

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

// Network endpoints based on Terraform provider
const (
	NetworkEndpoint = APIEndpoint + "/vnets"
)

// NetworkApi provides methods for interacting with VergeIO network (vnet) resources
type NetworkApi struct {
	client *Client
}

// NetworkInfo represents network information returned by the data source
type NetworkInfo struct {
	ID          int32  `json:"$key"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// NewNetworkApi creates a new NetworkApi instance
func NewNetworkApi(client *Client) *NetworkApi {
	return &NetworkApi{
		client: client,
	}
}

// GetNetworks retrieves networks from VergeIO API with optional filtering
func (na *NetworkApi) GetNetworks(ctx context.Context, filterName, filterType string) ([]NetworkInfo, error) {
	log.Printf("[VergeIO Network API]: Getting networks with filter_name='%s', filter_type='%s'", filterName, filterType)

	// Build query options
	opts := &Options{
		Fields: "description,name,$key", // Request ID, name, and description fields
	}

	// Build name filter if specified
	if filterName != "" {
		opts.Filter = fmt.Sprintf("name eq '%s'", filterName)
		log.Printf("[VergeIO Network API]: Added name filter: %s", opts.Filter)
	}

	// Build type filter if specified
	if filterType != "" {
		if opts.Filter != "" {
			opts.Filter = fmt.Sprintf("%s and type eq '%s'", opts.Filter, filterType)
		} else {
			opts.Filter = fmt.Sprintf("type eq '%s'", filterType)
		}
		log.Printf("[VergeIO Network API]: Added type filter: %s", opts.Filter)
	}

	// Call the VergeIO API
	log.Printf("[VergeIO Network API]: Making API call to %s", NetworkEndpoint)
	apiResp, err := na.client.Get(NetworkEndpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to call VergeIO API: %w", err)
	}

	if apiResp == nil {
		return nil, errors.New("missing response from the VergeIO API")
	}

	if apiResp.StatusCode != 200 {
		return nil, fmt.Errorf("VergeIO API returned status code %d", apiResp.StatusCode)
	}

	log.Printf("[VergeIO Network API]: Received successful response from API")

	// Decode the API response
	var networks []NetworkInfo
	if err := json.NewDecoder(apiResp.Body).Decode(&networks); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	log.Printf("[VergeIO Network API]: Successfully decoded %d networks from API response", len(networks))

	// Log each network found
	for _, network := range networks {
		log.Printf("[VergeIO Network API]: Network found - ID: %d, Name: %s, Description: %s", 
			network.ID, network.Name, network.Description)
	}

	return networks, nil
}

// GetNetworkByName retrieves a specific network by name
func (na *NetworkApi) GetNetworkByName(ctx context.Context, name string) (*NetworkInfo, error) {
	log.Printf("[VergeIO Network API]: Getting network by name: %s", name)

	networks, err := na.GetNetworks(ctx, name, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get network by name '%s': %w", name, err)
	}

	if len(networks) == 0 {
		return nil, fmt.Errorf("network with name '%s' not found", name)
	}

	if len(networks) > 1 {
		log.Printf("[VergeIO Network API]: Warning - found %d networks with name '%s', returning first one", len(networks), name)
	}

	network := &networks[0]
	log.Printf("[VergeIO Network API]: Found network - ID: %d, Name: %s", network.ID, network.Name)
	return network, nil
}