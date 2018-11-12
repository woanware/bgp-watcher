package main

import (
	"fmt"
	"os"

	"github.com/jackc/pgx"
	flags "github.com/jessevdk/go-flags"
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
	db      *pgx.Conn
	options Options
	asNames *AsNames
)

// ##### Methods ##############################################################

//
func main() {

	// asns := make(map[uint32]map[string]uint64)
	// if asns[1] == nil {
	// 	asns[1] = make(map[string]uint64)
	// }
	// asns[1]["abc"]++
	// asns[1]["abc"]++
	// asns[1]["efg"]++
	// if asns[2] == nil {
	// 	asns[2] = make(map[string]uint64)
	// }
	// asns[2]["zxc"]++
	// asns[2]["zxc"]++
	// asns[2]["zxc"]++
	// asns[2]["zxc"]++

	// for peer, a := range asns {
	// 	fmt.Println(peer)
	// 	//fmt.Println(v)
	// 	//fmt.Println(asns[k])

	// 	for route, count := range a {
	// 		fmt.Println(route)
	// 		fmt.Println(count)
	// 	}
	// }

	// return

	fmt.Println(fmt.Sprintf("\n%s v%s - woanware\n", APP_NAME, APP_VERSION))

	parseCommandLine()
	config = LoadConfig()
	configureDatabase()

	// asNames = NewAsNames()
	// err := asNames.Update()
	// if err != nil {
	// 	fmt.Printf("Error downloading AS data: %v\n", err)
	// 	return
	// }

	// for number, a := range asNames.Names {
	// 	fmt.Printf("%v\n", number)
	// 	fmt.Printf("%v\n", a.Country)
	// 	fmt.Printf("%v\n",pgx a.Name)
	// 	fmt.Printf("%v\n--pgx------------------------------\n", a.Description)
	// 	//fmt.Printf("%v\npgx", a)
	// }pgx

	h, err := NewHistory(config.HistoryMonths, config.Processes)
	if err != nil {
		return
	}
	h.Update()
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
