package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/thebsdbox/slb/pkg/api"
	"github.com/thebsdbox/slb/pkg/kernel"
)

var m api.Manager

// Release - this struct contains the release information populated when building slb
var Release struct {
	Version string
	Build   string
}

var slbCmd = &cobra.Command{
	Use:   "slb",
	Short: "This is a server for providing load balancer services for metal (steel)",
}

func init() {
	slbServer.PersistentFlags().StringVar(&m.SSLBConfig.Adapter, "adapter", "", "Adapter to bind Load-Balancer VIPs too")
	slbServer.PersistentFlags().IntVar(&m.SSLBConfig.Port, "port", 10001, "API Port")
	slbServer.PersistentFlags().BoolVar(&m.SSLBConfig.Arp, "arp", false, "Enable ARP for SLB")
	slbServer.PersistentFlags().BoolVar(&m.SSLBConfig.Arp, "bgp", false, "Enable BGP for SLB")

	slbServer.PersistentFlags().StringVar(&m.SSLBConfig.IpamRange, "ipamRange", "", "Range of IP addresses to use for IPAM")

	// vendor specific
	slbServer.PersistentFlags().BoolVar(&m.SSLBConfig.EquinixMetal, "equinixMetal", false, "Query the Equinix Metal API for BGP information")
	slbServer.PersistentFlags().StringVar(&m.SSLBConfig.ProjectID, "project", "", "a project uuid")
	slbServer.PersistentFlags().StringVar(&m.SSLBConfig.Facility, "facility", "", "a facility uuid")

	slbCmd.AddCommand(slbServer)
	slbCmd.AddCommand(slbVersion)
}

// Execute - starts the command parsing process
func Execute() {
	if err := slbCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

var slbServer = &cobra.Command{
	Use:   "server",
	Short: "Start the Steel Load-Balancer server",
	Run: func(cmd *cobra.Command, args []string) {
		log.Infoln("Starting the Steel Load-Balancer")
		log.Infof("API Server will be exposed on [0.0.0.0:%d]", m.SSLBConfig.Port)
		err := kernel.EnableMasq(m.SSLBConfig.Adapter)
		if err != nil {
			log.Fatal(err)
		}
		err = kernel.ConfigureIPForwarding()
		if err != nil {
			log.Fatal(err)
		}
		err = kernel.ConfigureConntrack()
		if err != nil {
			log.Fatal(err)
		}
		m.Start()
	},
}

var slbVersion = &cobra.Command{
	Use:   "version",
	Short: "Version and Release information about the SLB",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Steel Load-Balancer Release Information\n")
		fmt.Printf("Version:  %s\n", Release.Version)
		fmt.Printf("Build:    %s\n", Release.Build)
	},
}
