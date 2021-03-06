// Package tplink collects the device state information.
package tplink

import (
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/shauncampbell/golang-tplink-hs100/pkg/configuration"
	"github.com/shauncampbell/golang-tplink-hs100/pkg/hs100"
	"github.com/shauncampbell/tplink2mqtt/pkg/tplink"
)

// TPLink collects the device state information.
type TPLink interface {
	CollectDeviceStates() ([]*tplink.Device, error)
	CollectDeviceState(address string) (*tplink.Device, error)
}

type tplinkImpl struct {
	logger  *zerolog.Logger
	timeout time.Duration
	subnet  string
}

// CollectDeviceStates collects the status of the device
func (t *tplinkImpl) CollectDeviceStates() ([]*tplink.Device, error) {
	logger := t.logger.With().Str("subnet", t.subnet).Dur("timeout", t.timeout).Logger()
	logger.Info().Msgf("beginning discovery")
	devices, err := hs100.Discover(t.subnet,
		configuration.Default().WithTimeout(t.timeout),
	)

	if err != nil {
		logger.Err(err).Msgf("failed to collect device states")
		return nil, err
	}

	logger.Info().Msgf("found %d devices", len(devices))
	states := make([]*tplink.Device, 0)
	for _, d := range devices {
		state, err := t.collectDeviceState(d)
		if err != nil {
			logger.Error().Msgf("failed to collect device state for %s: %s", d.Address, err.Error())
			continue
		}
		states = append(states, state)
	}

	return states, nil
}

func (t *tplinkImpl) collectDeviceState(d *hs100.Hs100) (*tplink.Device, error) {
	info, err := d.GetInfo()
	if err != nil {
		t.logger.Error().Msgf("failed to retrieve device info: %s", err.Error())
		return nil, fmt.Errorf("failed to collect device state: %w", err)
	}

	state := &tplink.Device{
		ID: fmt.Sprintf("0x%s", strings.ToLower(info.System.SystemInfo.DeviceID)),
		State: tplink.DeviceState{
			IsOn: info.System.SystemInfo.RelayState == 1,
		},
		Info: tplink.DeviceInfo{
			FriendlyName:   info.System.SystemInfo.Alias,
			Model:          info.System.SystemInfo.Model,
			NetworkAddress: d.Address,
			Vendor:         "TPLink",
			Exposes:        []tplink.DeviceAttribute{tplink.OnDeviceAttribute},
		},
	}

	powerConsumption, err := d.GetCurrentPowerConsumption()
	if err == nil {
		state.State.Voltage = powerConsumption.Voltage
		state.State.Power = powerConsumption.Power
		state.State.Current = powerConsumption.Current
		state.Info.Exposes = append(state.Info.Exposes,
			tplink.VoltageDeviceAttribute, tplink.PowerDeviceAttribute, tplink.CurrentDeviceAttribute)
	} else {
		t.logger.Warn().Msgf("failed to collect power consumption: %s", err.Error())
	}
	return state, nil
}

// CollectDeviceState collects the device state for a single device.
func (t *tplinkImpl) CollectDeviceState(address string) (*tplink.Device, error) {
	return t.collectDeviceState(hs100.NewHs100(address, configuration.Default()))
}

// New creates a new TPLink instance.
func New(subnet string, timeout time.Duration, logger *zerolog.Logger) TPLink {
	return &tplinkImpl{
		logger:  logger,
		timeout: timeout,
		subnet:  subnet,
	}
}
