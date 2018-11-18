package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	pgx "github.com/jackc/pgx"
	flags "github.com/jessevdk/go-flags"
	viper "github.com/spf13/viper"
)

// ##### Constants #####################################################################################################

const APP_NAME string = "bgp-monitor (bgpm)"
const APP_VERSION string = "0.0.2"

// ##### Variables #####################################################################################################

var (
	configReader *viper.Viper
	config       *Config
	pool         *pgx.ConnPool
	options      Options
	asNames      *AsNames
	history      *History
)

// ##### Methods ##############################################################

//
func main() {

	fmt.Println(fmt.Sprintf("\n%s v%s - woanware\n", APP_NAME, APP_VERSION))

	parseCommandLine()
	initialiseConfiguration()
	config = parseConfiguration()
	configureDatabase()

	asNames = NewAsNames()
	err := asNames.Update()
	if err != nil {
		fmt.Printf("Error downloading AS data: %v\n", err)
		return
	}

	history = NewHistory()
	detector := NewDetector(config)
	historic := NewHistoric(detector, config)

	if options.Reparse == true {
		historic.Update()
	} else {
		if historic.Existing() == false {
			historic.Update()
		} else {
			history.Load()
		}
	}

	history.Summary()

	monitor := NewMonitor(detector, config.Processes)
	monitor.Start()

	// Ensure the application does not exit and we capture CTRL-C
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		_ = <-sigs
		done <- true
	}()
	<-done

	fmt.Printf("\nPersisting historic data\n")
	history.Persist()
	fmt.Println("Persistance complete")
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

	connPoolConfig := pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     config.DatabaseServer,
			User:     config.DatabaseUsername,
			Password: config.DatabasePassword,
			Database: config.Database,
		},
		MaxConnections: 5,
		//AfterConnect:   configurePreparedStatements,
	}

	var err error
	pool, err = pgx.NewConnPool(connPoolConfig)
	if err != nil {
		fmt.Printf("Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
}

// func configurePreparedStatements(conn *pgx.Conn) (err error) {

// 	_, err = conn.Prepare("get_route_count", `select count from routes where peer_as=$1 and route=$2`)
// 	if err != nil {
// 		fmt.Printf("Error preparing 'get_route_count' statement: %v\n", err)
// 	}

// 	return
// }

//
func reloadConfig() {

	config = parseConfiguration()
}
