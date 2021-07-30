// Package tplink contains model structs
package tplink

// DeviceInfo represents mostly static information about the device.
type DeviceInfo struct {
	FriendlyName   string   `json:"friendly_name"`
	Model          string   `json:"model"`
	NetworkAddress string   `json:"network_address"`
	Vendor         string   `json:"vendor"`
	Exposes        []DeviceAttribute `json:"exposes"`
}

// DeviceAttribute is an attribute which the device exposes to the user.
type DeviceAttribute struct {
	Access      int    `json:"access"`
	Description string `json:"description"`
	Name        string `json:"name"`
	Property    string `json:"property"`
	Type        string `json:"type"`
	Unit        string `json:"unit"`
	ValueMax    int    `json:"value_max"`
	ValueMin    int    `json:"value_min"`
}

// OnDeviceAttribute is the attribute for power status.
var OnDeviceAttribute = DeviceAttribute{
	Access: 1,
	Description: "Power Status of the Switch",
	Property: "on",
	Name: "on",
	Type: "binary",
	Unit: "",
	ValueMax: 1,
	ValueMin: 0,
}

// VoltageDeviceAttribute is the attribute for voltage.
var VoltageDeviceAttribute = DeviceAttribute{
	Access:      1,
	Description: "Measured input voltage",
	Name:        "voltage",
	Property:    "voltage",
	Type:        "numeric",
	Unit:        "V",
	ValueMax:    -250,
	ValueMin:    250,
}

// PowerDeviceAttribute is the attribute for power consumption.
var PowerDeviceAttribute = DeviceAttribute{
	Access:      1,
	Description: "Measured output power",
	Name:        "power",
	Property:    "power",
	Type:        "numeric",
	Unit:        "W",
	ValueMax:    2400,
	ValueMin:    0,
}

// CurrentDeviceAttribute is the attribute for output current.
var CurrentDeviceAttribute = DeviceAttribute{
	Access:      1,
	Description: "Measured output current",
	Name:        "current",
	Property:    "current",
	Type:        "numeric",
	Unit:        "A",
	ValueMax:    15,
	ValueMin:    0,
}

// DeviceState represents information about the device which changes.
type DeviceState struct {
	IsOn    bool    `json:"is_on"`
	Current float32 `json:"current"`
	Power   float32 `json:"power"`
	Voltage float32 `json:"voltage"`
}

// Device represents the hs1xx device.
type Device struct {
	ID    string      `json:"id"`
	State DeviceState `json:"state"`
	Info  DeviceInfo  `json:"info"`
}
