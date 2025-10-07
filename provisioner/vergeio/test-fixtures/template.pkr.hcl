# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

source "null" "basic-example" {
  communicator = "none"
}

build {
  sources = [
    "source.null.basic-example"
  ]

  provisioner "vergeio-my-provisioner" {
    mock = "my-mock-config"
  }
}
