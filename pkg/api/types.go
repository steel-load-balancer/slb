package api

import (
	"github.com/kube-vip/kube-vip/pkg/vip"
	loadbalancer "github.com/thebsdbox/slb/pkg/ipvs"
)

// LoadBalancer is a loadBalancer instance
type loadBalancer struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`

	vip vip.Network

	EIP  string `json:"eip"`
	Port int    `json:"port"`

	ipvs    *loadbalancer.IPVSLoadBalancer
	Backend []backends `json:"backends"`
}

type backends struct {
	UUID string `json:"uuid"`
	IP   string `json:"ip"`
	Port int    `json:"port"`
}
