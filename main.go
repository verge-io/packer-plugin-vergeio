// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"log"
	"os"

	vergeio "github.com/vergeio/packer-plugin-vergeio/builder/vergeio"
	vergeioData "github.com/vergeio/packer-plugin-vergeio/datasource/vergeio"
	vergeioPP "github.com/vergeio/packer-plugin-vergeio/post-processor/vergeio"
	vergeioProv "github.com/vergeio/packer-plugin-vergeio/provisioner/vergeio"
	vergeioVersion "github.com/vergeio/packer-plugin-vergeio/version"

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
	pps.SetVersion(vergeioVersion.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
