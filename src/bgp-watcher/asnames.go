package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	util "github.com/woanware/goutil"
)

// ##### Structs ##############################################################

// AsName encapsulates a single AS record
type AsName struct {
	Name        string
	Description string
	Country     string
}

// AsNames holds the various AsName structs extracted from the Cidr Report page
type AsNames struct {
	names map[uint32]*AsName
}

// ##### Constants ############################################################

const CIDR_DATA_URL string = "http://www.cidr-report.org/as2.0/autnums.html"

// ##### Methods ##############################################################

// NewAsNames returns a new AsNames struct
func NewAsNames() *AsNames {

	asNames := new(AsNames)
	asNames.names = make(map[uint32]*AsName)
	return asNames
}

// Country returns the country code associated with an AS
func (a *AsNames) Country(as uint32) string {

	if _, ok := a.names[as]; ok {
		return a.names[as].Country
	}

	return ""
}

// Update retrieves the AS data from www.cidr-report.org/as2.0/autnums.html
func (a *AsNames) Update() error {

	fmt.Println("Updating CIDR data")

	resp, err := http.Get(CIDR_DATA_URL)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var index int
	var as uint32
	var asName *AsName
	var temp string

	var re = regexp.MustCompile(`(?m)<a\shref=.*">AS(\d*)\s*<\/a>\s(.*)`)
	for _, match := range re.FindAllStringSubmatch(string(html), -1) {

		asName = new(AsName)

		index = strings.LastIndex(match[2], ",")
		asName.Country = strings.TrimSpace(match[2][index+1:])
		temp = strings.TrimSpace(match[2][:index])

		// The character group of " - " appears to split the AS
		// name and description in the majority of cases
		index = strings.Index(temp, " - ")
		if index > -1 {
			asName.Name = temp[0:index]
			asName.Description = temp[index+3:]
		} else {
			asName.Name = temp

			// Attempt to match "UBSGROUPAG-AS-AP UBS Group AG"
			// e.g. has "-", so split off the word after the first space
			if strings.Contains(temp, "-") == true {
				index = strings.Index(temp, " ")
				if index > -1 {
					asName.Name = temp[0:index]
					asName.Description = temp[index+1:]

				}
			}
		}

		as, err = util.ConvertStringToUint32(match[1])
		if err != nil {
			fmt.Printf("Error converting AS data AS number (%v): %v\n", match[1], err)
			continue
		}
		a.names[as] = asName
	}

	return nil
}
