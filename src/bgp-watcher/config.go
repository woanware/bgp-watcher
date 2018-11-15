package main

import (
	"log"
	"strings"

	fsnotify "github.com/fsnotify/fsnotify"
	bgp "github.com/osrg/gobgp/packet"
	viper "github.com/spf13/viper"
)

// ##### Structs ##############################################################

type DataSet struct {
	Name string `mapstructure:"name"`
	Url  string `mapstructure:"url"`
}

type DataSets struct {
	Data []DataSet `mapstructure:"data_sets"`
}

// Config holds configuration data for the application
type Config struct {
	DatabaseServer      string
	DatabasePort        int
	DatabaseUsername    string
	DatabasePassword    string
	Database            string
	HistoryMonths       int
	Processes           int
	DataSets            map[string]string
	MonitorCountryCodes map[string]struct{}
	TargetAs            map[uint32]struct{}
	NeighbourPeers      map[uint32]struct{}
	Prefixes            []*bgp.IPAddrPrefix
}

// ##### Methods ##############################################################

// LoadConfig loads the configuration data from the "bgpm" config file
func initialiseConfiguration() {

	configReader = viper.New()
	configReader.SetConfigName("bgpm")
	configReader.AddConfigPath(".")
	err := configReader.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file: %s \n", err)
	}

	viper.OnConfigChange(func(e fsnotify.Event) {
		reloadConfig()
	})

}

func parseConfiguration() *Config {

	config := new(Config)

	config.DataSets = make(map[string]string)
	config.MonitorCountryCodes = make(map[string]struct{})
	config.TargetAs = make(map[uint32]struct{})
	config.NeighbourPeers = make(map[uint32]struct{})
	config.Prefixes = make([]*bgp.IPAddrPrefix, 0)

	config.DatabaseServer = configReader.GetString("database_server")
	config.DatabasePort = configReader.GetInt("database_port")
	config.DatabaseUsername = configReader.GetString("database_username")
	config.DatabasePassword = configReader.GetString("database_password")
	config.Database = configReader.GetString("database")
	config.HistoryMonths = configReader.GetInt("history_months")
	config.Processes = configReader.GetInt("processes")

	var as uint32
	var err error

	// Convert string slice values (Target AS's) into uint32
	temp := configReader.GetStringSlice("target_as")
	for _, t := range temp {
		as, err = ConvertStringToUint32(t)
		if err != nil {
			log.Fatalf("Invalid AS: %s\n", as)
		}

		config.TargetAs[as] = struct{}{}
	}

	// Convert string slice values (Neighbour Peers) into uint32
	temp = configReader.GetStringSlice("neighbour_peers")
	for _, t := range temp {
		as, err = ConvertStringToUint32(t)
		if err != nil {
			log.Fatalf("Invalid AS: %s\n", as)
		}

		config.NeighbourPeers[as] = struct{}{}
	}

	temp = configReader.GetStringSlice("prefixes")
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

		if bit > 32 {
			log.Fatalf("Invalid prefix bit: %s\n", parts[1])
		}

		config.Prefixes = append(config.Prefixes, bgp.NewIPAddrPrefix(bit, parts[0]))
	}

	// Convert string slice values (Target AS's) into uint32
	temp = configReader.GetStringSlice("monitor_country_codes")
	for _, t := range temp {
		config.MonitorCountryCodes[t] = struct{}{}
	}

	// Decode the data set info (name, URL)
	var dataSets DataSets
	err = configReader.Unmarshal(&dataSets)
	if err != nil {
		log.Fatalf("Erorr decoding config data sets: %v\n", err)
	}

	// Move the JSON data into our config
	for _, ds := range dataSets.Data {

		// Lets be nice and make sure that our URL's are consistent
		if strings.HasSuffix(ds.Url, "/") == false {
			ds.Url = ds.Url + "/"
		}

		config.DataSets[ds.Name] = ds.Url
	}

	return config
}
