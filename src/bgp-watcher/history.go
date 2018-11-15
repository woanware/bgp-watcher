package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx"
)

// ##### Structs ##############################################################

//
type History struct {
	DataSets  map[string]string
	Months    int
	Processes int
	detector  *Detector
}

// ##### Methods ##############################################################

//
func NewHistory(d *Detector, dataSets map[string]string, months int, processes int) *History {

	return &History{
		detector:  d,
		DataSets:  dataSets,
		Months:    months,
		Processes: processes,
	}
}

//
func (h *History) Existing() bool {

	files, err := ioutil.ReadDir("./cache/")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(len(files))

	if len(files) == 0 {
		return false
	}

	return true
}

// Downloads and loads/parses the BGP update files
func (h *History) Update() {

	// Get a constant value for NOW
	ts := time.Now()

	h.checkDirectories(ts)
	h.checkFiles(ts)
	h.download(ts)
	h.parse(ts)
}

// checkDirectories ensures that the required directory structure has been created
func (h *History) checkDirectories(ts time.Time) {

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
func (h *History) checkFiles(ts time.Time) {

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

func (h *History) download(ts time.Time) {

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
func (h *History) downloadUpdateFiles(name string, url string, year int, month int) {

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
func (h *History) parse(ts time.Time) {

	fmt.Println("START")
	fmt.Println(time.Now())

	var year int
	var month int

	historyStore := &HistoryStore{data: make(map[uint32]map[string]uint64)}
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

				go func(historyStore *HistoryStore, year int, month int, filePath string) {
					defer wg.Done()

					semaphore <- struct{}{} // Lock
					defer func() {
						<-semaphore // Unlock
					}()

					historyStore, err = mrtParser.ParseAndCollect(h.detector, historyStore, fmt.Sprintf("./cache/%s/%v/%v/%s", name, year, month, filePath))
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
				}(historyStore, year, month, file.Name())
			}
			wg.Wait()
		}
	}

	storeUpdates(historyStore)

	fmt.Println("FINISH")
	fmt.Println(time.Now())
}

//
func storeUpdates(historyStore *HistoryStore) {

	// Truncate table
	_, err := pool.Exec("truncate table routes")
	if err != nil {
		fmt.Printf("Error truncating historic data: %v\n", err)
	}

	var rows [][]interface{}

	for peer, a := range historyStore.data {
		for route, count := range a {
			rows = append(rows, []interface{}{peer, route, count})
		}
	}

	_, err = pool.CopyFrom(
		pgx.Identifier{"routes"},
		[]string{"peer_as", "route", "count"},
		pgx.CopyFromRows(rows))

	if err != nil {
		fmt.Printf("Error inserting historic data: %v\n", err)
		return
	}
}
