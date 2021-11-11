package equinixmetal

import (
	"fmt"

	"github.com/packethost/packngo"
)

func GetEIP(client *packngo.Client, project, facility string) (string, error) {
	req := packngo.IPReservationRequest{
		Type:        "public_ipv4",
		Quantity:    1,
		Description: "vippy loadbalancer EIP",
		Facility:    &facility,

		FailOnApprovalRequired: true,
	}

	ipReservation, _, err := client.ProjectIPs.Request(project, &req)
	if err != nil {
		return "", fmt.Errorf("failed to request an IP for the load balancer: %v", err)
	}
	return ipReservation.Address, nil
}

func DelEIP(client *packngo.Client, project, address string) error {
	ips, _, err := client.ProjectIPs.List(project, &packngo.ListOptions{})
	if err != nil {
		return err
	}
	for x := range ips {
		if ips[x].Address == address {
			_, err = client.ProjectIPs.Remove(ips[x].ID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
