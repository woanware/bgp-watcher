package main

import (
	"fmt"
	"net"

	bgp "github.com/osrg/gobgp/packet"
)

type DetectData struct {
	As     uint32
	PeerIP net.IP
	Paths  []uint32
	NLRI   []*bgp.IPAddrPrefix
}

type Detector struct {
	asNames  *AsNames
	queue    chan *UpdateData
	targetAs map[uint32]struct{}
	prefixes map[*bgp.IPAddrPrefix]struct{}
}

func NewDetector(an *AsNames) *Detector {

	d := new(Detector)
	d.asNames = an
	d.initialise()

	return d
}

//
func (d *Detector) initialise() {

	d.queue = make(chan *UpdateData)
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

			fmt.Printf("DETECT1 %v\n", ud.As)
			fmt.Printf("DETECT 2%v\n", ud.PeerIP)
			fmt.Printf("DETECT3 %v\n", ud.NLRI)
			fmt.Printf("DETECT4 %v\n", ud.Paths)

			go detect(dd)
			dd = nil
		}
	}()
}

//
func (d *Detector) Add(as uint32, peerIP net.IP, paths []uint32, nlri []*bgp.IPAddrPrefix) {

	d.queue <- &UpdateData{As: as, PeerIP: peerIP, Paths: paths, NLRI: nlri}
}

func (d *Detector) detect(dd *DetectData) {

}

func (d *Detector) detectAnomlousCountry(dd *DetectData) {

	// Get first AS Country

	// Get last AS Country

	// Check the country of the intermedant routes
}
