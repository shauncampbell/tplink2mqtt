// Package tplink collects the device state information.
package tplink

import (
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/shauncampbell/golang-tplink-hs100/pkg/configuration"
	"github.com/shauncampbell/golang-tplink-hs100/pkg/hs100"
)

// TPLink collects the device state information.
type TPLink interface {
	CollectDeviceStates() ([]*Device, error)
}

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

type tplink struct {
	logger  *zerolog.Logger
	timeout time.Duration
	subnet  string
}

// CollectDeviceStates collects the status of the device
func (t *tplink) CollectDeviceStates() ([]*Device, error) {
	logger := t.logger.With().Str("subnet", t.subnet).Dur("timeout", t.timeout).Logger()
	logger.Info().Msgf("beginning discovery")
	devices, err := hs100.Discover(t.subnet,
		configuration.Default().WithTimeout(t.timeout),
	)

	if err != nil {
		logger.Err(err).Msgf("failed to collect device states")
		return nil, err
	}

	t.logger.Info().Msgf("found %d devices", len(devices))
	states := make([]*Device, 0)
	for _, d := range devices {
		info, _ := d.GetInfo()

		state := &Device{
			ID: fmt.Sprintf("0x%s", strings.ToLower(info.System.SystemInfo.DeviceID)),
			State: DeviceState{
				IsOn: info.System.SystemInfo.RelayState == 1,
			},
			Info: DeviceInfo{
				FriendlyName:   info.System.SystemInfo.Alias,
				Model:          info.System.SystemInfo.Model,
				NetworkAddress: d.Address,
				Vendor:         "TPLink",
				Exposes:        []string{"on"},
			},
		}

		powerConsumption, err := d.GetCurrentPowerConsumption()
		if err == nil {
			state.State.Voltage = powerConsumption.Voltage
			state.State.Power = powerConsumption.Power
			state.State.Current = powerConsumption.Current
			state.Info.Exposes = append(state.Info.Exposes, "voltage", "power", "current")
		} else {
			t.logger.Warn().Msgf("failed to collect power consumption: %s", err.Error())
		}
		states = append(states, state)
	}

	return states, nil
}

// New creates a new TPLink instance.
func New(subnet string, timeout time.Duration, logger *zerolog.Logger) TPLink {
	return &tplink{
		logger:  logger,
		timeout: timeout,
		subnet:  subnet,
	}
}
