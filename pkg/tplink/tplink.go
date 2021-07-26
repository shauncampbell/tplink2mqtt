// Package tplink contains model structs
package tplink

// DeviceInfo represents mostly static information about the device.
type DeviceInfo struct {
	FriendlyName   string   `json:"friendly_name"`
	Model          string   `json:"model"`
	NetworkAddress string   `json:"network_address"`
	Vendor         string   `json:"vendor"`
	Exposes        []string `json:"exposes"`
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
