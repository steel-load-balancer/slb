package api

import (
	"fmt"

	"github.com/kube-vip/kube-vip/pkg/bgp"
	"github.com/packethost/packngo"
)

type Config struct {
	// Layer 2/3
	Arp bool
	Bgp bool

	// Device to bind to
	Adapter string

	// API Port
	Port int

	// IPAM Range
	IpamRange string

	// Provider Specific
	EquinixMetal        bool
	ProjectID, Facility string
	emClient            *packngo.Client
}

type Manager struct {
	// SLB configuration
	SSLBConfig Config

	// Networking configuration
	bgp *bgp.Server

	// LoadBalancers  holds ALL local loadbalancers
	LoadBalancers []loadBalancer
}

func NewManager(arp, bgp bool, adapter string) (*Manager, error) {

	if arp && bgp {
		return nil, fmt.Errorf("both layer2 and layer3 can't be active at the same time")
	}
	var m Manager
	m.SSLBConfig.Arp = arp
	m.SSLBConfig.Bgp = bgp
	m.SSLBConfig.Adapter = adapter
	return &m, nil
}

// InitEquinixMetal will use the Equinix Metal API to populate configuration details
func (m *Manager) InitEquinixMetal() error {
	var err error
	m.SSLBConfig.emClient, err = packngo.NewClient()
	if err != nil {
		return err
	}
	return nil
}
