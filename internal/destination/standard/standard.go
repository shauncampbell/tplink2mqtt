// Package standard publishes to standard mqtt channels.
package standard

import (
	"encoding/json"
	"fmt"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shauncampbell/tplink2mqtt/internal/destination"
	"github.com/shauncampbell/tplink2mqtt/pkg/tplink"
)

const configTopic = "tplink2mqtt/bridge/devices"

// Standard is a destination for mqtt events.
type Standard struct {
	options Options
	devices map[string]*tplink.Device
	logger  zerolog.Logger
	destination.Destination
}

// Options is a struct for storing options for the standard mqtt destination
type Options struct {
	// This is intentionally empty for now, but will have config options in future.
}

// New creates a new standard destination.
func New(options Options) destination.Destination {
	return &Standard{options: options, logger: log.Logger, devices: make(map[string]*tplink.Device)}
}

// Publish publishes the device state to the standard mqtt destinations
func (s *Standard) Publish(device *tplink.Device, client mqtt.Client) error {
	err := s.publishDeviceConfiguration(device, client)
	if err != nil {
		s.logger.Error().Msgf("failed to publish device configuration: %s", err.Error())
		return err
	}

	err = s.publishDeviceState(device, client)
	if err != nil {
		s.logger.Error().Msgf("failed to publish device state: %s", err.Error())
		return err
	}

	return nil
}

func (s *Standard) publishDeviceConfiguration(device *tplink.Device, client mqtt.Client) error {
	if s.devices[device.ID] != nil {
		// If the device has already been published no need to publish it again.
		return nil
	}

	event := s.getDeviceConfiguration()
	b, err := json.Marshal(event)
	if err != nil {
		s.logger.Error().Msgf("failed to create json: %s", err.Error())
		return err
	}

	s.logger.Info().Msgf("publishing device config to %s", configTopic)
	token := client.Publish(configTopic, 1, true, b)
	if token.Wait() && token.Error() != nil {
		s.logger.Error().Msgf("failed to publish device to home assistant: %s", token.Error().Error())
		return token.Error()
	}

	s.devices[device.ID] = device
	return nil
}

func (s *Standard) publishDeviceState(device *tplink.Device, client mqtt.Client) error {
	event := make(map[string]interface{})
	event["id"] = device.ID
	for _, field := range device.Info.Exposes {
		switch field.Property {
		case "on":
			event[field.Property] = device.State.IsOn
		case "voltage":
			event[field.Property] = device.State.Voltage
		case "current":
			event[field.Property] = device.State.Current
		case "power":
			event[field.Property] = device.State.Power
		}
	}

	b, err := json.Marshal(event)
	if err != nil {
		s.logger.Error().Msgf("failed to create json: %s", err.Error())
		return err
	}

	s.logger.Info().Msgf("publishing device state to tplink2mqtt/%s", sanitizeFriendlyName(device.Info.FriendlyName))
	token := client.Publish(
		fmt.Sprintf("tplink2mqtt/%s", sanitizeFriendlyName(device.Info.FriendlyName)), 1, false, b)
	if token.Wait() && token.Error() != nil {
		s.logger.Error().Msgf("failed to publish device list: %s", token.Error().Error())
		return err
	}
	return nil
}

func (s *Standard) getDeviceConfiguration() []*tplink.Device {
	out := make([]*tplink.Device, 0)
	for _, v := range s.devices {
		out = append(out, v)
	}
	return out
}

func sanitizeFriendlyName(friendlyName string) string {
	str := strings.ToLower(friendlyName)
	str = strings.ReplaceAll(str, " ", "_")
	return str
}
