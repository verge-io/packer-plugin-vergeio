# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# This metadata.hcl file describes the VergeIO Packer plugin
# This file and the adjacent `components` docs directory should
# be kept in a `.web-docs` directory at the root of your plugin repository.
integration {
  name = "VergeIO"
  description = "A comprehensive Packer plugin for creating and provisioning virtual machines on VergeIO virtualization platform"
  identifier = "packer/vergeio/vergeio"
  flags = [
    # This plugin conforms to the HCP Packer requirements.
    #
    # Please refer to our docs if you want your plugin to be compatible with
    # HCP Packer: https://developer.hashicorp.com/packer/docs/plugins/creation/hcp-support
    "hcp-ready",
  ]
  docs {
    # Publish docs on HashiCorp websites
    process_docs = true
    # Note that the README location is relative to this file
    readme_location = "./README.md"
    # Link back to the plugin repository
    external_url = "https://github.com/verge-io/packer-plugin-vergeio"
  }
  license {
    type = "MPL-2.0"
    url = "https://github.com/verge-io/packer-plugin-vergeio/blob/main/LICENSE"
  }
  component {
    type = "builder"
    name = "VergeIO Builder"
    slug = "builder"
  }
  component {
    type = "provisioner"
    name = "VergeIO Provisioner"
    slug = "provisioner"
  }
  component {
    type = "post-processor"
    name = "VergeIO Post-processor"
    slug = "post-processor"
  }
  component {
    type = "data-source"
    name = "VergeIO Networks"
    slug = "networks"
  }
  component {
    type = "data-source"
    name = "VergeIO VMs"
    slug = "vms"
  }
}
