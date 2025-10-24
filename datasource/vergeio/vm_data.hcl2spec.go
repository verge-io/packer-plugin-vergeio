// Code generated manually for VM data source; DO NOT EDIT.

package vergeio

import (
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

// FlatVMConfig is an auto-generated flat version of VMConfig.
// Where the contents of a field with a `mapstructure:,squash` tag are bubbled up.
type FlatVMConfig struct {
	Username     *string `mapstructure:"vergeio_username" required:"true" cty:"vergeio_username" hcl:"vergeio_username"`
	Password     *string `mapstructure:"vergeio_password" required:"true" cty:"vergeio_password" hcl:"vergeio_password"`
	Endpoint     *string `mapstructure:"vergeio_endpoint" required:"true" cty:"vergeio_endpoint" hcl:"vergeio_endpoint"`
	Port         *int    `mapstructure:"vergeio_port" required:"false" cty:"vergeio_port" hcl:"vergeio_port"`
	Insecure     *bool   `mapstructure:"vergeio_insecure" required:"false" cty:"vergeio_insecure" hcl:"vergeio_insecure"`
	FilterName   *string `mapstructure:"filter_name" required:"false" cty:"filter_name" hcl:"filter_name"`
	FilterId     *int    `mapstructure:"filter_id" required:"false" cty:"filter_id" hcl:"filter_id"`
	IsSnapshot   *bool   `mapstructure:"is_snapshot" required:"false" cty:"is_snapshot" hcl:"is_snapshot"`
}

// FlatMapstructure returns a new FlatVMConfig.
// FlatVMConfig is an auto-generated flat version of VMConfig.
// Where the contents a fields with a `mapstructure:,squash` tag are bubbled up.
func (*VMConfig) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatVMConfig)
}

// HCL2Spec returns the hcl spec of a VMConfig.
// This spec is used by HCL to read the fields of VMConfig.
// The decoded values from this spec will then be applied to a FlatVMConfig.
func (*FlatVMConfig) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"vergeio_username": &hcldec.AttrSpec{Name: "vergeio_username", Type: cty.String, Required: true},
		"vergeio_password": &hcldec.AttrSpec{Name: "vergeio_password", Type: cty.String, Required: true},
		"vergeio_endpoint": &hcldec.AttrSpec{Name: "vergeio_endpoint", Type: cty.String, Required: true},
		"vergeio_port":     &hcldec.AttrSpec{Name: "vergeio_port", Type: cty.Number, Required: false},
		"vergeio_insecure": &hcldec.AttrSpec{Name: "vergeio_insecure", Type: cty.Bool, Required: false},
		"filter_name":      &hcldec.AttrSpec{Name: "filter_name", Type: cty.String, Required: false},
		"filter_id":        &hcldec.AttrSpec{Name: "filter_id", Type: cty.Number, Required: false},
		"is_snapshot":      &hcldec.AttrSpec{Name: "is_snapshot", Type: cty.Bool, Required: false},
	}
	return s
}

// FlatVMInfo is an auto-generated flat version of VMInfo.
// Where the contents of a field with a `mapstructure:,squash` tag are bubbled up.
type FlatVMInfo struct {
	ID          *int32            `mapstructure:"id" cty:"id" hcl:"id"`
	Name        *string           `mapstructure:"name" cty:"name" hcl:"name"`
	Key         *int32            `mapstructure:"key" cty:"key" hcl:"key"`
	IsSnapshot  *bool             `mapstructure:"is_snapshot" cty:"is_snapshot" hcl:"is_snapshot"`
	CPUType     *string           `mapstructure:"cpu_type" cty:"cpu_type" hcl:"cpu_type"`
	MachineType *string           `mapstructure:"machine_type" cty:"machine_type" hcl:"machine_type"`
	OSFamily    *string           `mapstructure:"os_family" cty:"os_family" hcl:"os_family"`
	UEFI        *bool             `mapstructure:"uefi" cty:"uefi" hcl:"uefi"`
	Drives      []FlatVMDriveInfo `mapstructure:"drives" cty:"drives" hcl:"drives"`
	Nics        []FlatVMNicInfo   `mapstructure:"nics" cty:"nics" hcl:"nics"`
}

// FlatVMDriveInfo is an auto-generated flat version of VMDriveInfo.
type FlatVMDriveInfo struct {
	Key           *int32                        `mapstructure:"key" cty:"key" hcl:"key"`
	Name          *string                       `mapstructure:"name" cty:"name" hcl:"name"`
	Interface     *string                       `mapstructure:"interface" cty:"interface" hcl:"interface"`
	Media         *string                       `mapstructure:"media" cty:"media" hcl:"media"`
	Description   *string                       `mapstructure:"description" cty:"description" hcl:"description"`
	PreferredTier *string                       `mapstructure:"preferred_tier" cty:"preferred_tier" hcl:"preferred_tier"`
	MediaSource   *FlatVMDriveMediaSourceInfo   `mapstructure:"media_source" cty:"media_source" hcl:"media_source"`
}

// FlatVMDriveMediaSourceInfo is an auto-generated flat version of VMDriveMediaSourceInfo.
type FlatVMDriveMediaSourceInfo struct {
	Key            *int32 `mapstructure:"key" cty:"key" hcl:"key"`
	UsedBytes      *int64 `mapstructure:"used_bytes" cty:"used_bytes" hcl:"used_bytes"`
	AllocatedBytes *int64 `mapstructure:"allocated_bytes" cty:"allocated_bytes" hcl:"allocated_bytes"`
	Filesize       *int64 `mapstructure:"filesize" cty:"filesize" hcl:"filesize"`
}

// FlatVMNicInfo is an auto-generated flat version of VMNicInfo.
type FlatVMNicInfo struct {
	Key        *int32  `mapstructure:"key" cty:"key" hcl:"key"`
	Name       *string `mapstructure:"name" cty:"name" hcl:"name"`
	Interface  *string `mapstructure:"interface" cty:"interface" hcl:"interface"`
	Vnet       *string `mapstructure:"vnet" cty:"vnet" hcl:"vnet"`
	Status     *string `mapstructure:"status" cty:"status" hcl:"status"`
	Ipaddress  *string `mapstructure:"ipaddress" cty:"ipaddress" hcl:"ipaddress"`
	MacAddress *string `mapstructure:"macaddress" cty:"macaddress" hcl:"macaddress"`
}

// FlatMapstructure returns a new FlatVMInfo.
// FlatVMInfo is an auto-generated flat version of VMInfo.
// Where the contents a fields with a `mapstructure:,squash` tag are bubbled up.
func (*VMInfo) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatVMInfo)
}

// FlatMapstructure returns a new FlatVMDriveInfo.
func (*VMDriveInfo) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatVMDriveInfo)
}

// FlatMapstructure returns a new FlatVMDriveMediaSourceInfo.
func (*VMDriveMediaSourceInfo) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatVMDriveMediaSourceInfo)
}

// FlatMapstructure returns a new FlatVMNicInfo.
func (*VMNicInfo) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatVMNicInfo)
}

// HCL2Spec returns the hcl spec of a VMInfo.
// This spec is used by HCL to read the fields of VMInfo.
// The decoded values from this spec will then be applied to a FlatVMInfo.
func (*FlatVMInfo) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"id":           &hcldec.AttrSpec{Name: "id", Type: cty.Number, Required: false},
		"name":         &hcldec.AttrSpec{Name: "name", Type: cty.String, Required: false},
		"key":          &hcldec.AttrSpec{Name: "key", Type: cty.Number, Required: false},
		"is_snapshot":  &hcldec.AttrSpec{Name: "is_snapshot", Type: cty.Bool, Required: false},
		"cpu_type":     &hcldec.AttrSpec{Name: "cpu_type", Type: cty.String, Required: false},
		"machine_type": &hcldec.AttrSpec{Name: "machine_type", Type: cty.String, Required: false},
		"os_family":    &hcldec.AttrSpec{Name: "os_family", Type: cty.String, Required: false},
		"uefi":         &hcldec.AttrSpec{Name: "uefi", Type: cty.Bool, Required: false},
		"drives":       &hcldec.BlockListSpec{TypeName: "drives", Nested: hcldec.ObjectSpec((*FlatVMDriveInfo)(nil).HCL2Spec())},
		"nics":         &hcldec.BlockListSpec{TypeName: "nics", Nested: hcldec.ObjectSpec((*FlatVMNicInfo)(nil).HCL2Spec())},
	}
	return s
}

// HCL2Spec returns the hcl spec of a VMDriveInfo.
func (*FlatVMDriveInfo) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"key":            &hcldec.AttrSpec{Name: "key", Type: cty.Number, Required: false},
		"name":           &hcldec.AttrSpec{Name: "name", Type: cty.String, Required: false},
		"interface":      &hcldec.AttrSpec{Name: "interface", Type: cty.String, Required: false},
		"media":          &hcldec.AttrSpec{Name: "media", Type: cty.String, Required: false},
		"description":    &hcldec.AttrSpec{Name: "description", Type: cty.String, Required: false},
		"preferred_tier": &hcldec.AttrSpec{Name: "preferred_tier", Type: cty.String, Required: false},
		"media_source":   &hcldec.BlockSpec{TypeName: "media_source", Nested: hcldec.ObjectSpec((*FlatVMDriveMediaSourceInfo)(nil).HCL2Spec())},
	}
	return s
}

// HCL2Spec returns the hcl spec of a VMDriveMediaSourceInfo.
func (*FlatVMDriveMediaSourceInfo) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"key":             &hcldec.AttrSpec{Name: "key", Type: cty.Number, Required: false},
		"used_bytes":      &hcldec.AttrSpec{Name: "used_bytes", Type: cty.Number, Required: false},
		"allocated_bytes": &hcldec.AttrSpec{Name: "allocated_bytes", Type: cty.Number, Required: false},
		"filesize":        &hcldec.AttrSpec{Name: "filesize", Type: cty.Number, Required: false},
	}
	return s
}

// HCL2Spec returns the hcl spec of a VMNicInfo.
func (*FlatVMNicInfo) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"key":        &hcldec.AttrSpec{Name: "key", Type: cty.Number, Required: false},
		"name":       &hcldec.AttrSpec{Name: "name", Type: cty.String, Required: false},
		"interface":  &hcldec.AttrSpec{Name: "interface", Type: cty.String, Required: false},
		"vnet":       &hcldec.AttrSpec{Name: "vnet", Type: cty.String, Required: false},
		"status":     &hcldec.AttrSpec{Name: "status", Type: cty.String, Required: false},
		"ipaddress":  &hcldec.AttrSpec{Name: "ipaddress", Type: cty.String, Required: false},
		"macaddress": &hcldec.AttrSpec{Name: "macaddress", Type: cty.String, Required: false},
	}
	return s
}

// FlatVMOutput is an auto-generated flat version of VMOutput.
// Where the contents of a field with a `mapstructure:,squash` tag are bubbled up.
type FlatVMOutput struct {
	VMs []FlatVMInfo `mapstructure:"vms" cty:"vms" hcl:"vms"`
}

// FlatMapstructure returns a new FlatVMOutput.
// FlatVMOutput is an auto-generated flat version of VMOutput.
// Where the contents a fields with a `mapstructure:,squash` tag are bubbled up.
func (*VMOutput) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatVMOutput)
}

// HCL2Spec returns the hcl spec of a VMOutput.
// This spec is used by HCL to read the fields of VMOutput.
// The decoded values from this spec will then be applied to a FlatVMOutput.
func (*FlatVMOutput) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"vms": &hcldec.BlockListSpec{TypeName: "vms", Nested: hcldec.ObjectSpec((*FlatVMInfo)(nil).HCL2Spec())},
	}
	return s
}