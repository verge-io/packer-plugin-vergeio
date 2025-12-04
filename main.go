// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"log"
	"os"

	vergeio "github.com/verge-io/packer-plugin-vergeio/builder/vergeio"
	vergeioData "github.com/verge-io/packer-plugin-vergeio/datasource/vergeio"
	vergeioPP "github.com/verge-io/packer-plugin-vergeio/post-processor/vergeio"
	vergeioProv "github.com/verge-io/packer-plugin-vergeio/provisioner/vergeio"
	vergeioVersion "github.com/verge-io/packer-plugin-vergeio/version"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
)

func main() {
	log.Printf("Registering Vergeio plugin ...")

	// Initialize the plugin set and register the builder, provisioner, post-processor, and datasource.
	pps := plugin.NewSet()
	pps.RegisterBuilder(plugin.DEFAULT_NAME, new(vergeio.Builder))
	// pps.RegisterBuilder("farooq-builder", new(vergeio.Builder))
	pps.RegisterProvisioner("my-provisioner", new(vergeioProv.Provisioner))
	pps.RegisterPostProcessor("my-post-processor", new(vergeioPP.PostProcessor))
	pps.RegisterDatasource("my-datasource", new(vergeioData.Datasource))
	pps.RegisterDatasource("networks", new(vergeioData.NetworkDataSource))
	pps.RegisterDatasource("vms", new(vergeioData.VMDataSource))
	pps.SetVersion(vergeioVersion.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
