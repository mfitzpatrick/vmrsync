// +build integration

package main

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var realDB *sql.DB

func init() {
	if db, err := openDB(); err != nil {
		log.Fatalf("DB Open failed: %v", err)
	} else if err := db.Ping(); err != nil {
		log.Fatalf("No connection to DB: %v", err)
	} else {
		realDB = db
	}
}

func TestIntegDBQuery(t *testing.T) {
	dbc, err := realDB.Conn(context.Background())
	assert.Nil(t, err)
	defer dbc.Close()

	var count int
	err = dbc.QueryRowContext(context.Background(),
		"SELECT Count(*) FROM rdb$relations").Scan(&count)
	assert.Nil(t, err)
	if assert.Less(t, 0, count) {
		log.Printf("Relations Count: %d", count)
	}

	var name sql.NullString
	err = dbc.QueryRowContext(context.Background(),
		"SELECT CREWNAME FROM CREWS WHERE CREWNAME='WHITE'").
		Scan(&name)
	assert.Nil(t, err)
	assert.True(t, name.Valid)
	assert.Equal(t, "WHITE", strings.TrimSpace(name.String))
}

func TestRun(t *testing.T) {
	errlist := run(realDB)
	assert.Equal(t, 0, len(errlist), "Errors in list: %+v", errlist)
}
