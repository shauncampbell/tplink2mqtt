// Package listener defines an interface for listening to events
package listener

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/shauncampbell/tplink2mqtt/pkg/tplink"
)

// Listener is an interface which defines a location which will listen for events for a specific device.
type Listener interface {
	Listen(device *tplink.Device, client mqtt.Client, callback StateChangedCallback) error
}

// StateChangedCallback is an interface which defines a callback where a listener can tell the rest of
// the system a device state has changed.
type StateChangedCallback func(device *tplink.Device, client mqtt.Client)
