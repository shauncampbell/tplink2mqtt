// Package tplink2mqtt contains a handler for handling tplink updates and piping them to mqtt.
package tplink2mqtt

import (
	"time"

	"github.com/shauncampbell/tplink2mqtt/internal/destination"
	"github.com/shauncampbell/tplink2mqtt/internal/listener"

	"github.com/shauncampbell/tplink2mqtt/internal/tplink"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shauncampbell/tplink2mqtt/internal/config"
	tplinkModel "github.com/shauncampbell/tplink2mqtt/pkg/tplink"
)

// Handler handles zigbee2mqtt messages
type Handler struct {
	config       *config.Config
	stopped      bool
	logger       zerolog.Logger
	devices      map[string]*tplinkModel.Device
	destinations []destination.Destination
	listeners    []listener.Listener
}

// Connected is a handler which is called when the initial connection to the mqtt server is established.
func (h *Handler) Connected(client mqtt.Client) {
	h.stopped = false
	go h.publishDeviceList(client)
}

// Disconnected is a handler which is called when the connection to the mqtt server is severed.
func (h *Handler) Disconnected(client mqtt.Client, err error) {
	h.stopped = true
}

func (h *Handler) publishDeviceList(client mqtt.Client) {
	for {
		tpClient := tplink.New(h.config.Subnet, time.Second*time.Duration(h.config.Timeout), &log.Logger)
		devices, err := tpClient.CollectDeviceStates()
		if err != nil {
			h.logger.Error().Msgf("failed to collect device states: %s", err.Error())
			return
		}

		for _, device := range devices {
			h.publishDeviceStatus(device, client)
		}

		time.Sleep(time.Duration(h.config.Interval) * time.Second)
	}
}

func (h *Handler) publishDeviceStatus(device *tplinkModel.Device, client mqtt.Client) {
	var err error
	for _, dest := range h.destinations {
		err = dest.Publish(device, client)
		if err != nil {
			h.logger.Error().Msgf("failed to publish to destination: %s", err.Error())
			continue
		}
	}

	for _, list := range h.listeners {
		err = list.Listen(device, client, h.publishDeviceStatus)
		if err != nil {
			h.logger.Error().Msgf("failed to subscribe to listener: %s", err.Error())
			continue
		}
	}
}

// New creates a new handler.
func New(cfg *config.Config, destinations []destination.Destination, listeners []listener.Listener) *Handler {
	return &Handler{
		devices:      make(map[string]*tplinkModel.Device),
		destinations: destinations,
		listeners:    listeners,
		logger:       log.Logger,
		config:       cfg}
}
