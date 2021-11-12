package kernel

import (
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

// ConfigureIPForwarding will examine the kernel configuration and enable IPV4 forwarding if required
func ConfigureIPForwarding() error {
	err := ioutil.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte{'1', '\n'}, 0644)
	if err != nil {
		return fmt.Errorf("failed to write to /proc/sys/net/ipv4/vs/conntrack: %v", err)
	}
	log.Infoln("Enabled IPv4 Forwarding in the kernel")
	return nil
}

// ConfigureConntrack will examing the kernel configuration and enable connection tracking
func ConfigureConntrack() error {
	err := ioutil.WriteFile("/proc/sys/net/ipv4/vs/conntrack", []byte{'1', '\n'}, 0644)
	if err != nil {
		return fmt.Errorf("failed to write to /proc/sys/net/ipv4/vs/conntrack: %v", err)
	}
	log.Infoln("Enabled IPv4 Virtual Server connection tracking in the kernel")
	return nil

}
