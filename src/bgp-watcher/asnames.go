package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	util "github.com/woanware/goutil"
)

//
type AsName struct {
	Name        string
	Description string
	Country     string
}

//
type AsNames struct {
	Names map[int]*AsName
}

//
func NewAsNames() *AsNames {

	asNames := new(AsNames)
	asNames.Names = make(map[int]*AsName)
	return asNames
}

const CIDR_DATA_URL string = "http://www.cidr-report.org/as2.0/autnums.html"

// Retrieve the AS data from www.cidr-report.org/as2.0/autnums.html
func (a *AsNames) Update() error {

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
	var asName *AsName

	var re = regexp.MustCompile(`(?m)<a\shref=".*?">AS(\d*)\s*<\/a>\s(.*)\s-\s(.*)`)
	for _, match := range re.FindAllStringSubmatch(string(html), -1) {

		asName = new(AsName)

		asName.Name = strings.TrimSpace(match[2])
		index = strings.LastIndex(match[3], ",")
		asName.Description = strings.TrimSpace(match[3][:index])
		asName.Country = strings.TrimSpace(match[3][index+1:])

		// index = strings.Index(match[2], " - ")
		// asName.Name = strings.TrimSpace(match[2][:index])
		// asName.Description = strings.TrimSpace(match[2][:]index)

		a.Names[util.ConvertStringToInt(match[1])] = asName

		fmt.Println(asName.Name)
		fmt.Println(asName.Description)
		fmt.Println(asName.Country)

		//fmt.Println(country)

		//asnNames[]
	}

	return nil
}
