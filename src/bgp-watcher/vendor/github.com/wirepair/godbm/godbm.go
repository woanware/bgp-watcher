/*
The MIT License (MIT)

Copyright (c) 2016 isaac dawson

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

// A simple postgresql database manager/helper.
package godbm

import (
	"database/sql"
	"github.com/lib/pq"
	"sync"
)

// SqlStorer interface
type SqlStorer interface {
	Connect() error
	Disconnect() error
	Exec(query string, data ...interface{}) (sql.Result, error)
	Query(query string, data ...interface{}) (*sql.Rows, error)
	PrepareStatement(query string) (*sql.Stmt, error)
	PrepareAdd(key, query string) error
	HasStatement(key string) bool
	PrepareDel(key string) error
	QueryPrepared(key string, data ...interface{}) (*sql.Rows, error)
	ExecPrepared(key string, data ...interface{}) (sql.Result, error)
}

// UnknownStmtError holds the invalid key which was attempted in a look up.
type UnknownStmtError struct {
	StmtKey string // description of key
}

// Returned when the supplied key for looking up a prepared statement does not exist.
func (e *UnknownStmtError) Error() string {
	return "godbm: error " + e.StmtKey + " was not found"
}

// ConnectionError
type ConnectionError struct{}

// Returned when the supplied key for looking up a prepared statement does not exist.
func (e *ConnectionError) Error() string {
	return "godbm: error not connected to the database"
}

// SqlStore holds a reference to the database, a list of prepared statements
// and a boolean for if we are connected.
type SqlStore struct {
	sync.RWMutex                      // a mutex to synchronize adding/calling/removing new statements.
	Connected    bool                 // indicates if we are connected or not.
	db           *sql.DB              // the underlying database reference
	queries      map[string]*sql.Stmt // a map of prepared statements referenced by the key
	username     string               // database username
	password     string               // database password
	dbname       string               // database name to connect to
	host         string               // database host
	sslmode      string               // sslmode one of: require, verify-full, verify-ca, disable. (check postgres docs for more)
	opts         string               // add your own options.
}

// New creates a new *SqlStore with the connection properties as arguments.
func New(username, password, dbname, host, sslmode, opts string) *SqlStore {
	s := new(SqlStore)
	s.username = username
	s.password = password
	s.host = host
	s.dbname = dbname
	s.sslmode = sslmode
	return s
}

// Connect connects to the database. Returns err on sql.Open error or sets
// our connected state to true.
func (store *SqlStore) Connect() (err error) {
	store.Connected = false
	store.db, err = sql.Open("postgres", "user="+store.username+" password="+store.password+" dbname="+store.dbname+" host="+store.host+" sslmode="+store.sslmode+" "+store.opts)
	if err != nil {
		return err
	}
	store.Connected = true
	return err
}

// Disconnect iterates through any prepared statements and closes them then calls close
// on the db driver.
func (store *SqlStore) Disconnect() (err error) {
	for _, v := range store.queries {
		v.Close()
	}
	err = store.db.Close()
	store.Connected = false
	return err
}

// Exec creates a new prepared statement, executes and closes. Takes a query string as the first
// parameter and a variable number of arguments to be used in the statement. Closes the statement
// when finished and returns a sql.Result. You should only use this for testing as creating new
// statements every time is non-performant.
func (store *SqlStore) Exec(query string, data ...interface{}) (results sql.Result, err error) {
	if !store.Connected {
		return nil, &ConnectionError{}
	}

	stmt, err := store.PrepareStatement(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	return stmt.Exec(data...)

}

// Query creates a new prepared statement, executes and closes. Takes a query string as the first
// parameter and a variable number of arguments to be used in the statement. Closes the statement
// when finished and returns *sql.Rows if any. You should only use this for testing as creating new
// statements every time is non-performant.
func (store *SqlStore) Query(query string, data ...interface{}) (results *sql.Rows, err error) {
	if !store.Connected {
		return nil, &ConnectionError{}
	}

	stmt, err := store.PrepareStatement(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	return stmt.Query(data...)
}

// PrepareStatement prepares a query and returns the statement to the caller, or error
// if it is invalid.
func (store *SqlStore) PrepareStatement(query string) (stmt *sql.Stmt, err error) {
	if !store.Connected {
		return nil, &ConnectionError{}
	}

	stmt, err = store.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	return stmt, nil
}

// PrepareAdd creates a prepared statement and safely adds it to our map with the provided key.
func (store *SqlStore) PrepareAdd(key, query string) (err error) {
	if !store.Connected {
		return &ConnectionError{}
	}

	stmt, err := store.PrepareStatement(query)
	if err != nil {
		return err
	}
	defer store.Unlock()

	store.Lock()
	if store.queries != nil {
		store.queries[key] = stmt
	} else {
		store.queries = map[string]*sql.Stmt{key: stmt}
	}
	return nil
}

// PrepareDel safely removes a prepared statement from our store provided it exists.
func (store *SqlStore) PrepareDel(key string) (err error) {
	if !store.Connected {
		return &ConnectionError{}
	}
	defer store.Unlock()

	store.Lock()
	stmt, found := store.queries[key]
	if !found {
		return nil
	}
	err = stmt.Close()
	delete(store.queries, key)
	return err
}

// returns true if the statement has been added
func (store *SqlStore) HasStatement(key string) bool {
	store.RLock()
	_, found := store.queries[key]
	store.RUnlock()
	return found
}

// QueryPrepared executes a prepared statement which is looked up by the provided key. If the key was
// not found, an UnknownStmtError is returned. This method takes a variable number of arguments to
// pass to the underlying statement and returns *sql.Rows or an error.
func (store *SqlStore) QueryPrepared(key string, data ...interface{}) (rows *sql.Rows, err error) {
	if !store.Connected {
		return nil, &ConnectionError{}
	}
	defer store.RUnlock()

	store.RLock()
	stmt, found := store.queries[key]
	if !found {
		return nil, &UnknownStmtError{StmtKey: key}
	}
	return stmt.Query(data...)
}

// ExecPrepared executes a prepared statement which is looked up by the provided key. If the key was
// not found, an UnknownStmtError is returned. This method takes a variable number of arguments to
// pass to the underlying statement and returns sql.Result or an error.
func (store *SqlStore) ExecPrepared(key string, data ...interface{}) (result sql.Result, err error) {
	if !store.Connected {
		return nil, &ConnectionError{}
	}
	defer store.RUnlock()

	store.RLock()
	stmt, found := store.queries[key]
	if !found {
		return nil, &UnknownStmtError{StmtKey: key}
	}
	return stmt.Exec(data...)
}

// CopyStart opens up a transaction for us with the provided table and column names. Returns the transaction
// which we'll need to pass back to CopyCommit or CopyCancel along with the statement. The statement is also
// returned so you can Exec your inserts in a loop or however you want.
func (store *SqlStore) CopyStart(table string, columns ...string) (txn *sql.Tx, stmt *sql.Stmt, err error) {
	if !store.Connected {
		return nil, nil, &ConnectionError{}
	}

	txn, err = store.db.Begin()
	if err != nil {
		return nil, nil, err
	}
	stmt, err = store.copyStart(txn, table, columns...)
	return txn, stmt, err
}

// Same as above but uses the provided transaction that was already opened by the caller
func (store *SqlStore) CopyStartWithTxn(txn *sql.Tx, table string, columns ...string) (stmt *sql.Stmt, err error) {
	if !store.Connected {
		return nil, &ConnectionError{}
	}
	return store.copyStart(txn, table, columns...)
}

// Prepares the transaction for pq.CopyIn.
func (store *SqlStore) copyStart(txn *sql.Tx, table string, columns ...string) (stmt *sql.Stmt, err error) {
	stmt, err = txn.Prepare(pq.CopyIn(table, columns...))
	if err != nil {
		return nil, err
	}
	return stmt, nil
}

// CopyCommit takes the transaction with the statement that you added your inserts, at this point it
// is still open and waiting to be commited to the server (along with the inserts that were bulk loaded).
func (store *SqlStore) CopyCommit(txn *sql.Tx, stmt *sql.Stmt) error {
	if _, err := stmt.Exec(); err != nil {
		return err
	}

	if err := stmt.Close(); err != nil {
		return err
	}

	if err := txn.Commit(); err != nil {
		return err
	}
	return nil
}

// CopyCancel rolls back the transaction
func (store *SqlStore) CopyCancel(txn *sql.Tx, stmt *sql.Stmt) error {
	if err := stmt.Close(); err != nil {
		return err
	}
	return txn.Rollback()
}

// Allow access to underlying DB so user can create custom transactions.
func (store *SqlStore) Db() *sql.DB {
	return store.db
}
