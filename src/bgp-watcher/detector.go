package main

import (
	"fmt"
	"net"
	"time"

	bgp "github.com/osrg/gobgp/packet"
)

// ##### Structs ##############################################################

type DetectData struct {
	Name        string
	Timestamp   time.Time
	PeerIP      net.IP
	PeerAs      uint32
	Paths       []uint32
	PathsString string
	NLRI        []*bgp.IPAddrPrefix
}

type Detector struct {
	queue               chan *DetectData
	targetAs            map[uint32]struct{}
	monitorCountryCodes map[string]struct{}
	prefixes            map[*bgp.IPAddrPrefix]struct{}
}

// ##### Methods ##############################################################

func NewDetector(config *Config) *Detector {

	d := new(Detector)
	d.initialise()

	for cc := range config.MonitorCountryCodes {
		d.AddMonitorCountryCode(cc)
	}
	for as := range config.TargetAs {
		d.AddTargetAs(as)
	}
	for _, prefix := range config.Prefixes {
		d.AddPrefix(prefix)
	}

	return d
}

//
func (d *Detector) initialise() {

	d.queue = make(chan *DetectData)
	d.monitorCountryCodes = make(map[string]struct{})
	d.targetAs = make(map[uint32]struct{})
	d.prefixes = make(map[*bgp.IPAddrPrefix]struct{})
}

//
func (d *Detector) AddTargetAs(as uint32) {

	d.targetAs[as] = struct{}{}
}

//
func (d *Detector) AddPrefix(prefix *bgp.IPAddrPrefix) {

	d.prefixes[prefix] = struct{}{}
}

//
func (d *Detector) AddMonitorCountryCode(cc string) {

	d.monitorCountryCodes[cc] = struct{}{}
}

//
func (d *Detector) CheckTargetAs(as uint32) bool {

	if _, ok := d.targetAs[as]; ok {
		return true
	}

	return false
}

//
func (d *Detector) CheckPrefix(prefix *bgp.IPAddrPrefix) bool {

	if _, ok := d.prefixes[prefix]; ok {
		return true
	}

	return false
}

//
func (d *Detector) CheckMonitorCountryCode(cc string) bool {

	if _, ok := d.monitorCountryCodes[cc]; ok {
		return true
	}

	return false
}

//
func (d *Detector) Start() {

	go func() {
		var dd *DetectData

		for {
			dd = <-d.queue
			go d.detect(dd)
			dd = nil
		}
	}()
}

//
func (d *Detector) Add(name string, timestamp time.Time, peerAs uint32,
	peerIP net.IP, pathsString string, paths []uint32, nlri []*bgp.IPAddrPrefix) {

	d.queue <- &DetectData{Name: name, Timestamp: timestamp, PeerAs: peerAs, PeerIP: peerIP, PathsString: pathsString, Paths: paths, NLRI: nlri}
}

//
func (d *Detector) detect(dd *DetectData) {

	ret := d.isAnomlousCountry(dd)
	if ret == true {
		// We raised an alert so don't process further
		return
	}

	ret = d.isAnomlousPeer(dd)
	if ret == true {
		// We raised an alert so don't process further
		return
	}

	ret = d.isLowFrequency(dd)
	if ret == true {
		// We raised an alert so don't process further
		return
	}
}

// detectAnomlousCountry performs analysis on the countries
// the path goes through. Returns True if nothing suspicious
// identified
func (d *Detector) isAnomlousCountry(dd *DetectData) bool {

	// If the path length equals two then no middle AS
	if len(dd.Paths) == 2 {
		return false
	}

	firstAs := dd.Paths[0]
	firstCountry := asNames.Country(uint32(firstAs))

	// Get last AS Country
	lastAs := dd.Paths[len(dd.Paths)-1]
	lastCountry := asNames.Country(lastAs)

	// If the AS countries are the same then we cannot really check the middle routes
	if firstCountry != lastCountry {
		return false
	}

	var country string
	var count uint64
	ret := false

	// Check the country of the intermediary routes
	for i := 1; i < len(dd.Paths); i++ {
		country = asNames.Country(dd.Paths[i])

		if len(country) == 0 {
			continue
		}

		if country != firstCountry {

			// If country is in monitor list then alert
			if d.CheckMonitorCountryCode(country) == true {

				printAlert(PriorityHigh, dd.Timestamp.String(), dd.Paths[i], convertAsPath(dd.Paths), "Monitored Country",
					fmt.Sprintf("Internal Route: %s\nExternal Country: %s", firstCountry, country))
				ret = true
				continue
			}

			// err = pool.QueryRow("get_route_count", firstAs, dd.PathsString).Scan(&count)
			// if err != nil {
			// 	if strings.Contains(err.Error(), "no rows in result set") == false {
			// 		fmt.Printf("Error retrieving 'get_route_count' count: %v", err)
			// 		continue
			// 	}

			// 	count = 0
			// 	ret = true
			// }

			count = history.GetRouteCount(firstAs, dd.PathsString)

			if count == 0 {
				printAlert(PriorityHigh, dd.Timestamp.String(), dd.PeerAs, dd.PathsString, "First Appearance", "")
				ret = true
				continue

			} else if count > 0 && count < 5 {
				printAlert(PriorityHigh, dd.Timestamp.String(), dd.PeerAs, dd.PathsString, "Low Frequency", "")
				ret = true
				continue

			} else if count > 5 && count < 10 {
				printAlert(PriorityHigh, dd.Timestamp.String(), dd.PeerAs, dd.PathsString, "Moderate Frequency", "")
				ret = true
				continue
			}
		}
	}

	return ret
}

//
func (d *Detector) isAnomlousPeer(dd *DetectData) bool {

	// Get the last AS and check if it is one of ours, if so exit, else continue checking our prefixes
	lastAs := dd.Paths[len(dd.Paths)-1]
	if d.CheckTargetAs(lastAs) == true {
		return false
	}

	ret := false

	// Is one of the prefixes one of ours, if so then alert
	for _, n := range dd.NLRI {

		if d.CheckPrefix(n) == true {
			printAlert(PriorityHigh, dd.Timestamp.String(), dd.Paths[0], dd.PathsString, "Invalid Prefix Peer",
				fmt.Sprintf("Prefix: %s", n))

			ret = true
		}
	}

	return ret
}

//
func (d *Detector) isLowFrequency(dd *DetectData) bool {

	count := history.GetRouteCount(dd.PeerAs, dd.PathsString)

	if count == 0 {
		printAlert(PriorityHigh, dd.Timestamp.String(), dd.PeerAs, dd.PathsString, "First Appearance", "")
		return true

	} else if count > 0 && count < 5 {
		printAlert(PriorityHigh, dd.Timestamp.String(), dd.PeerAs, dd.PathsString, "Low Frequency", "")
		return true

	} else if count > 5 && count < 10 {
		printAlert(PriorityHigh, dd.Timestamp.String(), dd.PeerAs, dd.PathsString, "Moderate Frequency", "")
		return true
	}

	return false
}
