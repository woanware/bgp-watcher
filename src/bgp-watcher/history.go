package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	util "github.com/woanware/goutil"
	try "gopkg.in/matryer/try.v1"
)

//
type History struct {
	Months    int
	Processes int
}

//
func NewHistory(months int, processes int) (*History, error) {
	return &History{
		Months:    months,
		Processes: processes,
	}, nil
}

// Downloads and loads/parses the BGP update files
func (h *History) Update() {

	// Get a constant value for NOW
	ts := time.Now()

	h.download(ts)
	h.parse(ts)
}

func (h *History) download(ts time.Time) {

	var err error
	var year int
	var month int

	for i := h.Months - 1; i >= 0; i-- {

		year = int(ts.AddDate(0, -i, 0).Year())
		month = int(ts.AddDate(0, -i, 0).Month())

		err = checkDirectories(year, month)
		if err != nil {
			fmt.Printf("Error validating directory stores: %v", err)
			continue
		}

		h.downloadUpdateFiles(year, month)
	}
}

// Downloads RIPE page containing BGP update files, using a specific year/month
// index. Parses the page for update files, checks if the file has already been
// downloaded and the file header checked (GZIP)
func (h *History) downloadUpdateFiles(year int, month int) {

	// Download the update page that is specific for the year/month
	fmt.Println(fmt.Sprintf("Downloading update page: %s%v.%02d", RIPE_UPDATES, year, month))
	doc, err := goquery.NewDocument(fmt.Sprintf("%s%v.%02d", RIPE_UPDATES, year, month))
	if err != nil {
		fmt.Printf("Error downloading update page (%s): %v\n", fmt.Sprintf("%s%v.%v", RIPE_UPDATES, year, month), err)
		os.Exit(-1)
	}

	files := make([]string, 0)

	// Parse the HTML and extract all "a" elements
	doc.Find("a[href]").Each(func(index int, item *goquery.Selection) {

		href, _ := item.Attr("href")

		// Ensure that the extracted link has "updates." in the file name
		if strings.HasPrefix(href, "updates.") == true {

			// Check if update file has been cached, if not then download
			if util.DoesFileExist(fmt.Sprintf("./cache/%v/%v/%s", year, month, href)) == false {

				files = append(files, href)

			}
		}
	})

	// Now perform the actual downloading concurrentl
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

			fmt.Printf("Uncached update file: %s\n", fileName)
			err = downloadUpdateFile(year, month, fileName)
			if err != nil {
				fmt.Printf("Error downloading update file (%s): %v\n", fileName, err)
			} else {

			}

		}(year, month, file)
	}
	wg.Wait()
}

// Performs the actual BGP update file downloading
func downloadUpdateFile(year int, month int, href string) error {

	err := try.Do(func(attempt int) (bool, error) {
		var err error

		// Download the file to the "temp" directory
		err = DownloadFile(fmt.Sprintf("./temp/%v/%v/%s", year, month, href),
			fmt.Sprintf("%s/%v.%02d/%s", RIPE_UPDATES, year, month, href))

		if err != nil {
			fmt.Printf("Error downloading update file (%s): %v\n", href, err)
		} else {
			// Make sure the file header reads OK (gzip)
			err = validateGzipFile(fmt.Sprintf("./temp/%v/%v/%s", year, month, href))
			if err == nil {
				// Move the file to the "cache" directory
				err = os.Rename(fmt.Sprintf("./temp/%v/%v/%s", year, month, href), fmt.Sprintf("./cache/%v/%v/%s", year, month, href))
				if err != nil {
					fmt.Printf("Error moving temp update file to cache (%s): %v\n", href, err)
				}
				return false, nil
			}
		}

		return attempt < 3, err // try 3 times
	})

	return err
}

//
func (h *History) parse(ts time.Time) {

	fmt.Println("START")
	fmt.Println(time.Now())

	var year int
	var month int
	asns := make(map[uint32]map[string]uint64)

	mrtParser := new(MrtParser)
	for i := h.Months - 1; i >= 0; i-- {

		year = int(ts.AddDate(0, -i, 0).Year())
		month = int(ts.AddDate(0, -i, 0).Month())

		files, err := ioutil.ReadDir(fmt.Sprintf("./cache/%v/%v", year, month))
		if err != nil {
			log.Fatal(err)
		}

		var wg sync.WaitGroup
		semaphore := make(chan struct{}, h.Processes)
		for _, file := range files {
			wg.Add(1)

			go func(asns map[uint32]map[string]uint64, year int, month int, filePath string) {
				defer wg.Done()

				semaphore <- struct{}{} // Lock
				defer func() {
					<-semaphore // Unlock
				}()

				_, err := mrtParser.Parse(asns, fmt.Sprintf("./cache/%v/%v/%s", year, month, filePath))
				if err != nil {
					if strings.Contains(err.Error(), "gzip: invalid header") == true {
						err = os.Remove(fmt.Sprintf("./cache/%v/%v/%s", year, month, filePath))
						if err != nil {
							fmt.Println("Error deleting malformed BGP file (%s): %v\n", fmt.Sprintf("./cache/%v/%v/%s", year, month, filePath), err)
						}
					} else {
						fmt.Println("Error parsing BGP file (%s): %v\n", filePath, err)
					}
				}
			}(asns, year, month, file.Name())
		}
		wg.Wait()
	}

	for k, v := range asns {
		fmt.Printf("%v: %v\n", k, v)
	}

	fmt.Println("FINISH")
	fmt.Println(time.Now())
}
