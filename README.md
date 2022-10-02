# MQTT to command-line applications gateway

Create virtual MQTT switches from command-line applications. Expose apps running locally on your laptop or desktop to
your home automation server, like [Home Assistant](https://home-assistant.io). 

# Installation

## MacOS

```bash
# Add tap to your Homebrew
$ brew tap haimgel/tools

# Install it
$ brew install mqtt2cmd

# Configure (see below for details)
$ mkdir -p ~/Library/Application\ Support/mqtt2cmd
$ vi ~/Library/Application\ Support/mqtt2cmd/config.yaml

# Run it in the background, autostart upon boot
$ brew services start mqtt2cmd

# View the logs
$ tail -f ~/Library/Application\ Support/mqtt2cmd/mqtt2cmd.log
```

# Configuration

This application expects a configuration file named `config.yaml`, located in:
* `$HOME/Library/Application Support/mqtt2cmd` on MacOS
* `$XDG_CONFIG_HOME/mqtt2cmd` or `$HOME/.config/mqtt2cmd` on Linux

Sample configuration (controls Slack status across multiple Slack workspaces using [slack_status](https://github.com/haimgel/slack_status))
```yaml
mqtt:
  broker: "tcp://your-mqtt-server-address:1883"
switches:
  - name: lunch
    # How often to run the `get_state` command and update the state: useful if the state changes by means
    # other than this application
    refresh: "10m"
    # A command to turn the switch on
    turn_on: "slack_status lunch"
    # A command to turn the switch off
    turn_off: "slack_status clear"
    # A command to query the state of the switch: exit status = 0 is "ON", exit status = 1 is "OFF"
    get_state: "slack_status --get lunch"
```

Using the configuration above, `mqtt2cmd` will subscribe to MQTT topic `mqtt2cmd/switches/lunch/set` and will
publish the current state to `mqtt2cmd/switches/lunch`
