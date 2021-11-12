package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/kube-vip/kube-vip/pkg/bgp"
	"github.com/kube-vip/kube-vip/pkg/vip"
	log "github.com/sirupsen/logrus"
	"github.com/thebsdbox/slb/pkg/equinixmetal"
	"github.com/thebsdbox/slb/pkg/ipam"
	loadbalancer "github.com/thebsdbox/slb/pkg/ipvs"
)

// Start, brings up the API server
func (m *Manager) Start() {
	if m.SSLBConfig.EquinixMetal {
		err := m.InitEquinixMetal()
		if err != nil {
			log.Fatal(err)
		}
	}
	if m.SSLBConfig.Bgp {
		if m.SSLBConfig.EquinixMetal {
			c, err := equinixmetal.FindBGPConfig(m.SSLBConfig.emClient, m.SSLBConfig.ProjectID)
			if err != nil {
				log.Fatal(err)
			}
			m.bgp, err = bgp.NewBGPServer(c)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	// start API server, this is blocking
	m.apiServerStart()
}

func (m *Manager) apiServerStart() {
	// creates a new instance of a mux router
	myRouter := mux.NewRouter().StrictSlash(true)
	// replace http.HandleFunc with myRouter.HandleFunc

	myRouter.HandleFunc("/loadbalancer", m.createNewLoadBalancer).Methods("POST")
	myRouter.HandleFunc("/loadbalancers", m.returnAllLoadBalancers)
	myRouter.HandleFunc("/loadbalancer/{uuid}", m.returnSingleLoadBalancer).Methods("GET")
	myRouter.HandleFunc("/loadbalancer/{uuid}", m.deleteLoadBalancer).Methods("DELETE")
	myRouter.HandleFunc("/loadbalancer/{uuid}/backend", m.createNewBackend).Methods("POST")
	myRouter.HandleFunc("/loadbalancer/{uuid}/backend/{backenduuid}", m.deleteBackend).Methods("DELETE")

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", m.SSLBConfig.Port), myRouter))
}

func (m *Manager) createNewLoadBalancer(w http.ResponseWriter, r *http.Request) {
	var err error
	reqBody, _ := ioutil.ReadAll(r.Body)
	var newLB loadBalancer
	json.Unmarshal(reqBody, &newLB)

	// we will need a unique UUID first, then we will need an EIP
	newLB.UUID = uuid.NewString()

	// Determine if load-balancer address comes from Equinix Metal API or internal IPAM
	if m.SSLBConfig.EquinixMetal {
		newLB.EIP, err = equinixmetal.GetEIP(m.SSLBConfig.emClient, m.SSLBConfig.ProjectID, m.SSLBConfig.Facility)
		if err != nil {
			log.Error(err)
		}
	} else {
		newLB.EIP, err = ipam.FindAvailableHostFromRange("", m.SSLBConfig.IpamRange)
		if err != nil {
			log.Error(err)
		}
	}

	if err != nil {
		log.Error(err)
	}
	newLB.vip, err = vip.NewConfig(newLB.EIP, m.SSLBConfig.Adapter, false)
	if err != nil {
		log.Error(err)
	}
	err = newLB.vip.AddIP()
	if err != nil {
		log.Error(err)
	}

	if m.SSLBConfig.Arp {
		err = vip.ARPSendGratuitous(newLB.EIP, m.SSLBConfig.Adapter)
		if err != nil {
			log.Error(err)
		}
	}

	if m.SSLBConfig.Bgp {
		err = m.bgp.AddHost(fmt.Sprintf("%s/32", newLB.EIP))
		if err != nil {
			log.Error(err)
		}
	}
	newLB.ipvs, err = loadbalancer.NewIPVSLB(newLB.EIP, newLB.Port)
	if err != nil {
		log.Error(err)
	}

	m.LoadBalancers = append(m.LoadBalancers, newLB)
	log.Infof("Created new Loadbalancer IP [%s] UUID [%s]", newLB.EIP, newLB.UUID)
}

func (m *Manager) returnAllLoadBalancers(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(m.LoadBalancers)
}

func (m *Manager) returnSingleLoadBalancer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["uuid"]

	for _, loadbalancer := range m.LoadBalancers {
		if loadbalancer.UUID == key {
			json.NewEncoder(w).Encode(loadbalancer)
		}
	}
}

func (m *Manager) createNewBackend(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["uuid"]

	for i, loadbalancer := range m.LoadBalancers {
		if loadbalancer.UUID == key {
			reqBody, _ := ioutil.ReadAll(r.Body)
			fmt.Println(string(reqBody))
			var backend backends
			backend.UUID = uuid.NewString()
			err := json.Unmarshal(reqBody, &backend)
			if err != nil {
				log.Error(err)
			}
			// Default to MASQUERADING
			if backend.Type == "" {
				backend.Type = "MASQ"
			}
			loadbalancer.ipvs.AddBackend(backend.IP, backend.Type, backend.Port)
			m.LoadBalancers[i].Backend = append(loadbalancer.Backend, backend)

			log.Infof("Created new backend for IP [%s] w/IP [%s] UUID [%s]", loadbalancer.EIP, backend.IP, backend.UUID)
		}
	}
}

func (m *Manager) deleteBackend(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["uuid"]
	backendKey := vars["backenduui"]

	for _, loadbalancer := range m.LoadBalancers {
		if loadbalancer.UUID == key {
			for i, backend := range loadbalancer.Backend {
				if backend.UUID == backendKey {
					loadbalancer.ipvs.RemoveBackend(backend.IP, backend.Port)
					loadbalancer.Backend = append(loadbalancer.Backend[:i], loadbalancer.Backend[i+1:]...)

				}

			}

		}
	}
}

func (m *Manager) deleteLoadBalancer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["uuid"]

	for index, loadbalancer := range m.LoadBalancers {
		if loadbalancer.UUID == key {
			if m.SSLBConfig.EquinixMetal {
				err := equinixmetal.DelEIP(m.SSLBConfig.emClient, m.SSLBConfig.ProjectID, loadbalancer.EIP)
				if err != nil {
					log.Error(err)
				}
			}
			err := loadbalancer.ipvs.RemoveIPVSLB()
			if err != nil {
				log.Error(err)
			}
			if m.SSLBConfig.Bgp {
				err = m.bgp.DelHost(fmt.Sprintf("%s/32", loadbalancer.EIP))
				if err != nil {
					log.Error(err)
				}
			}
			err = loadbalancer.vip.DeleteIP()
			if err != nil {
				log.Error(err)
			}
			m.LoadBalancers = append(m.LoadBalancers[:index], m.LoadBalancers[index+1:]...)
		}
	}
}
