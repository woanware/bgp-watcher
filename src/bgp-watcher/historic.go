package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// ##### Structs ##############################################################

//
type Historic struct {
	DataSets  map[string]string
	Months    int
	Processes int
	detector  *Detector
}

// ##### Methods ##############################################################

//
func NewHistoric(d *Detector, config *Config) *Historic {

	return &Historic{
		detector:  d,
		DataSets:  config.DataSets,
		Months:    config.HistoryMonths,
		Processes: config.Processes,
	}
}

//
func (h *Historic) Existing() bool {

	files, err := ioutil.ReadDir("./cache/")
	if err != nil {
		log.Fatal(err)
	}

	if len(files) == 0 {
		return false
	}

	return true
}

// Downloads and loads/parses the BGP update files
func (h *Historic) Update() {

	fmt.Println("Performing historic data refresh")

	// Get a constant value for NOW
	ts := time.Now()

	h.checkDirectories(ts)
	h.checkFiles(ts)
	h.download(ts)
	h.parse(ts)
}

// checkDirectories ensures that the required directory structure has been created
func (h *Historic) checkDirectories(ts time.Time) {

	var year int
	var month int
	var err error

	for name := range config.DataSets {
		for i := h.Months - 1; i >= 0; i-- {

			year = int(ts.AddDate(0, -i, 0).Year())
			month = int(ts.AddDate(0, -i, 0).Month())

			err = checkDirectory(name, year, month)
			if err != nil {
				fmt.Printf("Error validating directory stores (%s/%v/%v): %v", name, year, month, err)
				continue
			}
		}
	}
}

// checkDirectories ensures that the required directory structure has been created
func (h *Historic) checkFiles(ts time.Time) {

	var year int
	var month int

	for name := range config.DataSets {
		for i := h.Months - 1; i >= 0; i-- {

			year = int(ts.AddDate(0, -i, 0).Year())
			month = int(ts.AddDate(0, -i, 0).Month())

			files, err := ioutil.ReadDir(fmt.Sprintf("./cache/%s/%d/%d", name, year, month))
			if err != nil {
				log.Fatal(err)
			}

			for _, file := range files {
				err = validateGzipFile(fmt.Sprintf("./cache/%s/%d/%d/%s", name, year, month, file.Name()))
				if err != nil {
					fmt.Printf("Corrupt update file ./cache/%s/%d/%d/%s", name, year, month, file.Name())
					err = os.Remove(fmt.Sprintf("./cache/%s/%d/%d/%s", name, year, month, file.Name()))
					if err != nil {
						fmt.Printf("Error deleting corrupt update file (%s/%d/%d/%s): %v\n", name, year, month, file.Name(), err)
					}
				}
			}
		}
	}
}

func (h *Historic) download(ts time.Time) {

	var year int
	var month int

	for name, url := range config.DataSets {
		for i := h.Months - 1; i >= 0; i-- {

			year = int(ts.AddDate(0, -i, 0).Year())
			month = int(ts.AddDate(0, -i, 0).Month())

			h.downloadUpdateFiles(name, url, year, month)
		}
	}
}

// Downloads RIPE page containing BGP update files, using a specific year/month
// index. Parses the page for update files, checks if the file has already been
// downloaded and the file header checked (GZIP)
func (h *Historic) downloadUpdateFiles(name string, url string, year int, month int) {

	files, err := getUpdateFiles(name, url, year, month)
	if err != nil {
		fmt.Printf("Error retrieving update file list (%s): %v\n", fmt.Sprintf("%s%v.%v", url, year, month), err)
		return
	}

	// Now perform the actual downloading concurrently
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, h.Processes)
	for _, file := range files {
		wg.Add(1)

		go func(year int, month int, fileName string) {
			defer wg.Done()

			semaphore <- struct{}{} // Lock
			defer func() {
				<-semaphore // Unlock
			}()

			fmt.Printf("Uncached update file (%s): %s\n", name, fileName)
			err = downloadUpdateFile(name, url, year, month, fileName)
			if err != nil {
				fmt.Printf("Error downloading update file (%s): %v\n", fileName, err)
			} else {

			}

		}(year, month, file)
	}
	wg.Wait()
}

//
func (h *Historic) parse(ts time.Time) {

	fmt.Println("START")
	fmt.Println(time.Now())

	var year int
	var month int

	//historyStore := &HistoryStore{data: make(map[uint32]map[string]uint64)}
	//asns := make(map[uint32]map[string]uint64)

	mrtParser := new(MrtParser)
	for name := range config.DataSets {
		for i := h.Months - 1; i >= 0; i-- {

			year = int(ts.AddDate(0, -i, 0).Year())
			month = int(ts.AddDate(0, -i, 0).Month())

			files, err := ioutil.ReadDir(fmt.Sprintf("./cache/%s/%v/%v", name, year, month))
			if err != nil {
				log.Fatal(err)
			}

			var wg sync.WaitGroup
			semaphore := make(chan struct{}, h.Processes)
			for _, file := range files {
				wg.Add(1)

				go func(year int, month int, filePath string) {
					defer wg.Done()

					semaphore <- struct{}{} // Lock
					defer func() {
						<-semaphore // Unlock
					}()

					err = mrtParser.ParseAndCollect(h.detector, fmt.Sprintf("./cache/%s/%v/%v/%s", name, year, month, filePath))
					if err != nil {
						if strings.Contains(err.Error(), "gzip: invalid header") == true {
							err = os.Remove(fmt.Sprintf("./cache/%s/%v/%v/%s", name, year, month, filePath))
							if err != nil {
								fmt.Println("Error deleting malformed BGP file (%s): %v\n", fmt.Sprintf("./cache/%s/%v/%v/%s", name, year, month, filePath), err)
							}
						} else {
							fmt.Println("Error parsing BGP file (%s): %v\n", filePath, err)
						}
					}
				}(year, month, file.Name())
			}
			wg.Wait()
		}
	}

	history.Persist()

	fmt.Println("FINISH")
	fmt.Println(time.Now())
}
