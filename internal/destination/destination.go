// Package destination provides interfaces for publishing device status change events to mqtt.
package destination

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/shauncampbell/tplink2mqtt/pkg/tplink"
)

// Destination is an interface which defines somewhere which events are published when a device changes state.
type Destination interface {
	Publish(device *tplink.Device, client mqtt.Client) error
}
