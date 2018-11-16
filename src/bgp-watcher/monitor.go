package main

import (
	"fmt"
	"sync"
	"time"

	cron "github.com/robfig/cron"
)

// ##### Structs ##############################################################

//
type Monitor struct {
	Processes int
	updating  bool
	detector  *Detector
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
	c := cron.New()
	c.AddFunc("@every 1m", m.check)
	c.Start()

	// DEBUG
	//m.check()
}

//
func (m *Monitor) check() {

	if m.updating == true {
		return
	}

	m.updating = true
	defer func() { m.updating = false }()

	// We use a temp history store so that detection is based
	// on historic data, not data that is incoming. Hopefully our
	// detection processing is fast enough to out run the adding of
	// the temp data into the primary history store
	tempHistory := new(History)
	mrtParser := new(MrtParser)

	// Get a constant value for NOW
	ts := time.Now()

	year := ts.Year()
	month := int(ts.Month())

	//DEBUG
	// for name := range config.DataSets {

	// 	files, err := ioutil.ReadDir("./cache/" + name + "/2018/11/")
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	// Now perform the actual downloading concurrently
	// 	var wg sync.WaitGroup
	// 	semaphore := make(chan struct{}, m.Processes)
	// 	fmt.Println(len(files))

	// 	for _, file := range files {
	// 		wg.Add(1)

	// 		go func(file string) {
	// 			defer wg.Done()

	// 			semaphore <- struct{}{} // Lock
	// 			defer func() {
	// 				<-semaphore // Unlock
	// 			}()

	// 			if strings.Contains(file, "20181109") == false && strings.Contains(file, "20181110") == false &&
	// 				strings.Contains(file, "20181111") == false && strings.Contains(file, "20181112") == false &&
	// 				strings.Contains(file, "20181113") == false && strings.Contains(file, "20181114") == false {
	// 				return
	// 			}

	// 			//fmt.Println("Parsing file: " + file.Name())
	// 			tempHistory, err = mrtParser.ParseAndDetect(m.detector, name, "./cache/"+name+"/2018/11/"+file)
	// 			if err != nil {
	// 				fmt.Printf("Error parsing update file (%s): %v\n", file, err)
	// 			}

	// 		}(file.Name())
	// 	}
	// 	wg.Wait()
	// }

	fmt.Printf("Processing updates starting: %v\n", time.Now().Format("2006-01-02T15:04:05"))

	var wg sync.WaitGroup
	var files []string
	var file string
	var err error

	semaphore := make(chan struct{}, m.Processes)
	for name, url := range config.DataSets {

		files, err = getUpdateFiles(name, url, year, month)
		if err != nil {
			return
		}

		for _, file = range files {
			wg.Add(1)

			go func(year int, month int, fileName string) {
				defer wg.Done()

				semaphore <- struct{}{} // Lock
				defer func() {
					<-semaphore // Unlock
				}()

				fmt.Printf("Uncached update file: %s\n", fileName)
				err = downloadUpdateFile(name, url, year, month, fileName)
				if err != nil {
					fmt.Printf("Error downloading update file (%s): %v\n", fileName, err)
				} else {
					tempHistory, err = mrtParser.ParseAndDetect(m.detector, name, fmt.Sprintf("./cache/%s/%d/%d/%s", name, year, month, fileName))
					if err != nil {
						fmt.Printf("Error parsing update file (%s): %v\n", file, err)
					}
				}
			}(year, month, file)
		}
	}
	wg.Wait()

	// Now update our data to the primary history object
	for peer, a := range tempHistory.data {
		for route, count := range a {
			history.SetAdd(peer, route, count)
		}
	}

	fmt.Printf("Processing updates finished: %v\n", time.Now().Format("2006-01-02T15:04:05"))
}
