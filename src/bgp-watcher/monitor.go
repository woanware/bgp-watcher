package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"sync"
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

	// c := cron.New()
	// c.AddFunc("@every 1m", m.check)
	// c.Start()

	m.check()
}

//
func (m *Monitor) check() {

	if m.updating == true {
		return
	}

	m.updating = true
	defer func() { m.updating = false }()

	files, err := ioutil.ReadDir("./cache/2018/11/")
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

			// if strings.Contains(file, "201811") == false {
			// 	continue
			// }

			//fmt.Println("Parsing file: " + file.Name())
			err = mrtParser.Parse(m.detector, "./cache/2018/11/"+file)
			if err != nil {
				fmt.Printf("Error parsing update file (%s): %v\n", file, err)
			}

		}(file.Name())
	}
	wg.Wait()

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
	// 			err = mrtParser.Parse(m.detector, fmt.Sprintf("./cache/%v/%v/%s", year, month, fileName))
	// 			if err != nil {
	// 				fmt.Printf("Error parsing update file (%s): %v\n", fileName, err)
	// 			}
	// 		}
	// 	}(year, month, file)
	// }
	// wg.Wait()
}
