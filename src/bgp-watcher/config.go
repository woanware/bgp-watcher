package main

import (
	"log"

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
	TargetAs         int
	NeighbourPeers   []string
	Prefixes         []string
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

	config.DatabaseServer = confReader.GetString("database_server")
	config.DatabasePort = confReader.GetInt("database_port")
	config.DatabaseUsername = confReader.GetString("database_username")
	config.DatabasePassword = confReader.GetString("database_password")
	config.Database = confReader.GetString("database")
	config.HistoryMonths = confReader.GetInt("history_months")
	config.Processes = confReader.GetInt("processes")
	config.NeighbourPeers = confReader.GetStringSlice("neighbour_peers")
	config.Prefixes = confReader.GetStringSlice("prefixes")

	return config
}
