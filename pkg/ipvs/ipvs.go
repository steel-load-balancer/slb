package loadbalancer

import (
	"fmt"
	"net"
	"strings"

	"github.com/cloudflare/ipvs"
	log "github.com/sirupsen/logrus"
)

const (
	ROUNDROBIN = "rr"
)

type IPVSLoadBalancer struct {
	client              ipvs.Client
	loadBalancerService ipvs.Service
	Port                int
}

func NewIPVSLB(address string, port int) (*IPVSLoadBalancer, error) {
	// Create IPVS client
	c, err := ipvs.New()
	if err != nil {
		return nil, fmt.Errorf("error creating IPVS client: %v", err)
	}

	// Generate out API Server LoadBalancer instance
	svc := ipvs.Service{
		Family:    ipvs.INET,
		Protocol:  ipvs.TCP,
		Port:      uint16(port),
		Address:   ipvs.NewIP(net.ParseIP(address)),
		Scheduler: ROUNDROBIN,
	}
	err = c.CreateService(svc)
	// If we've an error it could be that the IPVS lb instance has been left from a previous leadership
	if err != nil && strings.Contains(err.Error(), "file exists") {
		log.Warnf("load balancer for API server already exists, attempting to remove and re-create")
		err = c.RemoveService(svc)
		if err != nil {
			return nil, fmt.Errorf("error re-creating IPVS service: %v", err)
		}
		err = c.CreateService(svc)
		if err != nil {
			return nil, fmt.Errorf("error re-creating IPVS service: %v", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("error creating IPVS service: %v", err)
	}

	lb := &IPVSLoadBalancer{
		Port:                port,
		client:              c,
		loadBalancerService: svc,
	}
	// Return our created load-balancer
	return lb, nil
}

func (lb *IPVSLoadBalancer) RemoveIPVSLB() error {
	err := lb.client.RemoveService(lb.loadBalancerService)
	if err != nil {
		return fmt.Errorf("error removing existing IPVS service: %v", err)
	}
	return nil

}

func (lb *IPVSLoadBalancer) AddBackend(address, fwdtype string, port int) error {
	var forwardingMethod ipvs.ForwardType
	switch strings.ToUpper(fwdtype) {
	case "MASQ":
		forwardingMethod = ipvs.Masquarade
	case "LOCAL":
		forwardingMethod = ipvs.Local
	}
	dst := ipvs.Destination{
		Address:   ipvs.NewIP(net.ParseIP(address)),
		Port:      uint16(port),
		Family:    ipvs.INET,
		Weight:    1,
		FwdMethod: forwardingMethod,
	}

	err := lb.client.CreateDestination(lb.loadBalancerService, dst)
	// Swallow error of existing back end, the node watcher may attempt to apply
	// the same back end multiple times
	if err != nil && !strings.Contains(err.Error(), "file exists") {
		return fmt.Errorf("error creating backend: %v", err)
	}
	return nil
}

func (lb *IPVSLoadBalancer) RemoveBackend(address string, port int) error {
	dst := ipvs.Destination{
		Address: ipvs.NewIP(net.ParseIP(address)),
		Port:    uint16(port),
		Family:  ipvs.INET,
		Weight:  1,
	}
	err := lb.client.RemoveDestination(lb.loadBalancerService, dst)
	if err != nil {
		return fmt.Errorf("error removing backend: %v", err)
	}
	return nil
}
