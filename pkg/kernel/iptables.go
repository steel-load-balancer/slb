package kernel

import "os/exec"

//TODO - change the adapter so it works out side of Equinix Metal

// EnableMasq This handles the enabling of the MASQUERADING
func EnableMasq(adapter string) error {
	if _, err := exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-o", adapter, "-j", "MASQUERADE").CombinedOutput(); err != nil {
		return err
	}

	return nil
}
