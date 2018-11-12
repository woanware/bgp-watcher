# godbm 
This library is a simple wrapper/helper for postgresql databases using the pq driver (available from https://github.com/lib/pq).

### installation
go get github.com/wirepair/godbm

### example  

```Go
package main

import (
	"fmt"
	"github.com/wirepair/godbm"
	"log"
)

const (
	username = "postgres"
	password = "testpass"
	dbname   = "test"
	host     = "127.0.0.1"
)

func main() {
	dbm := godbm.New(username, password, dbname, host, "verify-full")
	if err := dbm.Connect(); err != nil {
		log.Fatalf("Error connecting to database: %v\n", err)
	}
	defer dbm.Disconnect()
	// a query with no arguments.
	dbm.PrepareAdd("getusername", "select * from user")
	rows, err := dbm.QueryPrepared("getusername")
	if err != nil {
		log.Fatalf("Error executing prepared statement: %v\n", err)
	}

	for rows.Next() {
		var user string
		err := rows.Scan(&user)
		if err != nil {
			log.Fatalf("Error getting username: %v\n", err)
		}
		if user != username {
			log.Fatalf("error returned username is different, we got: %v!", user)
		} else {
			fmt.Printf("got user %s back.\n", user)
		}
	}
	dbm.PrepareAdd("sleep", "select pg_sleep($1)")
	fmt.Print("sleeping for 3 seconds...")
	_, err = dbm.QueryPrepared("sleep", 3)
	if err != nil {
		log.Fatalf("Error sleeping: %v", err)
	}
	fmt.Print("Done.")
}
```

### more examples
See the tests!