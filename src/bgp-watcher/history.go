package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	pgx "github.com/jackc/pgx"
	util "github.com/woanware/goutil"
)

// ##### Structs ##############################################################

//
type History struct {
	mux  sync.Mutex
	data map[uint32]map[string]uint64
}

// ##### Methods ##############################################################

//
func NewHistory() *History {

	return &History{
		data: make(map[uint32]map[string]uint64),
	}
}

//
func (h *History) GetRouteCount(as uint32, route string) uint64 {

	h.mux.Lock()
	defer h.mux.Unlock()

	if h.data[as] == nil {
		return 0
	}

	return h.data[as][route]
}

//
func (h *History) Set(as uint32, route string) {

	h.mux.Lock()
	if h.data[as] == nil {
		h.data[as] = make(map[string]uint64)
	}
	h.data[as][route]++
	defer h.mux.Unlock()
}

//
func (h *History) SetCount(as uint32, route string, count uint64) {

	h.mux.Lock()
	defer h.mux.Unlock()
	if h.data[as] == nil {
		h.data[as] = make(map[string]uint64)
	}
	h.data[as][route] = count

}

//
func (h *History) SetAdd(as uint32, route string, count uint64) {

	h.mux.Lock()
	defer h.mux.Unlock()
	if h.data[as] == nil {
		h.data[as] = make(map[string]uint64)
	}
	h.data[as][route] += count
}

//
func (h *History) Persist() {

	// Truncate table
	_, err := pool.Exec("truncate table routes")
	if err != nil {
		fmt.Printf("Error truncating historic data: %v\n", err)
	}

	// Massage the data into a format that can be used with
	// the postgres COPY functionality e.g. fastest inserts
	var rows [][]interface{}
	for peer, a := range h.data {
		for route, count := range a {
			rows = append(rows, []interface{}{peer, route, count})
		}
	}

	_, err = pool.CopyFrom(
		pgx.Identifier{"routes"},
		[]string{"peer_as", "route", "count"},
		pgx.CopyFromRows(rows))

	if err != nil {
		fmt.Printf("Error inserting historic data: %v\n", err)
		return
	}

	// // FILE PERSISTANCE
	// h.mux.Lock()
	// defer h.mux.Unlock()

	// buffer := new(bytes.Buffer)
	// encoder := gob.NewEncoder(buffer)

	// err := encoder.Encode(h.data)
	// if err != nil {
	// 	fmt.Printf("Error encoding historic data: %v\n", err)
	// }

	// err = ioutil.WriteFile("./db/history.db", buffer.Bytes(), 0770)
	// if err != nil {
	// 	fmt.Printf("Error persisting historic data: %v\n", err)
	// }
}

//
func (h *History) Load() {

	rows, _ := pool.Query("select peer_as, route, count from routes")

	var peerAs uint32
	var route string
	var count uint64
	var err error

	for rows.Next() {
		err = rows.Scan(&peerAs, &route, &count)
		if err != nil {
			fmt.Printf("Error loading historic data: %v\n", err)
			continue
		}

		h.SetAdd(peerAs, route, count)
	}

	// // FILE PERSISTANCE
	// h.mux.Lock()
	// defer h.mux.Unlock()

	// data, err := ioutil.ReadFile("./db/history.db")
	// reader := bytes.NewReader(data)
	// if err != nil {
	// 	fmt.Printf("Error reading historic data: %v\n", err)
	// }

	// decoder := gob.NewDecoder(reader)
	// err = decoder.Decode(&h.data)
	// if err != nil {
	// 	fmt.Printf("Error loading historic data: %v\n", err)
	// }
}

//
func (h *History) Summary() {

	if util.DoesDirExist("./summary") == false {
		err := os.MkdirAll("./summary", 0770)
		if err != nil {
			fmt.Printf("Error creating summary directory: %v\n", err)
			return
		}
	}

	var parts []string
	var part string
	var peerAs uint32
	var b bytes.Buffer
	var route string
	var temp string
	var count uint64

	b.WriteString("SRC-DST, PEER_AS, COUNT, PATH1, PATH2, PATH3, PATH4, PATH5, PATH6, PATH6, PATH7, PATH8, PATH9, PATH10\n")

	for peer, a := range h.data {
		for route, count = range a {
			parts = strings.Split(route, " ")

			// Get originating peer details
			temp = parts[0]
			peerAs, _ = ConvertStringToUint32(temp)
			firstCountry := asNames.Country(uint32(peerAs))
			b.WriteString(firstCountry)
			b.WriteString("->")

			// Get destination peer details
			temp = parts[len(parts)-1]
			peerAs, _ = ConvertStringToUint32(temp)
			lastCountry := asNames.Country(peerAs)
			b.WriteString(lastCountry)
			b.WriteString(", ")
			b.WriteString(ConvertUInt32ToString(peer))
			b.WriteString(", ")
			b.WriteString(ConvertUint64ToString(count))
			b.WriteString(", ")

			for _, part = range parts {
				b.WriteString(part)
				b.WriteString(" (")
				peerAs, _ = ConvertStringToUint32(part)
				b.WriteString(asNames.Country(peerAs))
				b.WriteString("), ")
			}

			b.WriteString("\n")
		}
	}

	err := ioutil.WriteFile("./summary/summary.csv", b.Bytes(), 0770)
	if err != nil {
		fmt.Printf("Error writing summary data: %v\n", err)
	}
	// _, err = pool.CopyFrom(
	// 	pgx.Identifier{"routes"},
	// 	[]string{"peer_as", "route", "count"},
	// 	pgx.CopyFromRows(rows))

	// if err != nil {
	// 	fmt.Printf("Error inserting historic data: %v\n", err)
	// 	return
	// }
}
