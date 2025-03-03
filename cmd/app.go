package cmd

import (
	"time"

	"screendaemon/internal/config"
	"screendaemon/internal/logging"
	"screendaemon/internal/mqtt"

	"github.com/spf13/viper"
)

// Execute initializes the application, loads configuration, and starts the MQTT client
func Execute(version string, exit func(int), args []string) {
	// Load application configuration
	appConfig, err := config.Load(version, exit, args)
	if err != nil {
		panic(err)
	}

	// Initialize logger
	logger := logging.CreateLogger(appConfig.LoggerConfig.Path)
	defer logger.Sync() // Ensure logger buffer is flushed on exit
	logger.Infow("Application started", "config_file", viper.ConfigFileUsed())

	// Instrument MQTT logging
	logging.InstrumentMqtt(logger)

	// Initialize MQTT client with app configuration
	client, err := mqtt.Init(appConfig.AppId, &appConfig.Mqtt, appConfig.Switches, logger)
	if err != nil {
		logger.Panic(err)
	}

	// Start client refresh loop
	for {
		client.Refresh()
		time.Sleep(10 * time.Second)
	}
}
