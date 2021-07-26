// Package config contains configuration utilities.
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config is a struct which contains the configuration for the application.
type Config struct {
	MQTT struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	} `mapstructure:"mqtt"`
	Subnet   string `mapstructure:"subnet"`
	Timeout  int    `mapstructure:"timeout"`
	Interval int    `mapstructure:"interval"`
}

// Read reads in the configuration from the environment.
func Read() (*Config, error) {
	// MQTT config options
	viper.SetDefault("mqtt.host", "")
	viper.SetDefault("mqtt.port", 1883)
	viper.SetDefault("mqtt.username", "")
	viper.SetDefault("mqtt.password", "")
	viper.SetDefault("subnet", "192.168.2.0/24")
	viper.SetDefault("timeout", 5)
	viper.SetDefault("interval", 30)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("TPLINK")
	viper.AutomaticEnv()

	var config Config
	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal configuration: %w", err)
	}

	return &config, nil
}
