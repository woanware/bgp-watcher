package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"log"
	"os"

	bgp "github.com/osrg/gobgp/pkg/packet/bgp"
	mrt "github.com/osrg/gobgp/pkg/packet/mrt"
)

// ##### Structs ##############################################################

type MrtParser struct {
}

// ##### Methods ##############################################################

//
func (b *MrtParser) ParseAndCollect(detector *Detector, filePath string) error {

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}

	gzipReader, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("couldn't create gzip reader: %v", err)
	}

	scanner := bufio.NewScanner(gzipReader)
	scanner.Split(mrt.SplitMrt)

	var last uint32
	var data []byte
	var hdr *mrt.MRTHeader
	var msg *mrt.MRTMessage
	var bgp4mp *mrt.BGP4MPMessage
	var bgpUpdate *bgp.BGPUpdate
	var pa bgp.PathAttributeInterface
	var paAsPath *bgp.PathAttributeAsPath
	var asValue bgp.AsPathParamInterface

entries:
	for scanner.Scan() {
		data = scanner.Bytes()

		hdr = &mrt.MRTHeader{}
		err = hdr.DecodeFromBytes(data[:mrt.MRT_COMMON_HEADER_LEN])
		if err != nil {
			return err
		}

		msg, err = mrt.ParseMRTBody(hdr, data[mrt.MRT_COMMON_HEADER_LEN:])
		if err != nil {
			log.Printf("could not parse mrt body: %v", err)
			continue entries
		}

		switch msg.Body.(type) {
		case *mrt.BGP4MPMessage:

			bgp4mp = msg.Body.(*mrt.BGP4MPMessage)

			switch bgp4mp.BGPMessage.Body.(type) {
			case *bgp.BGPUpdate:

				bgpUpdate = bgp4mp.BGPMessage.Body.(*bgp.BGPUpdate)

				// for a, b := range bgpUpdate.WithdrawnRoutes {
				// 	fmt.Printf("%v:%v\n", a, b)0
				// }

				for _, pa = range bgpUpdate.PathAttributes {

					if pa.GetType() != bgp.BGP_ATTR_TYPE_AS_PATH {
						continue
					}

					paAsPath = pa.(*bgp.PathAttributeAsPath)

					for _, asValue = range paAsPath.Value {

						switch asValue.(type) {
						case *bgp.As4PathParam:
							//temp1 := asValue.(*bgp.As4PathParam).AS
							//fmt.Printf("LAST: %v\n", temp1[len(temp1)-1])

							//last = temp1[len(temp1)-1]

							last = asValue.(*bgp.As4PathParam).AS[len(asValue.(*bgp.As4PathParam).AS)-1]

							// Is the last part of the path one of ours
							if detector.CheckTargetAs(last) == true {

								history.Set(bgp4mp.PeerAS, asValue.String())
								// if asns[bgp4mp.PeerAS] == nil {
								// 	asns[bgp4mp.PeerAS] = make(map[string]uint64)
								// }
								// asns[bgp4mp.PeerAS][asValue.String()]++
								continue entries
							}

							// if last == 34178 {

							// 	fmt.Println(bgp4mp.String())
							// 	for _, b := range bgpUpdate.NLRI {bgp

							// 		// if d.CheckPrefix(b) == false {
							// 		// 	continue
							// 		// }

							// 		fmt.Printf("%v\n", b)
							// 	}
							// 	fmt.Printf("-----------------------\n")

							// 	//fmt.Println(bgp4mp.String())
							// 	//fmt.Println("QINETIQ")
							// 	//fmt.Printf("%v\n", asValue)
							// 	//fmt.Printf("-----------------------\n")

							// 	if asns[bgp4mp.PeerAS] == nil {
							// 		asns[bgp4mp.PeerAS] = make(map[string]uint64)
							// 	}
							// 	asns[bgp4mp.PeerAS][asValue.String()]++

							// 	// for _, e := range asValue.(*bgp.As4PathParam).AS {
							// 	// 	fmt.Printf("EE1 %v\n", e)
							// 	// 	//fmt.Printf("FF1 %v\n", f)
							// 	// }
							// }
							//fmt.Printf("LAST2: %v", asValue.(*bgp.As4PathParam).AS[len(asValue.(*bgp.As4PathParam).AS)-1])
							// for _, e := range asValue.(*bgp.As4PathParam).AS {
							// 	fmt.Printf("EE1 %v\n", e)
							// 	//fmt.Printf("FF1 %v\n", f)
							// }
							break

						case *bgp.AsPathParam:
							//temp1 := asValue.(*bgp.AsPathParam).AS
							//fmt.Printf("LAST: %v", asValue.(*bgp.AsPathParam).AS[len(asValue.(*bgp.AsPathParam).AS)-1])

							break
						}
					}
				}

				// case *mrt.PeerIndexTable:
				// 	// IGNORED

				// case *bgp.Rib:
				// 	// IGNORED

				// case *mrt.BGP4MPStateChange:
				// 	// IGNORED

				// case *bgp.BGPKeepAlive:
				// 	// IGNORED
			}
		}
	}

	return nil
}

//
func (b *MrtParser) ParseAndDetect(detector *Detector, name string, filePath string) (*History, error) {

	history := &History{data: make(map[uint32]map[string]uint64)}

	f, err := os.Open(filePath)
	if err != nil {
		return history, err
	}

	gzipReader, err := gzip.NewReader(f)
	if err != nil {
		return history, fmt.Errorf("couldn't create gzip reader: %v", err)
	}

	scanner := bufio.NewScanner(gzipReader)
	scanner.Split(mrt.SplitMrt)

	var last uint32
	var data []byte
	var hdr *mrt.MRTHeader
	var msg *mrt.MRTMessage
	var bgp4mp *mrt.BGP4MPMessage
	var bgpUpdate *bgp.BGPUpdate
	var pa bgp.PathAttributeInterface
	var paAsPath *bgp.PathAttributeAsPath
	var asValue bgp.AsPathParamInterface

entries:
	for scanner.Scan() {
		data = scanner.Bytes()

		hdr = &mrt.MRTHeader{}
		err = hdr.DecodeFromBytes(data[:mrt.MRT_COMMON_HEADER_LEN])
		if err != nil {
			return history, err
		}

		msg, err = mrt.ParseMRTBody(hdr, data[mrt.MRT_COMMON_HEADER_LEN:])
		if err != nil {
			log.Printf("could not parse mrt body: %v", err)
			continue entries
		}

		switch msg.Body.(type) {
		case *mrt.BGP4MPMessage:

			bgp4mp = msg.Body.(*mrt.BGP4MPMessage)

			switch bgp4mp.BGPMessage.Body.(type) {
			//case *bgp.PeerIndexTable:
			// IGNORED

			//case *mrt.Rib:
			// IGNORED

			//case *mrt.BGP4MPStateChange:
			// IGNORED

			//case *bgp.BGPKeepAlive:
			// IGNORED

			case *bgp.BGPUpdate:

				bgpUpdate = bgp4mp.BGPMessage.Body.(*bgp.BGPUpdate)

				// for a, b := range bgpUpdate.WithdrawnRoutes {
				// 	fmt.Printf("%v:%v\n", a, b)0
				// }

				for _, pa = range bgpUpdate.PathAttributes {

					if pa.GetType() != bgp.BGP_ATTR_TYPE_AS_PATH {
						continue
					}

					paAsPath = pa.(*bgp.PathAttributeAsPath)

					for _, asValue = range paAsPath.Value {

						switch asValue.(type) {
						case *bgp.As4PathParam:

							last = asValue.(*bgp.As4PathParam).AS[len(asValue.(*bgp.As4PathParam).AS)-1]

							// Is the last part of the path one of ours
							if detector.CheckTargetAs(last) == true {

								//historyStore.Set(bgp4mp.PeerAS, asValue.String())

								//fmt.Println(bgp4mp.String())
								detector.Add(name, hdr.GetTime(), bgp4mp.PeerAS, bgp4mp.PeerIpAddress,
									asValue.(*bgp.As4PathParam).String(), asValue.(*bgp.As4PathParam).AS, bgpUpdate.NLRI)
								continue entries
							}

							// Is one of the prefixes one of ours
							for _, b := range bgpUpdate.NLRI {

								if detector.CheckPrefix(b) == false {
									continue
								}

								//fmt.Println(bgp4mp.String())

								detector.Add(name, hdr.GetTime(), bgp4mp.PeerAS, bgp4mp.PeerIpAddress,
									asValue.(*bgp.As4PathParam).String(), asValue.(*bgp.As4PathParam).AS, bgpUpdate.NLRI)
								continue entries
							}

							break

						case *bgp.AsPathParam:
							//temp1 := asValue.(*bgp.AsPathParam).AS
							//fmt.Printf("LAST: %v", asValue.(*bgp.AsPathParam).AS[len(asValue.(*bgp.AsPathParam).AS)-1])

							break
						}
					}
				}
			}
		}
	}

	return history, nil
}
