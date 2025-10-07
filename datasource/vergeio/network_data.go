// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc mapstructure-to-hcl2 -type NetworkConfig,NetworkOutput
package vergeio

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	client "github.com/vergeio/packer-plugin-vergeio/client"
	"github.com/zclconf/go-cty/cty"
)

type NetworkConfig struct {
	// VergeIO connection configuration (reusing the cluster config pattern)
	Username string `mapstructure:"vergeio_username" required:"true"`
	Password string `mapstructure:"vergeio_password" required:"true"`
	Endpoint string `mapstructure:"vergeio_endpoint" required:"true"`
	Port     int    `mapstructure:"vergeio_port" required:"false"`
	Insecure bool   `mapstructure:"vergeio_insecure" required:"false"`

	// Filter options for network query
	FilterName string `mapstructure:"filter_name" required:"false"`
	FilterType string `mapstructure:"filter_type" required:"false"`
}

type NetworkDataSource struct {
	config NetworkConfig
}

type NetworkInfo struct {
	ID          int32  `mapstructure:"id"`
	Name        string `mapstructure:"name"`
	Description string `mapstructure:"description"`
}

type NetworkOutput struct {
	Networks []NetworkInfo `mapstructure:"networks"`
}

func (d *NetworkDataSource) ConfigSpec() hcldec.ObjectSpec {
	return d.config.FlatMapstructure().HCL2Spec()
}

func (d *NetworkDataSource) Configure(raws ...interface{}) error {
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

	log.Printf("[VergeIO Network DataSource]: Configured to connect to %s with user %s",
		d.config.Endpoint, d.config.Username)
	log.Printf("[VergeIO Network DataSource]: Filter settings - name='%s', type='%s'",
		d.config.FilterName, d.config.FilterType)

	return nil
}

func (d *NetworkDataSource) OutputSpec() hcldec.ObjectSpec {
	return (&NetworkOutput{}).FlatMapstructure().HCL2Spec()
}

func (d *NetworkDataSource) Execute() (cty.Value, error) {
	log.Printf("[VergeIO Network DataSource]: Starting network data source execution")

	// Create VergeIO client using the configured credentials
	vergeClient := client.NewClient(d.config.Endpoint, d.config.Username, d.config.Password, d.config.Insecure)
	networkAPI := client.NewNetworkApi(vergeClient)

	// Query networks from VergeIO API using the real API
	networks, err := networkAPI.GetNetworks(context.Background(), d.config.FilterName, d.config.FilterType)
	if err != nil {
		return cty.NilVal, fmt.Errorf("failed to get networks from VergeIO API: %w", err)
	}

	log.Printf("[VergeIO Network DataSource]: Found %d networks from VergeIO API", len(networks))

	// Convert to output format
	var networkInfos []NetworkInfo
	for _, network := range networks {
		networkInfos = append(networkInfos, NetworkInfo{
			ID:          network.ID,
			Name:        network.Name,
			Description: network.Description,
		})
		log.Printf("[VergeIO Network DataSource]: Network - ID: %d, Name: %s, Description: %s",
			network.ID, network.Name, network.Description)
	}

	output := NetworkOutput{
		Networks: networkInfos,
	}

	log.Printf("[VergeIO Network DataSource]: Successfully processed %d networks from VergeIO", len(networkInfos))
	return hcl2helper.HCL2ValueFromConfig(output, d.OutputSpec()), nil
}
