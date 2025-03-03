package mqtt

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"screendaemon/internal/config"
	"screendaemon/internal/controls"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
)

type DisplayWrapper struct {
	display      controls.Display
	available    bool
	availableSet bool
	lastRefresh  time.Time
}

// Refresh ruft für jedes Display im Client die refreshOne-Methode auf.
func (client *Client) Refresh() {
	defer client.syncLog()
	for _, dw := range client.displays {
		client.refreshOne(dw)
	}
}

type Client struct {
	appName  string
	handle   MQTT.Client
	displays []*DisplayWrapper
	logger   *zap.SugaredLogger
}

func Init(appName string, config *config.MqttConfig, displays []controls.Display, logger *zap.SugaredLogger) (*Client, error) {
	client, err := Connect(appName, config, displays, logger)
	if err != nil {
		client.handle.Disconnect(0)
		return nil, err
	}
	return client, nil
}

func Connect(appName string, config *config.MqttConfig, displays []controls.Display, logger *zap.SugaredLogger) (*Client, error) {
	dws := make([]*DisplayWrapper, len(displays))
	for i, disp := range displays {
		dws[i] = &DisplayWrapper{display: disp}
	}
	client := &Client{appName: appName, displays: dws, logger: logger}

	opts := MQTT.NewClientOptions()
	opts.AddBroker(config.Broker)
	opts.SetOrderMatters(false)
	opts.SetClientID(client.generateClientId())
	opts.SetWill(client.appAvailabilityTopic(), "0", 1, true)
	if config.User != nil && config.Password != nil {
		opts.SetUsername(*config.User)
		opts.SetPassword(*config.Password)
	}

	opts.SetOnConnectHandler(func(handle MQTT.Client) {
		client.setAppAvailable()
		if err := client.Subscribe(); err != nil {
			client.logger.Errorw("Cannot subscribe", "error", err)
		}
	})

	client.handle = MQTT.NewClient(opts)
	if token := client.handle.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	logger.Infow("Connected to MQTT", "broker", config.Broker)
	return client, nil
}

func (client *Client) generateClientId() string {
	return fmt.Sprintf("%s-%016x", client.appName, rand.Uint64())
}

// Subscribe abonniert für jedes Display den entsprechenden Command-Topic.
func (client *Client) Subscribe() error {
	for _, dw := range client.displays {
		topic := client.commandTopic(dw)
		client.logger.Debugw("Subscribing", "topic", topic)
		if token := client.handle.Subscribe(topic, 1, func(mqttClient MQTT.Client, message MQTT.Message) {
			client.processSetPayload(dw, string(message.Payload()))
		}); token.Wait() && token.Error() != nil {
			return token.Error()
		}
	}
	return nil
}

// processSetPayload verarbeitet empfangene MQTT-Nachrichten für ein Display.
func (client *Client) processSetPayload(dw *DisplayWrapper, payload string) {
	defer client.syncLog()
	logger := client.logger.With(zap.String("display", dw.display.Name))
	logger.Infow("Received display command", "payload", payload)

	var commandPayload struct {
		Display string `json:"display"`
		Command string `json:"command"`
		Value   string `json:"value,omitempty"`
	}
	if err := json.Unmarshal([]byte(payload), &commandPayload); err != nil {
		logger.Errorw("Invalid JSON payload", "error", err)
		return
	}
	var value string
	switch strings.ToUpper(commandPayload.Command) {
	case "ON":
		value = "0x01" // Beispiel: ON entspricht "0x01"
	case "OFF":
		value = "0x00" // OFF entspricht "0x00"
	case "SET":
		// Erwarte, dass der Wert als hexadezimaler String übergeben wird
		value = commandPayload.Value
	default:
		logger.Errorw("Invalid command", "command", commandPayload.Command)
		return
	}
	// Führe den VCP-Befehl aus – hier wird SetValue aufgerufen, das den Wert direkt setzt.
	response, err := dw.display.Command.SetValue(value)
	if err != nil {
		logger.Errorw("Error running set value command", "error", err, "output", response)
		client.setAvailable(dw, false)
		return
	}
	logger.Debugw("Executed value command successfully", "output", response)
	topic := client.stateTopic(dw)
	if err := client.publish(topic, value); err != nil {
		logger.Errorw("Error publishing state to MQTT", "error", err)
		return
	}
	client.refreshOne(dw)
}

// refreshOne fragt den aktuellen Input (als hexadezimaler Wert) ab und veröffentlicht ihn.
func (client *Client) refreshOne(dw *DisplayWrapper) {
	logger := client.logger.With(zap.String("display", dw.display.Name))
	if dw.display.Refresh != 0 && time.Now().After(dw.lastRefresh.Add(dw.display.Refresh)) {
		response, err := dw.display.Command.GetValue()
		if err != nil {
			logger.Errorw("Error running VCP query command", "error", err, "output", response)
		}
		dw.lastRefresh = time.Now()
		topic := client.stateTopic(dw)
		if err := client.publish(topic, response); err != nil {
			logger.Errorw("Error publishing state to MQTT", "error", err)
		}
		client.setAvailable(dw, err == nil)
	}
}

func (client *Client) setAvailable(dw *DisplayWrapper, available bool) {
	if dw.availableSet && dw.available == available {
		return
	}
	topic := client.availabilityTopic(dw)
	logger := client.logger.With(
		zap.String("display", dw.display.Name),
		zap.String("topic", topic),
		zap.Bool("available", available),
	)
	payload := "0"
	if available {
		payload = "1"
	}
	if err := client.publish(topic, payload); err != nil {
		logger.Errorw("Error publishing availability to MQTT", "error", err)
		return
	}
	dw.available = available
	dw.availableSet = true
	logger.Debugw("Published availability to MQTT")
}

// setAppAvailable markiert die gesamte Anwendung als online.
func (client *Client) setAppAvailable() {
	topic := client.appAvailabilityTopic()
	logger := client.logger.With(zap.String("topic", topic))
	if err := client.publish(topic, "1"); err != nil {
		logger.Errorw("Error publishing application availability to MQTT", "error", err)
	} else {
		logger.Debugw("Published application availability to MQTT")
	}
}

func (client *Client) commandTopic(dw *DisplayWrapper) string {
	return fmt.Sprintf("%s/displays/%s/set", client.appName, dw.display.Name)
}

func (client *Client) stateTopic(dw *DisplayWrapper) string {
	return fmt.Sprintf("%s/displays/%s", client.appName, dw.display.Name)
}

func (client *Client) availabilityTopic(dw *DisplayWrapper) string {
	return fmt.Sprintf("%s/displays/%s/available", client.appName, dw.display.Name)
}

func (client *Client) appAvailabilityTopic() string {
	return fmt.Sprintf("%s/available", client.appName)
}

func (client *Client) syncLog() {
	defer client.logger.Sync()
}

func (client *Client) publish(topic, payload string) error {
	token := client.handle.Publish(topic, 1, true, payload)
	token.Wait()
	return token.Error()
}
