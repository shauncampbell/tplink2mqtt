package main

import (
	"fmt"
	"os"
	"time"

	"github.com/shauncampbell/tplink2mqtt/internal/destination"
	"github.com/shauncampbell/tplink2mqtt/internal/destination/homeassistant"
	"github.com/shauncampbell/tplink2mqtt/internal/destination/standard"

	"github.com/shauncampbell/tplink2mqtt/internal/config"
	"github.com/shauncampbell/tplink2mqtt/internal/tplink2mqtt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:  "tplink2mqtt",
	RunE: runApplication,
}

const defaultTimeout = 10 * time.Second

func runApplication(cmd *cobra.Command, args []string) error {
	cfg, err := config.Read()

	if err != nil {
		return fmt.Errorf("failed to read configuration: %w", err)
	}

	handler := tplink2mqtt.New(cfg, []destination.Destination{
		standard.New(standard.Options{}),
		homeassistant.New(homeassistant.Options{}),
	})

	mqttOptions := mqtt.NewClientOptions()
	mqttOptions.AddBroker(fmt.Sprintf("tcp://%s:%d", cfg.MQTT.Host, cfg.MQTT.Port))
	if cfg.MQTT.Username != "" {
		mqttOptions.Username = cfg.MQTT.Username
	}
	if cfg.MQTT.Password != "" {
		mqttOptions.Password = cfg.MQTT.Password
	}
	mqttOptions.SetClientID("tplink2mqtt")
	mqttOptions.OnConnect = handler.Connected
	mqttOptions.OnConnectionLost = handler.Disconnected
	mqttClient := mqtt.NewClient(mqttOptions)
	token := mqttClient.Connect()
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}

	for {
		time.Sleep(defaultTimeout)
		if !mqttClient.IsConnected() {
			log.Error().Msg("connection to mqtt was severed")
			break
		}
	}

	return err
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Error().Msgf("unable to run application: %s", err.Error())
		if _, e := fmt.Fprintf(os.Stderr, "unable to run application: %s\n", err.Error()); e != nil {
			log.Error().Msgf("unable to write to stderr: %s", e.Error())
			log.Error().Msgf("unable to run application: %s", err.Error())
		}
		os.Exit(1)
	}
	os.Exit(0)
}
