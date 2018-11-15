package main

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	pgx "github.com/jackc/pgx"
)

//
type History struct {
	mux  sync.Mutex
	data map[uint32]map[string]uint64
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
	if h.data[as] == nil {
		h.data[as] = make(map[string]uint64)
	}
	h.data[as][route] = count
	defer h.mux.Unlock()
}

//
func (h *History) SetAdd(as uint32, route string, count uint64) {

	h.mux.Lock()
	if h.data[as] == nil {
		h.data[as] = make(map[string]uint64)
	}
	h.data[as][route] += count
	defer h.mux.Unlock()
}

//
func (h *History) Persist() {

	// Truncate table
	_, err := pool.Exec("truncate table routes")
	if err != nil {
		fmt.Printf("Error truncating historic data: %v\n", err)
	}

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
}

//
func (h *History) Summary() {

	var parts []string
	var part string
	var peerAs uint32
	var b bytes.Buffer
	var route string
	//var err error

	for _, a := range h.data {
		for route = range a {
			parts = strings.Split(route, " ")

			for _, part = range parts {
				b.WriteString(part)
				b.WriteString("(")
				peerAs, _ = ConvertStringToUint32(part)
				b.WriteString(asNames.Country(peerAs))
				b.WriteString(") # ")
			}

			fmt.Println(b.String())

			//rows = append(rows, []interface{}{peer, route, count})
		}
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
