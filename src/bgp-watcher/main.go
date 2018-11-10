package main

import (
	"fmt"
	"log"
	"os"

	flags "github.com/jessevdk/go-flags"
	"github.com/spf13/viper"
)

// ##### Constants #####################################################################################################

// App Constants
const APP_NAME string = "bgp-monitor (bgpm)"
const APP_VERSION string = "0.0.1"
const RIPE_UPDATES string = "http://data.ris.ripe.net/rrc00/"
const HISTORY_MONTHS int = 6

// ##### Variables #####################################################################################################

var (
	config  *Config
	options Options
)

// ##### Methods ##############################################################

//
func main() {

	fmt.Println(fmt.Sprintf("\n%s v%s - woanware\n", APP_NAME, APP_VERSION))

	var parser = flags.NewParser(&options, flags.Default)
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	config = new(Config)
	if loadConfig() == false {
		return
	}

	h, err := NewHistory(config.HistoryMonths, config.Processes)
	if err != nil {
		return
	}
	h.Update()
}

//
func loadConfig() bool {

	confReader := viper.New()
	confReader.SetConfigName("bgpm")
	confReader.AddConfigPath(".")
	err := confReader.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file: %s \n", err)
	}

	config.HistoryMonths = confReader.GetInt("history_months")
	config.Processes = confReader.GetInt("processes")

	return true
}
