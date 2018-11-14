package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/jackc/pgx"
	flags "github.com/jessevdk/go-flags"
	viper "github.com/spf13/viper"
)

// ##### Constants #####################################################################################################

const APP_NAME string = "bgp-monitor (bgpm)"
const APP_VERSION string = "0.0.1"

//const RIPE_UPDATES string = "http://data.ris.ripe.net/rrc00/"
//const RIPE_UPDATES string = "http://data.ris.ripe.net/rrc11/"
const RIPE_UPDATES string = "http://data.ris.ripe.net/rrc14/"
const HISTORY_MONTHS int = 6

// ##### Variables #####################################################################################################

var (
	configReader *viper.Viper
	config       *Config
	db           *pgx.Conn
	options      Options
	asNames      *AsNames
)

// ##### Methods ##############################################################

//
func main() {

	fmt.Println(fmt.Sprintf("\n%s v%s - woanware\n", APP_NAME, APP_VERSION))

	parseCommandLine()
	initialiseConfiguration()
	config = parseConfiguration()
	configureDatabase()

	// for name, url := range config.DataSets {
	// 	fmt.Println(name)
	// 	fmt.Println(url)
	// }

	// return

	asNames = NewAsNames()
	err := asNames.Update()
	if err != nil {
		fmt.Printf("Error downloading AS data: %v\n", err)
		return
	}

	// for number, a := range asNames.names {
	// 	fmt.Printf("%v\n", number)
	// 	fmt.Printf("%v\n", a.Country)
	// 	fmt.Printf("%v\n", a.Name)
	// 	fmt.Printf("%v\n--------------------------------\n", a.Description)
	// 	//fmt.Printf("%v\npgx", a)
	// }

	h, err := NewHistory(config.DataSets, config.HistoryMonths, config.Processes)
	if err != nil {
		return
	}
	h.Update()

	// return

	detector := NewDetector(asNames)
	for as := range config.TargetAs {
		detector.AddTargetAs(as)
	}
	for _, prefix := range config.Prefixes {
		detector.AddPrefix(prefix)
	}

	monitor := NewMonitor(detector, config.Processes)
	monitor.Start()

	// Stop application from exiting
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}

//
func parseCommandLine() {

	var parser = flags.NewParser(&options, flags.Default)
	if _, err := parser.Parse(); err != nil {
		fmt.Printf("%v\n", err)
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}
}

//
func configureDatabase() {

	c, err := pgx.ParseDSN(fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		config.DatabaseServer, config.DatabasePort, config.Database, config.DatabaseUsername, config.DatabasePassword))

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing database config: %v\n", err)
		os.Exit(1)
	}

	db, err = pgx.Connect(c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
		os.Exit(1)
	}
}

//
func reloadConfig() {

	config = parseConfiguration()
}
