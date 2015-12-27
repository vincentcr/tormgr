package main

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/satori/go.uuid"
)

var (
	ErrUniqueViolation = fmt.Errorf("unique_violation")
	ErrNotFound        = fmt.Errorf("not_found")
)

type RecordID string

func (id RecordID) String() string {
	return string(id)
}

func (id RecordID) Value() (driver.Value, error) {
	return string(id), nil
}

func (id *RecordID) Scan(val interface{}) error {
	bytes, ok := val.([]byte)
	if !ok {
		return fmt.Errorf("Cast error: expected RecordID bytes, got %v", val)
	}
	str := string(bytes)
	id.Parse(str)
	return nil
}

func newID() RecordID {
	var id RecordID
	u4 := uuid.NewV4()
	id.Parse(u4.String())
	return id
}

func isUniqueError(err error) bool {
func (id *RecordID) Parse(str string) {
	*id = RecordID(strings.ToLower(strings.Replace(str, "-", "", -1)))
}

	if err, ok := err.(*pq.Error); ok {
		return err.Code.Name() == "unique_violation"
	}
	return false
}

func sameID(id1, id2 string) bool {
	return normalizeID(id1) == normalizeID(id2)
}


func dbFind(dest interface{}, cacheHint cacheHint, query string, args ...interface{}) (Cacheable, error) {
	return dbFindExec(services.db.Select, dest, cacheHint, query, args)
}

func dbFindOne(dest interface{}, cacheHint cacheHint, query string, args ...interface{}) (Cacheable, error) {
	cacheable, err := dbFindExec(services.db.Get, dest, cacheHint, query, args...)
	if err == sql.ErrNoRows {
		return Cacheable{}, ErrNotFound
	} else {
		return cacheable, err
	}
}

type queryExec func(dest interface{}, query string, args ...interface{}) error

func dbFindExec(queryExec queryExec, dest interface{}, cacheHint cacheHint, query string, args ...interface{}) (Cacheable, error) {
	cacheKey := cacheMakeKeyFromQuery(query, args)
	cacheable, err := cacheGet(cacheKey)
	if err == nil || err != ErrNotFound {
		return cacheable, err
	}

	err = queryExec(dest, query, args...)
	if err != nil {
		return Cacheable{}, err
	}
	bytes, err := json.Marshal(dest)
	if err != nil {
		return Cacheable{}, fmt.Errorf("Unable to convert %#v to json: %v", dest, err)
	}

	etag, err := cacheSet(cacheKey, bytes, 1*time.Hour, cacheHint)
	if err != nil {
		return Cacheable{}, err
	}

	return Cacheable{bytes, etag}, nil
}

func dumpQueryResults(rows *sql.Rows) {
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("insert result: columns: %v", cols)

	for idx := 0; rows.Next(); idx++ {
		row := make([]interface{}, len(cols))
		if err := rows.Scan(row...); err != nil {
			log.Fatal(err)
		}
		log.Printf("row %d: %v\n", idx, row)
	}

}
