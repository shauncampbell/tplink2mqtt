// Package homeassistant provides a listener for listening to events from home assistant mqtt channels.
package homeassistant

import (
	"fmt"
	"regexp"
	"time"

	"github.com/rs/zerolog/log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog"
	"github.com/shauncampbell/golang-tplink-hs100/pkg/configuration"
	"github.com/shauncampbell/golang-tplink-hs100/pkg/hs100"
	"github.com/shauncampbell/tplink2mqtt/internal/listener"
	"github.com/shauncampbell/tplink2mqtt/internal/tplink"
	tplinkModel "github.com/shauncampbell/tplink2mqtt/pkg/tplink"
)

const (
	homeAssistantTopicFmt          = "homeassistant/switch/%s/%s"
	homeAssistantTopicRegexMatches = 2
	on                             = "ON"
	off                            = "OFF"
)

var homeAssistantTopicRegex = regexp.MustCompile(`^homeassistant/switch/(.+?)/set$`)

// HomeAssistant is a listener for home assistant events.
type HomeAssistant struct {
	options Options
	devices map[string]*tplinkModel.Device
	logger  zerolog.Logger
	listener.Listener
}

// Options is a struct for storing options for the home assistant listener.
type Options struct {
	Timeout int
	Subnet  string
}

// Listen listens for events on home assistant mqtt channels.
func (h *HomeAssistant) Listen(device *tplinkModel.Device, client mqtt.Client, callback listener.StateChangedCallback) error {
	if h.devices[device.ID] != nil {
		return nil
	}

	if !client.IsConnected() {
		token := client.Connect()
		if token.Wait() && token.Error() != nil {
			h.logger.Error().Msgf("failed to connect to mqtt: %s", token.Error().Error())
			return token.Error()
		}
	}

	setTopic := fmt.Sprintf(homeAssistantTopicFmt, device.ID, "set")
	token := client.Subscribe(setTopic, 1, h.handleHomeAssistantUpdate(callback))
	if token.Wait() && token.Error() != nil {
		h.logger.Error().Msgf("failed to subscribe to home assistant device state: %s", token.Error().Error())
		return token.Error()
	}
	h.logger.Info().Msgf("subscribed to %s", setTopic)
	h.devices[device.ID] = device
	return nil
}

func (h *HomeAssistant) handleHomeAssistantUpdate(callback listener.StateChangedCallback) mqtt.MessageHandler {
	return func(client mqtt.Client, message mqtt.Message) {
		logger := h.logger.With().Str("topic", message.Topic()).Logger()
		if !homeAssistantTopicRegex.MatchString(message.Topic()) {
			logger.Error().Msgf("unable to determine device id from topic")
			return
		}

		deviceID := homeAssistantTopicRegex.FindStringSubmatch(message.Topic())
		if len(deviceID) < homeAssistantTopicRegexMatches {
			logger.Error().Msgf("unable to determine device id from topic")
			return
		}
		logger = logger.With().Str("device_id", deviceID[1]).Logger()
		payload := message.Payload()
		logger.Info().Msgf("received request to set state of device to %s", string(payload))

		device := h.devices[deviceID[1]]
		if device == nil {
			logger.Error().Msgf("unknown device: %s", deviceID[1])
			return
		}

		ipAddress := device.Info.NetworkAddress

		conn := hs100.NewHs100(ipAddress, configuration.Default().WithTimeout(time.Duration(h.options.Timeout)*time.Second))

		if string(payload) == on {
			err := conn.TurnOn()
			if err != nil {
				logger.Error().Msgf("failed to turn on device: %s", err.Error())
				return
			}
		} else if string(payload) == off {
			err := conn.TurnOff()
			if err != nil {
				logger.Error().Msgf("failed to turn off device: %s", err.Error())
				return
			}
		}

		t := tplink.New(h.options.Subnet, time.Duration(h.options.Timeout)*time.Second, &logger)
		dstate, err := t.CollectDeviceState(ipAddress)
		if err != nil {
			logger.Error().Msgf("failed to collect device state: %s", err.Error())
			return
		}

		callback(dstate, client)
	}
}

// New creates a new Home Assistant destination.
func New(options Options) listener.Listener {
	return &HomeAssistant{options: options, logger: log.Logger, devices: make(map[string]*tplinkModel.Device)}
}
