//go:build integration

package main

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	// Set a flag which will cause TestMain to connect to the DB.
	shouldOpenDB = true
}

func TestIntegDBQuery(t *testing.T) {
	dbc, err := realDB.Conn(context.Background())
	assert.Nil(t, err)
	defer dbc.Close()

	var count int
	err = dbc.QueryRowContext(context.Background(),
		"SELECT Count(*) FROM rdb$relations").Scan(&count)
	assert.Nil(t, err)
	assert.Equal(t, 101, count)

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
