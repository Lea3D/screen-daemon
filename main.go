package config

import (
	"fmt"
	"os"
	"strings"

	"screendaemon/internal/controls"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const AppName = "screendaemon"

// MqttConfig holds the MQTT broker connection details
type MqttConfig struct {
	Broker   string  `mapstructure:"broker"`   // MQTT broker address
	User     *string `mapstructure:"username"` // Optional MQTT username
	Password *string `mapstructure:"password"` // Optional MQTT password
}

// LoggerConfig defines the configuration for logging
type LoggerConfig struct {
	Path string `mapstructure:"path"` // Path to the log file
}

// ApplicationConfig aggregates all application-specific configuration.
type ApplicationConfig struct {
	AppId        string             `mapstructure:"app-id"`   // Application ID used as MQTT topic prefix
	Mqtt         MqttConfig         `mapstructure:"mqtt"`     // MQTT configuration
	Displays     []controls.Display `mapstructure:"displays"` // List of configured displays
	LoggerConfig LoggerConfig       `mapstructure:"log"`      // Logger configuration
}

// Load configuration from file and environment variables.
func Load(version string, exit func(int), args []string) (*ApplicationConfig, error) {
	processCommandLineArguments(version, exit, args)

	// Bind command line flags to viper.
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return nil, err
	}

	// Setup environment variable handling.
	viper.SetEnvPrefix(strings.ToUpper(AppName))
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Load the config.yaml file.
	viper.SetConfigFile(defaultConfigFile())
	viper.SetDefault("app-id", AppName)
	err = viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config.yaml: %w", err)
	}

	// Unmarshal the configuration into the ApplicationConfig struct.
	config := ApplicationConfig{}
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config.yaml: %w", err)
	}

	// Debug: Vollständige Display-Objekte ausgeben
	for _, display := range config.Displays {
		fmt.Printf("DEBUG: Loaded Display: %+v\n", display)
		if display.Command.GetCmd == "" {
			fmt.Printf("ERROR: get_state not defined for display %s\n", display.Name)
		} else {
			fmt.Printf("DEBUG: Loaded get_state for display %s: %s\n", display.Name, display.Command.GetCmd)
		}
	}

	// Fallback to environment variables if MQTT config values are not set.
	if config.Mqtt.Broker == "" {
		config.Mqtt.Broker = os.Getenv("MQTT_BROKER")
	}
	if config.Mqtt.User == nil || *config.Mqtt.User == "" {
		envUser := os.Getenv("MQTT_USERNAME")
		config.Mqtt.User = &envUser
	}
	if config.Mqtt.Password == nil || *config.Mqtt.Password == "" {
		envPassword := os.Getenv("MQTT_PASSWORD")
		config.Mqtt.Password = &envPassword
	}

	return &config, nil
}

// processCommandLineArguments sets up and parses CLI flags.
func processCommandLineArguments(versionStr string, exit func(int), args []string) {
	pflag.StringP("config", "c", defaultConfigFile(), "Configuration file path")
	pflag.StringP("mqtt.broker", "b", "tcp://localhost:1883", "MQTT broker")
	pflag.StringP("log.path", "l", defaultLogFile(), "Log file path")
	pflag.BoolP("version", "v", false, "Show version")

	_ = pflag.CommandLine.Parse(args)

	// Display version and exit if the version flag is provided.
	if viper.GetBool("version") {
		fmt.Printf("%s version %s\n", AppName, versionStr)
		exit(0)
	}
}

// defaultConfigFile returns the default path to the config.yaml file.
func defaultConfigFile() string {
	return "/workspace/internal/config/config.yaml"
}

// defaultLogFile returns the default path to the log file.
func defaultLogFile() string {
	return "/workspace/internal/config/screendaemon.log"
}
