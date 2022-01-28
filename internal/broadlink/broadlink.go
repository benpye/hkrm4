package broadlink

import (
	"fmt"
	"net"
)

const defaultTimeout = 5 // seconds

func NewDevice(ip net.IP, mac net.HardwareAddr, deviceType int) (*Device, error) {
	devChar := isKnownDevice(deviceType)
	if !devChar.supported {
		return nil, fmt.Errorf("device type %v (0x%04x) is not supported", deviceType, deviceType)
	}

	device, err := newDevice(ip, mac, defaultTimeout, devChar)
	if err != nil {
		return nil, err
	}

	return device, nil
}
