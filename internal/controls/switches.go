package controls

import (
	"fmt"
	"os/exec"
	"time"
)

// VCPCommand repräsentiert einen VCP-Befehl für den Input Source.
type VCPCommand struct {
	Name            string        `mapstructure:"name"`      // z.B. "HDMI"
	SetCmd          string        `mapstructure:"set_value"` // z.B. "ddcutil setvcp 0x60 %s"
	GetCmd          string        `mapstructure:"get_state"` // z.B. "ddcutil getvcp 0x60"
	RefreshInterval time.Duration `mapstructure:"refresh"`   // Optional, kann überschrieben werden
}

// SetValue setzt einen Wert mithilfe des Set-Command-Templates.
// Dabei wird der übergebene String (z.B. "0x11") eingesetzt.
func (vc *VCPCommand) SetValue(value string) (string, error) {
	if vc.SetCmd != "" {
		cmd := fmt.Sprintf(vc.SetCmd, value)
		return run(cmd)
	}
	return "", fmt.Errorf("set_value command not defined for VCP command %s", vc.Name)
}

// GetValue führt den Get-Befehl aus und gibt den aktuellen Wert als String zurück.
func (vc *VCPCommand) GetValue() (string, error) {
	if vc.GetCmd != "" {
		return run(vc.GetCmd)
	}
	return "", fmt.Errorf("get_state command not defined for VCP command %s", vc.Name)
}

// run führt einen Shell-Befehl aus und gibt den kombinierten Output zurück.
func run(command string) (string, error) {
	cmd := exec.Command("/bin/sh", "-c", command)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// Display repräsentiert ein Hardware-Display mit einem bestimmten Input.
type Display struct {
	Name    string        `mapstructure:"name"`    // z.B. "Monitor 1"
	Refresh time.Duration `mapstructure:"refresh"` // z.B. "1m"
	Command VCPCommand    `mapstructure:"command"` // VCP-Befehl für den Input Source
}
