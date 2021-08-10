// Package tplink2mqtt contains a handler for handling tplink updates and piping them to mqtt.
package tplink2mqtt

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/shauncampbell/tplink2mqtt/internal/tplink"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shauncampbell/tplink2mqtt/internal/config"
	tplinkModel "github.com/shauncampbell/tplink2mqtt/pkg/tplink"
)

// Handler handles zigbee2mqtt messages
type Handler struct {
	config  *config.Config
	stopped bool
	logger  zerolog.Logger
	devices map[string]*tplinkModel.Device
}

// Connected is a handler which is called when the initial connection to the mqtt server is established.
func (h *Handler) Connected(client mqtt.Client) {
	h.stopped = false
	h.publishDeviceList(client)
}

// Disconnected is a handler which is called when the connection to the mqtt server is severed.
func (h *Handler) Disconnected(client mqtt.Client, err error) {
	h.stopped = true
}

func (h *Handler) publishDeviceList(client mqtt.Client) {
	tpClient := tplink.New(h.config.Subnet, time.Second*time.Duration(h.config.Timeout), &log.Logger)
	devices, err := tpClient.CollectDeviceStates()
	if err != nil {
		h.logger.Error().Msgf("failed to collect device states: %s", err.Error())
		return
	}

	b, err := json.Marshal(&devices)
	if err != nil {
		h.logger.Error().Msgf("failed to create json: %s", err.Error())
		return
	}

	h.logger.Info().Msgf("publishing device list")
	token := client.Publish("tplink2mqtt/bridge/devices", 1, true, b)
	if token.Wait() && token.Error() != nil {
		h.logger.Error().Msgf("failed to publish device list: %s", token.Error().Error())
	}

	for _, device := range devices {
		if h.devices[device.ID] == nil || !h.devices[device.ID].IsEqualTo(device) {
			h.publishDeviceStatus(device, client)
			h.subscribeToHomeAssistant(device, client)
		}

		h.devices[device.ID] = device
	}

	time.Sleep(time.Duration(h.config.Interval) * time.Second)
	go func() {
		h.publishDeviceList(client)
	}()
}

func (h *Handler) publishDeviceStatus(device *tplinkModel.Device, client mqtt.Client) {
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
		h.logger.Error().Msgf("failed to create json: %s", err.Error())
		return
	}

	h.logger.Info().Msgf("publishing device state to tplink2mqtt/%s", sanitizeFriendlyName(device.Info.FriendlyName))
	token := client.Publish(
		fmt.Sprintf("tplink2mqtt/%s", sanitizeFriendlyName(device.Info.FriendlyName)), 1, false, b)
	if token.Wait() && token.Error() != nil {
		h.logger.Error().Msgf("failed to publish device list: %s", token.Error().Error())
	}
	h.publishDeviceToHomeAssistant(device, client)
}

func (h *Handler) subscribeToHomeAssistant(device *tplinkModel.Device, client mqtt.Client) {
	if !client.IsConnected() {
		token := client.Connect()
		if token.Wait() && token.Error() != nil {
			h.logger.Error().Msgf("failed to connect to mqtt: %s", token.Error().Error())
		}
	}

	token := client.Subscribe(fmt.Sprintf("homeassistant/switch/%s/set", device.ID), 1, h.handleHomeAssistantUpdate)
	if token.Wait() && token.Error() != nil {
		h.logger.Error().Msgf("failed to subscribe to home assistant device state: %s", token.Error().Error())
	}
	h.logger.Info().Msgf("subscribed to %s", fmt.Sprintf("homeassistant/switch/%s/set", device.ID))
}
func (h *Handler) handleHomeAssistantUpdate(client mqtt.Client, message mqtt.Message) {
	payload := message.Payload()
	h.logger.Info().Msgf("received message from home assistant with %s", string(payload))
}
func (h *Handler) publishDeviceToHomeAssistant(device *tplinkModel.Device, client mqtt.Client) {
	event := make(map[string]interface{})
	event["name"] = device.Info.FriendlyName
	event["command_topic"] = "homeassistant/switch/" + device.ID + "/set"
	event["state_topic"] = "homeassistant/switch/" + device.ID + "/state"
	event["device"] = map[string]interface{}{
		"manufacturer": device.Info.Vendor,
		"connections":  [][]string{{"ip", device.Info.NetworkAddress}},
		"identifiers":  []string{device.ID},
		"model":        device.Info.Model,
		"name":         device.Info.FriendlyName,
	}
	event["unique_id"] = device.ID

	b, err := json.Marshal(event)
	if err != nil {
		h.logger.Error().Msgf("failed to create json: %s", err.Error())
		return
	}

	h.logger.Info().Msgf("publishing device config to %s", fmt.Sprintf("homeassistant/switch/%s/state", device.ID))
	token := client.Publish(fmt.Sprintf("homeassistant/switch/%s/config", device.ID), 1, false, b)
	if token.Wait() && token.Error() != nil {
		h.logger.Error().Msgf("failed to publish device to home assistant: %s", token.Error().Error())
	}

	var state = "OFF"
	if device.State.IsOn {
		state = "ON"
	}
	h.logger.Info().Msgf("publishing device state to %s", fmt.Sprintf("homeassistant/switch/%s/state", device.ID))
	token = client.Publish(fmt.Sprintf("homeassistant/switch/%s/state", device.ID), 1, false, []byte(state))
	if token.Wait() && token.Error() != nil {
		h.logger.Error().Msgf("failed to publish device to home assistant: %s", token.Error().Error())
	}
}

func sanitizeFriendlyName(friendlyName string) string {
	str := strings.ToLower(friendlyName)
	str = strings.ReplaceAll(str, " ", "_")
	return str
}

// New creates a new handler.
func New(cfg *config.Config) *Handler {
	return &Handler{devices: make(map[string]*tplinkModel.Device), logger: log.Logger, config: cfg}
}
