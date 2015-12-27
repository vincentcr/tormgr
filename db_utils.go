package main

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
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

func (id *RecordID) Parse(str string) {
	*id = RecordID(strings.ToLower(strings.Replace(str, "-", "", -1)))
}

func dbExecOnRecord(desc, query string, record CacheHinter) error {
	r, err := services.db.NamedExec(query, record)

	if dbIsUniqueError(err) {
		return ErrUniqueViolation
	} else if err != nil {
		return fmt.Errorf("db:%s failed on %v: %v", record, err)
	}

	if err = dbCheckRowsAffected(r, 1); err != nil {
		return fmt.Errorf("db:%s failed on %v: %v", record, err)
	}

	if err := cacheInvalidate(record.cacheHint()); err != nil {
		return fmt.Errorf("db:%s failed on %v: %v", record, err)
	}

	return nil
}

func dbIsUniqueError(err error) bool {
	if err, ok := err.(*pq.Error); ok {
		return err.Code.Name() == "unique_violation"
	}
	return false
}

func dbCheckRowsAffected(res sql.Result, expected int64) error {
	actual, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not check rows affected: %v", err)
	}
	if actual != expected {
		return fmt.Errorf("unexpected result: expected %v, got %v", expected, actual)
	}
	return nil
}

type queryExec func(dest interface{}, query string, args ...interface{}) error

func dbFind(dest CacheHinter, query string, args ...interface{}) (Cacheable, error) {
	t := reflect.TypeOf(dest)
	slice := reflect.MakeSlice(reflect.SliceOf(t), 0, 0)
	slicePtr := reflect.New(slice.Type())
	slicePtr.Elem().Set(slice)
	return dbFindExec(services.db.Unsafe().Select, slicePtr.Interface(), dest.cacheHint(), query, args)
}

func dbFindOne(dest CacheHinter, query string, args ...interface{}) (Cacheable, error) {
	t := reflect.TypeOf(dest)
	destPtr := reflect.New(t)
	destPtr.Elem().Set(reflect.ValueOf(dest))

	cacheable, err := dbFindExec(services.db.Unsafe().Get, destPtr.Interface(), dest.cacheHint(), query, args)
	if err == sql.ErrNoRows {
		return Cacheable{}, ErrNotFound
	} else {
		return cacheable, err
	}
}

func dbFindExec(queryExec queryExec, dest interface{}, cacheHint cacheHint, query string, args []interface{}) (Cacheable, error) {

	fail := func(err error, format string, fmtArgs ...interface{}) (Cacheable, error) {
		action := fmt.Sprintf(format, fmtArgs...)
		return Cacheable{}, fmt.Errorf("dbFind:%v failed for [%v, %#v] into %#v: %v", action, query, args, dest, err)
	}

	cacheKey := cacheMakeKeyFromQuery(query, args)
	cacheable, err := cacheGet(cacheKey)
	if err == nil {
		return cacheable, nil
	} else if err != ErrNotFound {
		return fail(err, "cacheGet[%v]", cacheKey)
	}

	err = queryExec(dest, query, args...)
	if err != nil {
		return fail(err, "queryExec")
	}
	bytes, err := json.Marshal(dest)
	if err != nil {
		return fail(err, "jsonify[%v]", dest)
	}

	etag, err := cacheSet(cacheKey, bytes, 1*time.Hour, cacheHint)
	if err != nil {
		return fail(err, "cacheSet[%v]", string(bytes))
	}

	return Cacheable{bytes, etag}, nil
}

func dbDumpQueryResults(rows *sql.Rows) {
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
