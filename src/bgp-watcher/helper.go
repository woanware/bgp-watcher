package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	color "github.com/labstack/gommon/color"
	"github.com/matryer/try"
	util "github.com/woanware/goutil"
)

// Ensures the appropriate year/month "temp" and "cache" directories exist
func checkDirectory(name string, year int, month int) error {

	if util.DoesDirExist(fmt.Sprintf("./cache/%s/%v/%v", name, year, month)) == false {
		err := os.MkdirAll(fmt.Sprintf("./cache/%s/%v/%v", name, year, month), 0770)
		if err != nil {
			return err
		}
	}

	if util.DoesDirExist(fmt.Sprintf("./temp/%s/%v/%v", name, year, month)) == false {
		err := os.MkdirAll(fmt.Sprintf("./temp/%s/%v/%v", name, year, month), 0770)
		if err != nil {
			return err
		}
	}

	return nil
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

//
func getUpdateFiles(name string, url string, year int, month int) ([]string, error) {

	files := make([]string, 0)

	// Download the update page that is specific for the year/month
	fmt.Println(fmt.Sprintf("Downloading update page: %s%v.%02d", url, year, month))
	doc, err := goquery.NewDocument(fmt.Sprintf("%s%v.%02d", url, year, month))
	if err != nil {
		return files, fmt.Errorf("Error downloading update page (%s): %v\n", fmt.Sprintf("%s%v.%v", url, year, month), err)
	}

	// Parse the HTML and extract all "a" elements
	doc.Find("a[href]").Each(func(index int, item *goquery.Selection) {

		href, _ := item.Attr("href")

		// Ensure that the extracted link has "updates." in the file name
		if strings.HasPrefix(href, "updates.") == false {
			return
		}

		// Check if update file has been cached, if not then download
		if util.DoesFileExist(fmt.Sprintf("./cache/%s/%v/%v/%s", name, year, month, href)) == true {
			return
		}

		files = append(files, href)
	})

	return files, err
}

// Performs the actual BGP update file downloading
func downloadUpdateFile(name string, year int, month int, href string) error {

	err := try.Do(func(attempt int) (bool, error) {
		var err error

		// Download the file to the "temp" directory
		err = DownloadFile(fmt.Sprintf("./temp/%s/%v/%v/%s", name, year, month, href),
			fmt.Sprintf("%s/%v.%02d/%s", RIPE_UPDATES, year, month, href))

		if err == nil {
			// Make sure the file header reads OK (gzip)
			err = validateGzipFile(fmt.Sprintf("./temp/%s/%v/%v/%s", name, year, month, href))
			if err == nil {
				// Move the file to the "cache" directory
				err = os.Rename(fmt.Sprintf("./temp/%s/%v/%v/%s", name, year, month, href), fmt.Sprintf("./cache/%s/%v/%v/%s", name, year, month, href))
				if err != nil {
					fmt.Printf("Error moving temp update file to cache (%s): %v\n", href, err)
				}
				return false, nil
			} else {
				err = os.Remove(fmt.Sprintf("./temp/%s/%v/%v/%s", name, year, month, href))
				if err != nil {
					fmt.Printf("Error deleting corrupt update file (%s/%v/%v/%s): %v\n", name, year, month, href, err)
				}
			}
		}

		return attempt < 3, err // try 3 times
	})

	return err
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

// Converts an UInt16 to a string
func ConvertUInt32ToString(data uint32) string {

	return strconv.FormatInt(int64(data), 10)
}

// Converts a string to an uint8
func ConvertStringToUint8(data string) (uint8, error) {

	ret, err := strconv.ParseInt(data, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint8(ret), nil
}

// Converts a string to an uint32
func ConvertStringToUint32(data string) (uint32, error) {

	ret, err := strconv.ParseInt(data, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(ret), nil
}

//
func outputAlert(ap AlertPriority, alert string) {

	switch ap {
	case PriorityHigh:
		color.Println(color.Red(alert))
	case PriorityMedium:
		color.Println(color.Yellow(alert))
	case PriorityLow:
		color.Println(color.Green(alert))
	}
}
