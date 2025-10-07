Include a short description about the data source. This is a good place
to call out what the data source does, and any requirements for the given
data source environment. See https://www.packer.io/docs/data-source/amazon-ami
-->

The vergeio data source is used to create endless Packer plugins using
a consistent plugin structure.

<!-- Data source Configuration Fields -->

**Required**

- `mock` (string) - The name of the mock to use for the VergeIO API.

<!--
  Optional Configuration Fields

  Configuration options that are not required or have reasonable defaults
  should be listed under the optionals section. Defaults values should be
  noted in the description of the field
-->

**Optional**

- `mock_api_url` (string) - The VergeIO API endpoint to connect to.
  Defaults to https://example.com

<!--
  A basic example on the usage of the data source. Multiple examples
  can be provided to highlight various build configurations.

-->

### OutPut

- `foo` (string) - The VergeIO output foo value.
- `bar` (string) - The VergeIO output bar value.

<!--
  A basic example on the usage of the data source. Multiple examples
  can be provided to highlight various build configurations.

-->

### Example Usage

```hcl
data "vergeio" "example" {
   mock = "bird"
 }
 source "vergeio" "example" {
   mock = data.vergeio.example.foo
 }

 build {
   sources = ["source.vergeio.example"]
 }
```
