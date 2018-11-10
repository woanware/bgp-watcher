package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"

	util "github.com/woanware/goutil"
)

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
