package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	try "github.com/matryer/try"
	util "github.com/woanware/goutil"
)

const RIPE_UPDATES string = "http://data.ris.ripe.net/rrc00/"
const HISTORY_MONTHS int = 2

//
func main() {

	loadHistoricalData()

	return

}

// Downloads and loads/parses the BGP update files
func loadHistoricalData() {

	// Get a constant value for NOW
	timestampNow := time.Now()

	var err error
	var year int
	var month int

	for i := HISTORY_MONTHS - 1; i >= 0; i-- {

		year = int(timestampNow.AddDate(0, -i, 0).Year())
		month = int(timestampNow.AddDate(0, -i, 0).Month())

		err = checkDirectories(year, month)
		if err != nil {
			fmt.Printf("Error validating directory stores: %v", err)
			continue
		}

		downloadUpdateFiles(year, month)
	}

	fmt.Println("START")
	fmt.Println(time.Now())

	mrtParser := new(MrtParser)
	for i := HISTORY_MONTHS - 1; i >= 0; i-- {

		//fmt.Printf("Update file in cache: %s\n", href)
		// 			bgpDump := new(BGPDump)

		// 			_, err := bgpDump.Parse(fmt.Sprintf("cache/%v/%v/%s", timestamp.Year(), int(timestamp.Month()), href))
		// 			if err != nil {

		// 			}

		year = int(timestampNow.AddDate(0, -i, 0).Year())
		month = int(timestampNow.AddDate(0, -i, 0).Month())

		files, err := ioutil.ReadDir(fmt.Sprintf("./cache/%v/%v", year, month))
		if err != nil {
			log.Fatal(err)
		}

		var wg sync.WaitGroup
		semaphore := make(chan struct{}, 4)
		for _, file := range files {
			wg.Add(1)

			go func(year int, month int, filePath string) {
				defer wg.Done()

				semaphore <- struct{}{} // Lock
				defer func() {
					<-semaphore // Unlock
				}()

				data, err := mrtParser.Parse(fmt.Sprintf("./cache/%v/%v/%s", year, month, filePath))
				if err != nil {
					if strings.Contains(err.Error(), "gzip: invalid header") == true {
						err = os.Remove(fmt.Sprintf("./cache/%v/%v/%s", year, month, filePath))
						if err != nil {
							fmt.Println("Error deleting malformed BGP file (%s): %v\n", fmt.Sprintf("./cache/%v/%v/%s", year, month, filePath), err)
						}
					} else {
						fmt.Println("Error parsing BGP file (%s): %v\n", filePath, err)
					}
				} else {
					for k, v := range data {
						fmt.Printf("%v: %v\n", k, v)
					}
				}
			}(year, month, file.Name())
		}
		wg.Wait()
	}

	fmt.Println("FINISH")
	fmt.Println(time.Now())
}

// Ensures the appropriate year/month "temp" and "cache" directories exist
func checkDirectories(year int, month int) error {

	if util.DoesDirExist(fmt.Sprintf("./cache/%v/%v", year, month)) == false {
		err := os.MkdirAll(fmt.Sprintf("./cache/%v/%v", year, month), 0770)
		if err != nil {
			return err
		}
	}

	if util.DoesDirExist(fmt.Sprintf("./temp/%v/%v", year, month)) == false {
		err := os.MkdirAll(fmt.Sprintf("./temp/%v/%v", year, month), 0770)
		if err != nil {
			return err
		}
	}

	return nil
}

// Downloads RIPE page containing BGP update files, using a specific year/month
// index. Parses the page for update files, checks if the file has already been
// downloaded and the file header checked (GZIP)
func downloadUpdateFiles(year int, month int) {

	fmt.Println(fmt.Sprintf("%s%v.%v", RIPE_UPDATES, year, month))
	doc, err := goquery.NewDocument(fmt.Sprintf("%s%v.%v", RIPE_UPDATES, year, month))
	if err != nil {
		fmt.Printf("Error downloading update page (%s): %v\n", fmt.Sprintf("%s%v.%v", RIPE_UPDATES, year, month), err)
		os.Exit(-1)
	}

	doc.Find("a[href]").Each(func(index int, item *goquery.Selection) {

		href, _ := item.Attr("href")

		if strings.HasPrefix(href, "updates.") == true {

			// Check if update file has been cached, if not then download
			if util.DoesFileExist(fmt.Sprintf("./cache/%v/%v/%s", year, month, href)) == false {

				fmt.Printf("Update file not in cache: %s\n", href)
				err = downloadUpdateFile(year, month, href)
				if err != nil {
					fmt.Printf("Error downloading update file: %s\n", href)
				}
			}
		}
	})
}

// Performs the actual BGP update file downloading
func downloadUpdateFile(year int, month int, href string) error {

	err := try.Do(func(attempt int) (bool, error) {
		var err error

		// Download the file to the "cache" directory
		err = DownloadFile(fmt.Sprintf("./temp/%v/%v/%s", year, month, href),
			fmt.Sprintf("%s/%v.%v/%s", RIPE_UPDATES, year, month, href))

		if err != nil {
			fmt.Printf("Error downloading update file (%s): %v\n", href, err)
		} else {
			err = validateGzipFile(fmt.Sprintf("./temp/%v/%v/%s", year, month, href))
			if err == nil {
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

// Validates a file to ensure that the GZIP header can be read
func validateGzipFile(filePath string) error {

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}

	gzipReader, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("couldn't create gzip reader: %v", err)
	}

	gzipReader.Close()
	f.Close()

	return nil
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
