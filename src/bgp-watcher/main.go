package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx"
	flags "github.com/jessevdk/go-flags"
	viper "github.com/spf13/viper"
)

// ##### Constants #####################################################################################################

const APP_NAME string = "bgp-monitor (bgpm)"
const APP_VERSION string = "0.0.1"

// ##### Variables #####################################################################################################

var (
	configReader *viper.Viper
	config       *Config
	pool         *pgx.ConnPool
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

	asNames = NewAsNames()
	err := asNames.Update()
	if err != nil {
		fmt.Printf("Error downloading AS data: %v\n", err)
		return
	}

	detector := NewDetector(asNames)
	for cc := range config.MonitorCountryCodes {
		detector.AddMonitorCountryCode(cc)
	}
	for as := range config.TargetAs {
		detector.AddTargetAs(as)
	}
	for _, prefix := range config.Prefixes {
		detector.AddPrefix(prefix)
	}

	h := NewHistory(detector, config.DataSets, config.HistoryMonths, config.Processes)
	if h.Existing() == false {
		h.Update()
	}

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

	fmt.Println("Persisting historic data")
	monitor.historyStore.Persist()
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
			//Logger:   logger,
		},
		MaxConnections: 5,
		AfterConnect:   configurePreparedStatements,
	}

	var err error
	pool, err = pgx.NewConnPool(connPoolConfig)
	if err != nil {
		fmt.Printf("Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}

	// c, err := pgx.ParseDSN(fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
	// 	config.DatabaseServer, config.DatabasePort, config.Database, config.DatabaseUsername, config.DatabasePassword))

	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Error parsing database config: %v\n", err)
	// 	os.Exit(1)
	// }

	// db, err = pgx.Connect(c)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
	// 	os.Exit(1)
	// }
}

func configurePreparedStatements(conn *pgx.Conn) (err error) {

	_, err = conn.Prepare("get_route_count", `select count from routes where peer_as=$1 and route=$2`)
	if err != nil {
		fmt.Printf("Error preparing 'get_route_count' statement: %v\n", err)
	}

	return
}

//
func reloadConfig() {

	config = parseConfiguration()
}
