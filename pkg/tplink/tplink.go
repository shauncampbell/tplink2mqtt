// Package tplink contains model structs
package tplink

// DeviceInfo represents mostly static information about the device.
type DeviceInfo struct {
	FriendlyName   string            `json:"friendly_name"`
	Model          string            `json:"model"`
	NetworkAddress string            `json:"network_address"`
	Vendor         string            `json:"vendor"`
	Exposes        []DeviceAttribute `json:"exposes"`
}

// IsEqualTo checks that this object is equal to another.
func (di *DeviceInfo) IsEqualTo(info *DeviceInfo) bool {
	attrEqual := len(di.Exposes) == len(info.Exposes)
	if !attrEqual {
		return false
	}

	for i := range di.Exposes {
		attrEqual = attrEqual && di.Exposes[i].IsEqualTo(&info.Exposes[i])
		if !attrEqual {
			return false
		}
	}

	return di.FriendlyName == info.FriendlyName &&
		di.Model == info.Model &&
		di.NetworkAddress == info.NetworkAddress &&
		di.Vendor == info.Vendor
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

// IsEqualTo checks that this object is equal to another.
func (da *DeviceAttribute) IsEqualTo(attribute *DeviceAttribute) bool {
	return da.Access == attribute.Access &&
		da.Description == attribute.Description &&
		da.Name == attribute.Name &&
		da.Property == attribute.Property &&
		da.Type == attribute.Type &&
		da.Unit == attribute.Unit &&
		da.ValueMax == attribute.ValueMax &&
		da.ValueMin == attribute.ValueMin
}

// OnDeviceAttribute is the attribute for power status.
var OnDeviceAttribute = DeviceAttribute{
	Access:      1,
	Description: "Power Status of the Switch",
	Property:    "on",
	Name:        "on",
	Type:        "binary",
	Unit:        "",
	ValueMax:    1,
	ValueMin:    0,
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

// IsEqualTo checks that this object is equal to another.
func (ds *DeviceState) IsEqualTo(deviceState DeviceState) bool {
	return ds.IsOn == deviceState.IsOn &&
		ds.Current == deviceState.Current &&
		ds.Power == deviceState.Power &&
		ds.Voltage == deviceState.Voltage
}

// Device represents the hs1xx device.
type Device struct {
	ID    string      `json:"id"`
	State DeviceState `json:"state"`
	Info  DeviceInfo  `json:"info"`
}

// IsEqualTo checks that this object is equal to another.
func (d *Device) IsEqualTo(device *Device) bool {
	return d.ID == device.ID && d.State.IsEqualTo(device.State) && d.Info.IsEqualTo(&device.Info)
}
