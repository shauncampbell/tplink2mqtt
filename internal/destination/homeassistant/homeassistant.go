// Package homeassistant provides a destination which submits events to mqtt queues for consumption by home assistant.
package homeassistant

import (
	"encoding/json"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shauncampbell/tplink2mqtt/internal/destination"
	"github.com/shauncampbell/tplink2mqtt/pkg/tplink"
)

const (
	homeAssistantTopicFmt = "homeassistant/switch/%s/%s"
	on                    = "ON"
	off                   = "OFF"
)

// HomeAssistant is a destination for home assistant events.
type HomeAssistant struct {
	options Options
	logger  zerolog.Logger
	destination.Destination
}

// Options is a struct for storing options for the home assistant destination.
type Options struct {
	// This is intentionally empty for now, but will have config options in future.
}

// Publish publishes the device state to Home Assistant
func (h *HomeAssistant) Publish(device *tplink.Device, client mqtt.Client) error {
	err := h.publishDeviceConfiguration(device, client)
	if err != nil {
		h.logger.Error().Msgf("failed to publish device configuration: %s", err.Error())
		return err
	}

	err = h.publishDeviceState(device, client)
	if err != nil {
		h.logger.Error().Msgf("failed to publish device state: %s", err.Error())
		return err
	}

	return nil
}

func (h *HomeAssistant) publishDeviceConfiguration(device *tplink.Device, client mqtt.Client) error {
	event := getDeviceConfiguration(device)
	b, err := json.Marshal(event)
	if err != nil {
		h.logger.Error().Msgf("failed to create json: %s", err.Error())
		return err
	}
	configTopic := fmt.Sprintf(homeAssistantTopicFmt, device.ID, "config")
	h.logger.Info().Msgf("publishing device config to %s", configTopic)
	token := client.Publish(configTopic, 1, true, b)
	if token.Wait() && token.Error() != nil {
		h.logger.Error().Msgf("failed to publish device to home assistant: %s", token.Error().Error())
		return token.Error()
	}
	return nil
}

func (h *HomeAssistant) publishDeviceState(device *tplink.Device, client mqtt.Client) error {
	stateTopic := fmt.Sprintf(homeAssistantTopicFmt, device.ID, "state")
	var state = off
	if device.State.IsOn {
		state = on
	}
	h.logger.Info().Msgf("publishing device state to %s", stateTopic)
	token := client.Publish(stateTopic, 1, true, []byte(state))
	if token.Wait() && token.Error() != nil {
		h.logger.Error().Msgf("failed to publish device to home assistant: %s", token.Error().Error())
		return token.Error()
	}
	return nil
}

func getDeviceConfiguration(device *tplink.Device) *deviceConfiguration {
	return &deviceConfiguration{
		Name:         device.Info.FriendlyName,
		CommandTopic: fmt.Sprintf(homeAssistantTopicFmt, device.ID, "set"),
		StateTopic:   fmt.Sprintf(homeAssistantTopicFmt, device.ID, "state"),
		Device: deviceInfo{
			Manufacturer: device.Info.Vendor,
			Connections:  []connection{{"ip", device.Info.NetworkAddress}},
			Identifiers:  []string{device.ID},
			Model:        device.Info.Model,
			Name:         device.Info.FriendlyName,
		},
		UniqueID: device.ID,
	}
}

// New creates a new Home Assistant destination.
func New(options Options) destination.Destination {
	return &HomeAssistant{options: options, logger: log.Logger}
}
