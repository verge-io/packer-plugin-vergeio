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
)

const (
	NICEndpoint = APIEndpoint + "/machine_nics"
)

func NewNicApi(c *Client) *NicApi {
	return &NicApi{
		name:   "NIC Api",
		client: c,
	}
}

type NicApi struct {
	name   string
	client *Client
}

func (na *NicApi) Name() string {
	return na.name
}

type VMNICAPIDataSourceModel struct {
	Key        int    `json:"$key,omitempty"`
	Name       string `json:"name,omitempty"`
	Interface  string `json:"interface,omitempty"`
	Vnet       string `json:"vnet,omitempty"`
	Status     string `json:"status,omitempty"`
	Ipaddress  string `json:"ipaddress,omitempty"`
	MacAddress string `json:"macaddress,omitempty"`
	// ExternalIP string `json:"external_ip,omitempty"`
}

type VMNicResourceModel struct {
	Key             int    `json:"$key,omitempty"`
	Machine         int    `json:"machine,omitempty"`
	Name            string `json:"name,omitempty"`
	Description     string `json:"description,omitempty"`
	Interface       string `json:"interface,omitempty"`
	Driver          string `json:"driver,omitempty"`
	Model           string `json:"model,omitempty"`
	VNET            int    `json:"vnet,omitempty"`
	MAC             string `json:"macaddress,omitempty"`
	IPAddress       string `json:"ipaddress,omitempty"`
	AssignIPAddress bool   `json:"assign_ipaddress,omitempty"`
	Enabled         bool   `json:"enabled,omitempty"`
}

type nicResponse struct {
	Key      string `json:"$key,omitempty"`
	Response string `json:"response,omitempty"`
	Error    string `json:"err,omitempty"`
}

func (na *NicApi) CreateVMNic(ctx context.Context, apiData *VMNicResourceModel) error {
	// Encode the API data
	encodedBuffer := new(bytes.Buffer)
	if err := json.NewEncoder(encodedBuffer).Encode(apiData); err != nil {
		return errors.New("invalid format received for NIC Item")
	}

	// Call the API and check the response
	apiResp, err := na.client.Post(NICEndpoint, encodedBuffer)
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
	var nicAPIResp nicResponse
	if err := json.NewDecoder(apiResp.Body).Decode(&nicAPIResp); err != nil {
		return fmt.Errorf("invalid format received for creating a NIC %v", err)
	}

	log.Printf("Created a NIC with Id %v", nicAPIResp.Key)

	return nil
}