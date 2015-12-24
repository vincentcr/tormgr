package main

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log"
	"strings"

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
	*id = normalizeID(str)
	return nil
}

func newID() RecordID {
	u4 := uuid.NewV4()
	u4str := normalizeID(u4.String())
	return u4str
}

func isUniqueError(err error) bool {
	if err, ok := err.(*pq.Error); ok {
		return err.Code.Name() == "unique_violation"
	}
	return false
}

func sameID(id1, id2 string) bool {
	return normalizeID(id1) == normalizeID(id2)
}

func normalizeID(id string) RecordID {
	return RecordID(strings.ToLower(strings.Replace(id, "-", "", -1)))
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
