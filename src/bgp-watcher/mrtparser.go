package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"log"
	"os"

	bgp "github.com/osrg/gobgp/packet"
)

// // BGPDump encapuslates downloading and importing of BGP dumps.
// type BGPDump struct {
// 	Date time.Time
// }

// // Path returns the absolute path to the target archive dump download file.
// func (b *BGPDump) Path() string {
// 	return filepath.Join(
// 		b.dir(), fmt.Sprintf("%s.gz", b.Date.Format("20060102")))
// }

// // Path returns the absolute path to the target archive dump download file.
// func (b *BGPDump) ParsedPath() string {
// 	return filepath.Join(
// 		b.dir(), fmt.Sprintf("%s.json", b.Date.Format("20060102")))
// }

// func (b *BGPDump) dir() string {
// 	return filepath.Join(
// 		dataDir, "cache", b.Date.Format("200601"))
// }

// func (b *BGPDump) day() string {
// 	return b.Date.Format("20060102")
// }

type MrtParser struct {
}

//
func (b *MrtParser) Parse(asns map[uint32]map[string]uint64, filePath string) (map[uint32]map[string]uint64, error) {

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
	count := 0

	// indexTableCount := 0
	// db := make(map[string]uint32, 0)

	var last uint32

entries:
	for scanner.Scan() {
		count++
		data := scanner.Bytes()

		hdr := &bgp.MRTHeader{}
		errh := hdr.DecodeFromBytes(data[:bgp.MRT_COMMON_HEADER_LEN])
		if err != nil {
			return asns, errh
		}

		msg, err := bgp.ParseMRTBody(hdr, data[bgp.MRT_COMMON_HEADER_LEN:])
		if err != nil {
			log.Printf("could not parse mrt body: %v", err)
			continue entries
		}

		switch msg.Body.(type) {
		case *bgp.BGP4MPMessage:

			bgp4mp := msg.Body.(*bgp.BGP4MPMessage)

			switch bgp4mp.BGPMessage.Body.(type) {
			case *bgp.BGPUpdate:

				bgpUpdate := bgp4mp.BGPMessage.Body.(*bgp.BGPUpdate)

				// for a, b := range bgpUpdate.WithdrawnRoutes {
				// 	fmt.Printf("%v:%v\n", a, b)0
				// }

				// for a, b := range bgpUpdate.NLRI {
				// 	fmt.Printf("%v:%v\n", a, b)
				// }

				for _, pa := range bgpUpdate.PathAttributes {

					if pa.GetType() != bgp.BGP_ATTR_TYPE_AS_PATH {
						continue
					}

					paAsPath := pa.(*bgp.PathAttributeAsPath)

					for _, asValue := range paAsPath.Value {

						switch asValue.(type) {
						case *bgp.As4PathParam:
							//temp1 := asValue.(*bgp.As4PathParam).AS
							//fmt.Printf("LAST: %v\n", temp1[len(temp1)-1])

							//last = temp1[len(temp1)-1]

							last = asValue.(*bgp.As4PathParam).AS[len(asValue.(*bgp.As4PathParam).AS)-1]

							if last == 34737 || last == 34178 {
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
