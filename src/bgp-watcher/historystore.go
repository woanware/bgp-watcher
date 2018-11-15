package main

import (
	"fmt"
	"sync"

	pgx "github.com/jackc/pgx"
)

//
type HistoryStore struct {
	mux  sync.Mutex
	data map[uint32]map[string]uint64
}

//
func (hs *HistoryStore) Set(as uint32, route string) {

	hs.mux.Lock()
	if hs.data[as] == nil {
		hs.data[as] = make(map[string]uint64)
	}
	hs.data[as][route]++
	defer hs.mux.Unlock()
}

//
func (hs *HistoryStore) SetCount(as uint32, route string, count uint64) {

	hs.mux.Lock()
	if hs.data[as] == nil {
		hs.data[as] = make(map[string]uint64)
	}
	hs.data[as][route] = count
	defer hs.mux.Unlock()
}

//
func (hs *HistoryStore) SetAdd(as uint32, route string, count uint64) {

	hs.mux.Lock()
	if hs.data[as] == nil {
		hs.data[as] = make(map[string]uint64)
	}
	hs.data[as][route] += count
	defer hs.mux.Unlock()
}

//
func (hs *HistoryStore) Persist() {

	// Truncate table
	_, err := pool.Exec("truncate table routes")
	if err != nil {
		fmt.Printf("Error truncating historic data: %v\n", err)
	}

	var rows [][]interface{}

	for peer, a := range hs.data {
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
