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

// Config struct definitions
type MqttConfig struct {
	Broker   string  `mapstructure:"broker"`
	User     *string `mapstructure:"username"`
	Password *string `mapstructure:"password"`
}

type LoggerConfig struct {
	Path string `mapstructure:"path"`
}

type ApplicationConfig struct {
	AppId        string            `mapstructure:"app-id"`
	Mqtt         MqttConfig        `mapstructure:"mqtt"`
	Switches     []controls.Switch `mapstructure:"switches"`
	LoggerConfig LoggerConfig      `mapstructure:"log"`
}

// Load configuration from file and environment variables
func Load(version string, exit func(int), args []string) (*ApplicationConfig, error) {
	processCommandLineArguments(version, exit, args)

	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return nil, err
	}

	viper.SetEnvPrefix(strings.ToUpper(AppName))
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Load config.yaml
	viper.SetConfigFile(defaultConfigFile())
	viper.SetDefault("app-id", AppName)
	err = viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config.yaml: %w", err)
	}

	config := ApplicationConfig{}
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config.yaml: %w", err)
	}

	// Fallback für leere MQTT-Werte
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

func processCommandLineArguments(versionStr string, exit func(int), args []string) {
	pflag.StringP("config", "c", defaultConfigFile(), "Configuration file path")
	pflag.StringP("mqtt.broker", "b", "tcp://localhost:1883", "MQTT broker")
	pflag.StringP("log.path", "l", defaultLogFile(), "Log file path")
	pflag.BoolP("version", "v", false, "Show version")

	_ = pflag.CommandLine.Parse(args)

	if viper.GetBool("version") {
		fmt.Printf("%s version %s\n", AppName, versionStr)
		exit(0)
	}
}

func defaultConfigFile() string {
	return "/workspace/internal/config/config.yaml"
}

func defaultLogFile() string {
	return "/workspace/internal/config/screendaemon.log"
}
