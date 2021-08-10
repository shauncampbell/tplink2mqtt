package tplink2mqtt

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/shauncampbell/golang-tplink-hs100/pkg/configuration"
	"github.com/shauncampbell/golang-tplink-hs100/pkg/hs100"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	tplinkModel "github.com/shauncampbell/tplink2mqtt/pkg/tplink"
)

const (
	homeAssistantTopicFmt = "homeassistant/switch/%s/%s"
	on                    = "ON"
	off                   = "OFF"
)

var homeAssistantTopicRegex = regexp.MustCompile(`^homeassistant/switch/(.+?)/set$`)

func (h *Handler) handleHomeAssistantUpdate(client mqtt.Client, message mqtt.Message) {
	logger := h.logger.With().Str("topic", message.Topic()).Logger()
	if !homeAssistantTopicRegex.MatchString(message.Topic()) {
		logger.Error().Msgf("unable to determine device id from topic")
		return
	}

	deviceID := homeAssistantTopicRegex.FindStringSubmatch(message.Topic())
	if len(deviceID) == 0 {
		logger.Error().Msgf("unable to determine device id from topic")
		return
	}
	logger = logger.With().Str("device_id", deviceID[0]).Logger()
	payload := message.Payload()
	logger.Info().Msgf("received request to set state of device to %s", string(payload))

	device := h.devices[deviceID[0]]
	if device == nil {
		logger.Error().Msgf("unknown device: %s", deviceID[0])
		return
	}

	ipAddress := device.Info.NetworkAddress

	conn := hs100.NewHs100(ipAddress, configuration.Default().WithTimeout(time.Duration(h.config.Timeout)*time.Second))

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
}

func (h *Handler) subscribeToHomeAssistant(device *tplinkModel.Device, client mqtt.Client) {
	if !client.IsConnected() {
		token := client.Connect()
		if token.Wait() && token.Error() != nil {
			h.logger.Error().Msgf("failed to connect to mqtt: %s", token.Error().Error())
		}
	}

	setTopic := fmt.Sprintf(homeAssistantTopicFmt, device.ID, "set")
	token := client.Subscribe(setTopic, 1, h.handleHomeAssistantUpdate)
	if token.Wait() && token.Error() != nil {
		h.logger.Error().Msgf("failed to subscribe to home assistant device state: %s", token.Error().Error())
	}
	h.logger.Info().Msgf("subscribed to %s", setTopic)
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

	configTopic := fmt.Sprintf(homeAssistantTopicFmt, device.ID, "config")
	stateTopic := fmt.Sprintf(homeAssistantTopicFmt, device.ID, "state")

	h.logger.Info().Msgf("publishing device config to %s", configTopic)
	token := client.Publish(configTopic, 1, false, b)
	if token.Wait() && token.Error() != nil {
		h.logger.Error().Msgf("failed to publish device to home assistant: %s", token.Error().Error())
	}

	var state = "OFF"
	if device.State.IsOn {
		state = "ON"
	}
	h.logger.Info().Msgf("publishing device state to %s", stateTopic)
	token = client.Publish(stateTopic, 1, false, []byte(state))
	if token.Wait() && token.Error() != nil {
		h.logger.Error().Msgf("failed to publish device to home assistant: %s", token.Error().Error())
	}
}
