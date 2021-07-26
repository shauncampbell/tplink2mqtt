package tplink2mqtt

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shauncampbell/tplink2mqtt/internal/config"
	"github.com/shauncampbell/tplink2mqtt/internal/tplink"
	"strings"
	"time"
)

const defaultTimeout = 5 * time.Second

// Handler handles zigbee2mqtt messages
type Handler struct {
	config  *config.Config
	stopped bool
	logger  zerolog.Logger
	devices map[string]*tplink.Device
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
		h.publishDeviceStatus(device, client)
	}

	go func() {
		time.Sleep(time.Duration(h.config.Interval)*time.Second)
		h.publishDeviceList(client)
	}()
}

func (h *Handler) publishDeviceStatus(device *tplink.Device, client mqtt.Client) {
	event := make(map[string]interface{})
	event["id"] = device.ID
	for _, field := range device.Info.Exposes {
		switch field {
		case "on":
			event[field] = device.State.IsOn
		case "voltage":
			event[field] = device.State.Voltage
		case "current":
			event[field] = device.State.Current
		case "power":
			event[field] = device.State.Power
		}
	}

	b, err := json.Marshal(event)
	if err != nil {
		h.logger.Error().Msgf("failed to create json: %s", err.Error())
		return
	}

	h.logger.Info().Msgf("publishing device state to tplink2mqtt/%s", sanitizeFriendlyName(device.Info.FriendlyName))
	token := client.Publish(
		fmt.Sprintf("tplink2mqtt/%s", sanitizeFriendlyName(device.Info.FriendlyName)), 1, true, b)
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
func New(config *config.Config) *Handler {
	return &Handler{devices: make(map[string]*tplink.Device), logger: log.Logger, config: config}
}
