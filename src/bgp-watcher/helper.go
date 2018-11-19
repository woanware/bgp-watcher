package main

import (
	"compress/gzip"
	"fmt"
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
func downloadUpdateFile(name string, url string, year int, month int, href string) error {

	err := try.Do(func(attempt int) (bool, error) {
		var err error

		// Download the file to the "temp" directory
		err = util.DownloadToFile(fmt.Sprintf("%s/%d.%02d/%s", url, year, month, href), fmt.Sprintf("./temp/%s/%d/%d/%s", name, year, month, href))

		if err == nil {
			// Make sure the file header reads OK (gzip)
			err = validateGzipFile(fmt.Sprintf("./temp/%s/%d/%d/%s", name, year, month, href))
			if err == nil {
				// Move the file to the "cache" directory
				err = os.Rename(fmt.Sprintf("./temp/%s/%d/%d/%s", name, year, month, href), fmt.Sprintf("./cache/%s/%d/%d/%s", name, year, month, href))
				if err != nil {
					fmt.Printf("Error moving temp update file to cache (%s): %v\n", href, err)
				}
				return false, nil
			} else {
				err = os.Remove(fmt.Sprintf("./temp/%s/%d/%d/%s", name, year, month, href))
				if err != nil {
					fmt.Printf("Error deleting corrupt update file (%s/%d/%d/%s): %v\n", name, year, month, href, err)
				}
			}
		}

		return attempt < 3, err // try 3 times
	})

	return err
}

// convertAsPath returns the integer value path route as a string
func convertAsPath(path []uint32) string {

	temp := []byte{}
	for _, n := range path {
		temp = strconv.AppendInt(temp, int64(n), 10)
		temp = append(temp, ' ')
	}
	temp = temp[:len(temp)-1]

	return string(temp)
}

// printAlert prints a formatted, coloured message to StdOut
func printAlert(ap AlertPriority, timestamp string, peerAs uint32, path string, reason string, data string) {

	switch ap {
	case PriorityHigh:
		color.Println(color.Red(fmt.Sprintf("Timestamp: %s\nReason: %s\nPeer AS: %d\nPath: %s\nData: %s\n", timestamp, reason, peerAs, path, data)))

	case PriorityMedium:
		color.Println(color.Yellow(fmt.Sprintf("Timestamp: %s\nReason: %s\nPeer AS: %d\nPath: %s\nData: %s\n", timestamp, reason, peerAs, path, data)))

	case PriorityLow:
		color.Println(color.Green(fmt.Sprintf("Timestamp: %s\nReason: %s\nPeer AS: %d\nPath: %s\nData: %s\n", timestamp, reason, peerAs, path, data)))
	}
}
