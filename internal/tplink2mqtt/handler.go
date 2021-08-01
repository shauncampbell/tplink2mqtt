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
		}

		h.devices[device.ID] = device
	}

	go func() {
		time.Sleep(time.Duration(h.config.Interval) * time.Second)
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
