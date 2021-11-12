package kernel

import (
	iptables "github.com/coreos/go-iptables/iptables"
	log "github.com/sirupsen/logrus"
)

//TODO - change the adapter so it works out side of Equinix Metal

// EnableMasq This handles the enabling of the MASQUERADING
func EnableMasq(adapter string) error {

	i, err := iptables.New()
	if err != nil {
		log.Fatalf("error creating iptables client : %s", err.Error())
	}

	ruleExists, err := i.Exists("nat", "POSTROUTING", "-o", adapter, "-j", "MASQUERADE")
	if err != nil {
		log.Fatalf("Unable to verify MASQUERADING iptables configuration: %s", err.Error())
	}

	if !ruleExists {
		err = i.Append("nat", "POSTROUTING", "-o", adapter, "-j", "MASQUERADE")
		if err != nil {
			log.Fatalf("Unable to verify MASQUERADING iptables configuration: %s", err.Error())
		}
		log.Info("Created iptables Masquerading rule for [%s]", adapter)
	}
	return nil
}
