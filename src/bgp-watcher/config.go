package main

import (
	"log"
	"strings"

	bgp "github.com/osrg/gobgp/packet"
	"github.com/spf13/viper"
)

// Config holds configuration data for the application
type Config struct {
	DatabaseServer   string
	DatabasePort     int
	DatabaseUsername string
	DatabasePassword string
	Database         string
	HistoryMonths    int
	Processes        int
	TargetAs         map[uint32]struct{}
	NeighbourPeers   map[uint32]struct{}
	Prefixes         []*bgp.IPAddrPrefix
}

// LoadConfig loads the configuration data from the "bgpm" config file
func LoadConfig() *Config {

	confReader := viper.New()
	confReader.SetConfigName("bgpm")
	confReader.AddConfigPath(".")
	err := confReader.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file: %s \n", err)
	}

	config := new(Config)

	config.TargetAs = make(map[uint32]struct{}, 0)
	config.NeighbourPeers = make(map[uint32]struct{}, 0)
	config.Prefixes = make([]*bgp.IPAddrPrefix, 0)
	config.DatabaseServer = confReader.GetString("database_server")
	config.DatabasePort = confReader.GetInt("database_port")
	config.DatabaseUsername = confReader.GetString("database_username")
	config.DatabasePassword = confReader.GetString("database_password")
	config.Database = confReader.GetString("database")
	config.HistoryMonths = confReader.GetInt("history_months")
	config.Processes = confReader.GetInt("processes")

	var as uint32

	// Convert string slice values (Target AS's) into uint32
	temp := confReader.GetStringSlice("target_as")
	for _, t := range temp {
		as, err = ConvertStringToUint32(t)
		if err != nil {
			log.Fatalf("Invalid AS: %s\n", as)
		}

		config.TargetAs[as] = struct{}{}
	}

	// Convert string slice values (Neighbour Peers) into uint32
	temp = confReader.GetStringSlice("neighbour_peers")
	for _, t := range temp {
		as, err = ConvertStringToUint32(t)
		if err != nil {
			log.Fatalf("Invalid AS: %s\n", as)
		}

		config.NeighbourPeers[as] = struct{}{}
	}

	temp = confReader.GetStringSlice("prefixes")
	// Convert string slice values (Prefixes) into IPAddrPrefix (from bgp lib)
	var parts []string
	var bit uint8
	for _, t := range temp {

		parts = strings.Split(t, "/")
		if len(parts) != 2 {
			log.Fatalf("Invalid prefix: %s\n", t)
		}

		bit, err = ConvertStringToUint8(parts[1])
		if err != nil {
			log.Fatalf("Invalid prefix bit: %s\n", parts[1])
		}

		config.Prefixes = append(config.Prefixes, bgp.NewIPAddrPrefix(bit, parts[0]))
	}

	// func NewIPAddrPrefix(length uint8, prefix string) *IPAddrPrefix {
	// 	return &IPAddrPrefix{
	// 		IPAddrPrefixDefault{length, net.ParseIP(prefix).To4()},
	// 		4,
	// 	}
	// }

	return config
}
