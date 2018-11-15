package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"sync"

	cron "github.com/robfig/cron"
)

// ##### Structs ##############################################################

//
type Monitor struct {
	Processes    int
	updating     bool
	historyStore *HistoryStore
	detector     *Detector
}

// ##### Methods ##############################################################

//
func NewMonitor(d *Detector, processes int) *Monitor {
	return &Monitor{
		Processes: processes,
		detector:  d,
	}
}

//
func (m *Monitor) Start() {

	m.detector.Start()
	m.load()
	c := cron.New()
	c.AddFunc("@every 1m", m.check)
	// c.Start()

	m.check()
}

//
func (m *Monitor) load() {

	m.historyStore = &HistoryStore{data: make(map[uint32]map[string]uint64)}

	rows, err := pool.Query("select peer_as, route, count from routes")
	if err != nil {
		fmt.Printf("Error loading historical data: %v\n", err)
		return
	}
	defer rows.Close()

	var peerAs uint32
	var route string
	var count uint64

	for rows.Next() {
		err = rows.Scan(&peerAs, &route, &count)
		if err != nil {
			fmt.Printf("Error loading historical data row: %v\n", err)
			continue
		}

		m.historyStore.SetCount(peerAs, route, count)
	}
}

//
func (m *Monitor) check() {

	if m.updating == true {
		return
	}

	m.updating = true
	defer func() { m.updating = false }()

	historyStore := new(HistoryStore)

	for name := range config.DataSets {

		files, err := ioutil.ReadDir("./cache/" + name + "/2018/11/")
		if err != nil {
			log.Fatal(err)
		}

		// Now perform the actual downloading concurrently
		var wg sync.WaitGroup
		semaphore := make(chan struct{}, m.Processes)
		fmt.Println(len(files))
		mrtParser := new(MrtParser)
		for _, file := range files {
			wg.Add(1)

			go func(file string) {
				defer wg.Done()

				semaphore <- struct{}{} // Lock
				defer func() {
					<-semaphore // Unlock
				}()

				if strings.Contains(file, "20181109") == false && strings.Contains(file, "20181110") == false &&
					strings.Contains(file, "20181111") == false && strings.Contains(file, "20181112") == false &&
					strings.Contains(file, "20181113") == false && strings.Contains(file, "20181114") == false {
					return
				}

				//fmt.Println("Parsing file: " + file.Name())
				historyStore, err = mrtParser.ParseAndDetect(m.detector, name, "./cache/"+name+"/2018/11/"+file)
				if err != nil {
					fmt.Printf("Error parsing update file (%s): %v\n", file, err)
				}

			}(file.Name())
		}
		wg.Wait()
	}

	// Now update our historical data
	for peer, a := range historyStore.data {
		for route, count := range a {
			m.historyStore.SetAdd(peer, route, count)
		}
	}

	fmt.Println("FINISHED")

	// // Get a constant value for NOW
	// ts := time.Now()

	// year := ts.Year()
	// month := int(ts.Month())

	// files, err := getUpdateFiles(year, month)
	// if err != nil {
	// 	return
	// }

	// mrtParser := new(MrtParser)

	// // Now perform the actual downloading concurrently
	// var wg sync.WaitGroup
	// semaphore := make(chan struct{}, m.Processes)
	// for _, file := range files {
	// 	wg.Add(1)

	// 	go func(year int, month int, fileName string) {
	// 		defer wg.Done()

	// 		semaphore <- struct{}{} // Lock
	// 		defer func() {
	// 			<-semaphore // Unlock
	// 		}()

	// 		fmt.Printf("Uncached update file: %s\n", fileName)
	// 		err = downloadUpdateFile(year, month, fileName)
	// 		if err != nil {
	// 			fmt.Printf("Error downloading update file (%s): %v\n", fileName, err)
	// 		} else {
	// 			historyStore, err = mrtParser.ParseAndDetect(m.detector, name, "./cache/"+name+"/2018/11/"+file)
	// if err != nil {
	// 	fmt.Printf("Error parsing update file (%s): %v\n", file, err)
	// }
	// 		}
	// 	}(year, month, file)
	// }
	// wg.Wait()
}
