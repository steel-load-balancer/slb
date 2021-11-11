package kernel

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

// ConfigureIPForwarding will examine the kernel configuration and enable IPV4 forwarding if required
func ConfigureIPForwarding() error {
	forwardingFile, err := os.OpenFile("/proc/sys/net/ipv4/ip_forward", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open ip_forward file")
	}
	defer forwardingFile.Close()

	fwdBuffer := make([]byte, 1)
	_, err = forwardingFile.Read(fwdBuffer)
	if err != nil {
		return fmt.Errorf("failed to write: %v", err)
	}

	// Compare current configuration
	if string(fwdBuffer) != "1" {
		_, err = forwardingFile.WriteString("1")
		if err != nil {
			return fmt.Errorf("failed to write: %v", err)
		}
		log.Infoln("Enabled IPv4 Forwarding in the kernel")

	}

	return nil
}

// ConfigureConntrack will examing the kernel configuration and enable connection tracking
func ConfigureConntrack() error {

	contrackFile, err := os.OpenFile("/proc/sys/net/ipv4/vs/conntrack", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open conntrack file")
	}
	defer contrackFile.Close()

	conntrackBuffer := make([]byte, 1)
	_, err = contrackFile.Read(conntrackBuffer)
	if err != nil {
		return fmt.Errorf("failed to write: %v", err)
	}

	// Compare current configuration
	if string(conntrackBuffer) != "1" {
		_, err = contrackFile.WriteString("1")
		if err != nil {
			return fmt.Errorf("failed to write: %v", err)
		}
		log.Infoln("Enabled Connection tracking in the kernel")

	}

	return nil

}
