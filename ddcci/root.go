package ddcci

import (
	"fmt"
	"github.com/d2r2/go-i2c"
)

const (
	I2CAddress                = 0x37
	MonitorBrightnessRegister = 0x10
)

func GetI2CBuses() ([]int, error) {
	var buses []int

	for i := 0; i < 256; i++ {
		// Try to open an I2C connection on this bus
		_, err := i2c.NewI2C(I2CAddress, i)
		if err == nil {
			// If no error occurred, then the bus is available
			buses = append(buses, i)
		}
	}

	if len(buses) == 0 {
		return nil, fmt.Errorf("no available I2C buses found")
	}

	return buses, nil
}

func GetMonitorBrightness(i2cBusNumber int) (byte, error) {
	monitor, err := i2c.NewI2C(I2CAddress, i2cBusNumber)
	if err != nil {
		return 0, fmt.Errorf("error creating I2C connection: %v", err)
	}
	defer monitor.Close()

	ddcCmd := []byte{0x01, MonitorBrightnessRegister}
	_, err = monitor.WriteBytes(ddcCmd)
	if err != nil {
		return 0, fmt.Errorf("error requesting monitor brightness: %v", err)
	}

	// Read the DDC/CI response from the monitor
	readBuf := make([]byte, 12)
	_, err = monitor.ReadBytes(readBuf)
	if err != nil {
		return 0, fmt.Errorf("error reading monitor brightness response: %v", err)
	}

	if readBuf[0] != 0x6E {
		return 0, fmt.Errorf("invalid response from monitor, invalid first byte: %x", readBuf[0])
	}

	if readBuf[4] != MonitorBrightnessRegister {
		return 0, fmt.Errorf("unexpected response from monitor, invalid register %x", readBuf[3])
	}

	return readBuf[9], nil

}

func SetMonitorBrightness(i2cBusNumber int, brightness byte) error {
	monitor, err := i2c.NewI2C(I2CAddress, i2cBusNumber)
	if err != nil {
		return fmt.Errorf("error creating I2C connection: %v", err)
	}
	defer monitor.Close()

	// Create the DDC/CI command to set the brightness
	ddcCmd := []byte{0x51, 0x84, 0x03, MonitorBrightnessRegister, 0x00, brightness}

	// Calculate the checksum
	ddcCmdWithChecksum := AddDDCCIChecksum(ddcCmd)

	// Send the DDC/CI command over I2C
	_, err = monitor.WriteBytes(ddcCmdWithChecksum)
	if err != nil {
		return fmt.Errorf("error setting monitor brightness: %v", err)
	}

	return nil
}

func AddDDCCIChecksum(cmd []byte) []byte {
	// Calculate the checksum by XORing all the data bytes together
	// and then XORing the result with 0x6e.
	var checksum byte
	for _, b := range cmd {
		checksum ^= b
	}
	checksum ^= 0x6e

	// Append the checksum to the command
	return append(cmd, checksum)
}
