package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"log"
	"os"

	bgp "github.com/osrg/gobgp/packet"
)

// ##### Structs ##############################################################

type MrtParser struct {
}

// ##### Methods ##############################################################

//
func (b *MrtParser) ParseAndCollect(asns map[uint32]map[string]uint64, filePath string) (map[uint32]map[string]uint64, error) {

	//asns := make(map[uint32]map[string]uint64)

	f, err := os.Open(filePath)
	if err != nil {
		return asns, err
	}

	gzipReader, err := gzip.NewReader(f)
	if err != nil {
		return asns, fmt.Errorf("couldn't create gzip reader: %v", err)
	}

	scanner := bufio.NewScanner(gzipReader)
	scanner.Split(bgp.SplitMrt)

	var last uint32
	var data []byte
	var hdr *bgp.MRTHeader
	var msg *bgp.MRTMessage
	var bgp4mp *bgp.BGP4MPMessage
	var bgpUpdate *bgp.BGPUpdate
	var pa bgp.PathAttributeInterface
	var paAsPath *bgp.PathAttributeAsPath
	var asValue bgp.AsPathParamInterface

entries:
	for scanner.Scan() {
		data = scanner.Bytes()

		hdr = &bgp.MRTHeader{}
		err = hdr.DecodeFromBytes(data[:bgp.MRT_COMMON_HEADER_LEN])
		if err != nil {
			return asns, err
		}

		msg, err = bgp.ParseMRTBody(hdr, data[bgp.MRT_COMMON_HEADER_LEN:])
		if err != nil {
			log.Printf("could not parse mrt body: %v", err)
			continue entries
		}

		switch msg.Body.(type) {
		case *bgp.BGP4MPMessage:

			bgp4mp = msg.Body.(*bgp.BGP4MPMessage)

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

							if last == 34178 {

								fmt.Println(bgp4mp.String())
								for _, b := range bgpUpdate.NLRI {

									// if d.CheckPrefix(b) == false {
									// 	continue
									// }

									fmt.Printf("%v\n", b)
								}
								fmt.Printf("-----------------------\n")

								//fmt.Println(bgp4mp.String())
								//fmt.Println("QINETIQ")
								//fmt.Printf("%v\n", asValue)
								//fmt.Printf("-----------------------\n")

								if asns[bgp4mp.PeerAS] == nil {
									asns[bgp4mp.PeerAS] = make(map[string]uint64)
								}
								asns[bgp4mp.PeerAS][asValue.String()]++

								// for _, e := range asValue.(*bgp.As4PathParam).AS {
								// 	fmt.Printf("EE1 %v\n", e)
								// 	//fmt.Printf("FF1 %v\n", f)
								// }
							}
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

			case *bgp.PeerIndexTable:
				// IGNORED

			case *bgp.Rib:
				// IGNORED

			case *bgp.BGP4MPStateChange:
				// IGNORED

			case *bgp.BGPKeepAlive:
				// IGNORED
			}
		}
	}

	return asns, nil
}

//
func (b *MrtParser) Parse(detector *Detector, filePath string) error {

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}

	gzipReader, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("couldn't create gzip reader: %v", err)
	}

	scanner := bufio.NewScanner(gzipReader)
	scanner.Split(bgp.SplitMrt)

	var last uint32
	var data []byte
	var hdr *bgp.MRTHeader
	var msg *bgp.MRTMessage
	var bgp4mp *bgp.BGP4MPMessage
	var bgpUpdate *bgp.BGPUpdate
	var pa bgp.PathAttributeInterface
	var paAsPath *bgp.PathAttributeAsPath
	var asValue bgp.AsPathParamInterface

entries:
	for scanner.Scan() {
		data = scanner.Bytes()

		hdr = &bgp.MRTHeader{}
		err = hdr.DecodeFromBytes(data[:bgp.MRT_COMMON_HEADER_LEN])
		if err != nil {
			return err
		}

		msg, err = bgp.ParseMRTBody(hdr, data[bgp.MRT_COMMON_HEADER_LEN:])
		if err != nil {
			log.Printf("could not parse mrt body: %v", err)
			continue entries
		}

		switch msg.Body.(type) {
		case *bgp.BGP4MPMessage:

			bgp4mp = msg.Body.(*bgp.BGP4MPMessage)

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
								//fmt.Println(bgp4mp.String())
								detector.Add(hdr.GetTime(), bgp4mp.PeerAS, bgp4mp.PeerIpAddress, asValue.(*bgp.As4PathParam).AS, bgpUpdate.NLRI)
								continue entries
							}

							// Is one of the prefixes one of ours
							for _, b := range bgpUpdate.NLRI {

								if detector.CheckPrefix(b) == false {
									continue
								}

								//fmt.Println(bgp4mp.String())

								detector.Add(hdr.GetTime(), bgp4mp.PeerAS, bgp4mp.PeerIpAddress, asValue.(*bgp.As4PathParam).AS, bgpUpdate.NLRI)
								continue entries
							}

							// if last == 34178 {
							// 	fmt.Println("ADD")
							// 	detector.Add(bgp4mp.PeerAS, bgp4mp.PeerIpAddress, asValue.(*bgp.As4PathParam).AS)

							// 	for _, b := range bgpUpdate.NLRI {

							// 		if detector.CheckPrefix(b) == false {
							// 			continue
							// 		}

							// 		fmt.Printf("%v\n", b)
							// 	}
							// 	fmt.Printf("-----------------------\n")

							// 	//fmt.Println(bgp4mp.String())
							// 	//fmt.Println("QINETIQ")
							// 	//fmt.Printf("%v\n", asValue)
							// 	//fmt.Printf("-----------------------\n")

							// 	// if asns[bgp4mp.PeerAS] == nil {
							// 	// 	asns[bgp4mp.PeerAS] = make(map[string]uint64)
							// 	// }
							// 	// asns[bgp4mp.PeerAS][asValue.String()]++

							// 	// for _, e := range asValue.(*bgp.As4PathParam).AS {
							// 	// 	fmt.Printf("EE1 %v\n", e)
							// 	// 	//fmt.Printf("FF1 %v\n", e.GetType())
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

			case *bgp.PeerIndexTable:
				// IGNORED

			case *bgp.Rib:
				// IGNORED

			case *bgp.BGP4MPStateChange:
				// IGNORED

			case *bgp.BGPKeepAlive:
				// IGNORED
			}
		}
	}

	return nil
}
