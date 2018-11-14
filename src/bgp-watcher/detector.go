package main

import (
	"fmt"
	"net"
	"time"

	bgp "github.com/osrg/gobgp/packet"
)

// ##### Structs ##############################################################

type DetectData struct {
	Timestamp time.Time
	As        uint32
	PeerIP    net.IP
	Paths     []uint32
	NLRI      []*bgp.IPAddrPrefix
}

type Detector struct {
	asNames  *AsNames
	queue    chan *DetectData
	targetAs map[uint32]struct{}
	prefixes map[*bgp.IPAddrPrefix]struct{}
}

// ##### Methods ##############################################################

func NewDetector(an *AsNames) *Detector {

	d := new(Detector)
	d.asNames = an
	d.initialise()

	return d
}

//
func (d *Detector) initialise() {

	d.queue = make(chan *DetectData)
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
func (d *Detector) CheckTargetAs(as uint32) bool {

	if _, ok := d.targetAs[as]; ok {
		return true
	}

	return false
}

func (d *Detector) CheckPrefix(prefix *bgp.IPAddrPrefix) bool {

	if _, ok := d.prefixes[prefix]; ok {
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

			//fmt.Printf("DETECT1 %v\n", ud.As)
			//fmt.Printf("DETECT 2%v\n", ud.PeerIP)
			//fmt.Printf("DETECT3 %v\n", ud.NLRI)
			//fmt.Printf("DETECT4 %v\n", ud.Paths)

			go d.detect(dd)
			dd = nil
		}
	}()
}

//
func (d *Detector) Add(timestamp time.Time, as uint32, peerIP net.IP, paths []uint32, nlri []*bgp.IPAddrPrefix) {

	d.queue <- &DetectData{Timestamp: timestamp, As: as, PeerIP: peerIP, Paths: paths, NLRI: nlri}
}

//
func (d *Detector) detect(dd *DetectData) {

	d.detectAnomlousCountry(dd)
}

//
func (d *Detector) detectAnomlousCountry(dd *DetectData) {

	// If the path length equals two then no middle AS
	if len(dd.Paths) == 2 {
		return
	}

	firstAs := dd.Paths[0]
	firstCountry := d.asNames.Country(uint32(firstAs))

	// Get last AS Country
	lastAs := dd.Paths[len(dd.Paths)-1]
	lastCountry := d.asNames.Country(lastAs)

	// If the AS countries are the same then we cannot really check the middle routes
	if firstCountry != lastCountry {
		return
	}

	// Check the country of the intermediary routes
	var country string
	for i := 1; i < len(dd.Paths); i++ {
		country = d.asNames.Country(dd.Paths[i])

		if len(country) == 0 {
			continue
		}

		if country != firstCountry {
			fmt.Printf("ALERT ALERT %v : %v # %v # %v # %v\n", dd.Timestamp, firstCountry, country, dd.Paths[i], dd.Paths)
		}
	}
}

//
func (d *Detector) detectAnomlousPeer(dd *DetectData) {

	// Get last AS Country
	lastAs := dd.Paths[len(dd.Paths)-1]
	correctTarget := d.CheckTargetAs(lastAs)
	//lastCountry := d.asNames.Country(lastAs)

	// Is one of the prefixes one of ours
	for _, n := range dd.NLRI {

		if d.CheckPrefix(n) == false {
			fmt.Printf("ALERT PRERFEXCX %v : %v # %v\n", dd.Timestamp, correctTarget, n)
		}
	}
}
