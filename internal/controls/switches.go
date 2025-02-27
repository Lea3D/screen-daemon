package controls

import (
	"os/exec"
	"time"
)

// Switch struct represents a controllable switch with associated commands.
type Switch struct {
	Name            string        `mapstructure:"name"`      // Name of the switch
	OnCmd           string        `mapstructure:"turn_on"`   // Command to turn the switch on
	OffCmd          string        `mapstructure:"turn_off"`  // Command to turn the switch off
	StateCmd        string        `mapstructure:"get_state"` // Command to get the current state of the switch
	ToggleCmd       string        `mapstructure:"toggle"`    // Command to toggle the switch
	RefreshInterval time.Duration `mapstructure:"refresh"`   // Interval to refresh the state of the switch
}

// SwitchOnOff executes the appropriate command to switch the device on or off.
func (sw *Switch) SwitchOnOff(state bool) (string, error) {
	if state {
		return run(sw.OnCmd) // Run the command to turn the switch on
	} else {
		return run(sw.OffCmd) // Run the command to turn the switch off
	}
}

// Toggle executes the toggle command if defined.
func (sw *Switch) Toggle() (string, error) {
	if sw.ToggleCmd != "" {
		return run(sw.ToggleCmd) // Run the command to toggle the switch
	}
	return "", nil
}

// GetState executes the state command and determines if the switch is on or off.
func (sw *Switch) GetState() (bool, string, error) {
	out, err := run(sw.StateCmd)
	if exitError, ok := err.(*exec.ExitError); ok {
		if exitError.ExitCode() == 1 {
			return false, out, nil // If exit code is 1, assume switch is off
		} else {
			return false, out, err // Other exit codes indicate an error
		}
	} else {
		return err == nil, out, err // If no error, assume the switch is on
	}
}

// run executes a shell command and returns its output.
func run(command string) (string, error) {
	cmd := exec.Command("/bin/sh", "-c", command) // Execute command in shell
	out, err := cmd.CombinedOutput()              // Capture combined stdout and stderr output
	return string(out), err
}
