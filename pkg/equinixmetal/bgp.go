package equinixmetal

import (
	"os"

	"github.com/kube-vip/kube-vip/pkg/bgp"
	"github.com/packethost/packngo"
	log "github.com/sirupsen/logrus"
)

// FindBGPConfig will return the device neighbours and peering info
func FindBGPConfig(client *packngo.Client, projectID string) (*bgp.Config, error) {
	var b bgp.Config
	thisDevice := findSelf(client, projectID)
	if thisDevice == nil {
		log.Fatalf("Unable to find device/server in Equinix Metal")
	}

	log.Infof("Querying BGP settings for [%s]\n", thisDevice.Hostname)
	neighbours, _, err := client.Devices.ListBGPNeighbors(thisDevice.ID, &packngo.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	if len(neighbours) == 0 {
		log.Fatalf("There are no neighbours being advertised, ensure BGP is enabled for this device")
	}

	b.RouterID = neighbours[0].CustomerIP
	b.AS = uint32(neighbours[0].CustomerAs)

	// Add the peer(s)
	for x := range neighbours[0].PeerIps {
		peer := bgp.Peer{
			Address: neighbours[0].PeerIps[x],
			AS:      uint32(neighbours[0].PeerAs),
		}
		b.Peers = append(b.Peers, peer)
	}
	return &b, nil
}

func findSelf(client *packngo.Client, projectID string) *packngo.Device {
	// Go through devices
	dev, _, _ := client.Devices.List(projectID, &packngo.ListOptions{})
	for _, d := range dev {
		me, _ := os.Hostname()
		if me == d.Hostname {
			return &d
		}
	}
	return nil
}
