# MQTT to Command-line Applications Gateway

[![Release](https://img.shields.io/github/release/haimgel/mqtt2cmd.svg?style=flat)](https://github.com/haimgel/mqtt2cmd/releases/latest)  
[![Software license](https://img.shields.io/github/license/haimgel/mqtt2cmd.svg?style=flat)](/LICENSE)  
[![Build status](https://img.shields.io/github/actions/workflow/status/haimgel/mqtt2cmd/release.yaml?style=flat)](https://github.com/haimgel/mqtt2cmd/actions?workflow=release)

Create virtual MQTT-controlled command-line gateways for hardware displays. Expose your locally running applications or devices (e.g. monitor input sources via ddcutil) to your home automation server such as [Home Assistant](https://home-assistant.io).

## ## Installation

Follow the standard installation instructions to set up the application and configure it to run as a Docker container or directly on your system.

## Configuration

This application expects a configuration file named `config.yaml`, located in:

- `$XDG_CONFIG_HOME/mqtt2cmd` or `$HOME/.config/mqtt2cmd` on Linux

## Sample Configuration

Below is an example configuration that sets up separate MQTT data points for each monitor input from 0x01 to 0x12. The `get_state` command is defined once per display (since ddcutil returns the currently active input), and each input has its fixed `set_value` command:

```yaml
app-id: "screendaemon"
mqtt:
  broker: "tcp://your-mqtt-server-address:1883"
  username: "mqtt-user"
  password: "your-mqtt-password"

displays:
  - name: "Monitor 1"
    refresh: "1m"
    get_state: "ddcutil getvcp 0x60"
    inputs:
      - name: "VGA-1"
        set_value: "ddcutil setvcp 0x60 0x01"
      - name: "VGA-2"
        set_value: "ddcutil setvcp 0x60 0x02"
      - name: "HDMI-1"
        set_value: "ddcutil setvcp 0x60 0x11"
      - name: "HDMI-2"
        set_value: "ddcutil setvcp 0x60 0x12"

log:
  path: "/workspace/internal/config/logs/screendaemon.log"
```

## How It Works

1. Subscribes to MQTT topics with a prefix based on `app-id` for each input.
2. Publishes the current input (queried via the shared `get_state` command) to a topic like `screendaemon/displays/Monitor 1`.
3. Publishes overall application availability to `screendaemon/available` and per-display availability to `screendaemon/displays/Monitor 1/available`.
4. Each input's fixed `set_value` command is executed when a corresponding MQTT command is received.

## Sample Home Assistant Configuration

```yaml
mqtt:
  switch:
    - name: "Monitor 1 HDMI-1"
      state_topic: "screendaemon/displays/Monitor 1"
      command_topic: "screendaemon/displays/Monitor 1/set"
      payload_on: "0x11"
      availability:
        - topic: "screendaemon/available"
        - topic: "screendaemon/displays/Monitor 1/available"
      availability_mode: "all"
```

## Docker Run Example

```bash
docker run --privileged -v /srv/docker/appdata/screendaemon/config:/workspace/internal/config -v /dev:/dev screen-daemon
```

This command mounts the configuration and device files correctly.

## Conclusion

This updated README reflects the new concept: Each display uses a shared `get_state` command (since ddcutil always returns the active input) and separate, fixed `set_value` commands for each input—ideal for integration into Home Assistant.
