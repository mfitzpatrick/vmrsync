// +build integration

package main

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDBWorks(t *testing.T) {
	conn, err := openDB()
	if assert.Nil(t, err) {
		defer conn.Close()
	}

	// Check that data in DB can be updated
	result, err := conn.ExecContext(context.Background(),
		"UPDATE MEMBERS SET PHONE_MOBILE=? WHERE MEMBERNOLOCAL=1",
		"+61423456789")
	assert.Nil(t, err)
	rowCount, err := result.RowsAffected()
	assert.Nil(t, err)
	assert.Equal(t, int64(1), rowCount)

	// Check that data in DB was updated correctly
	rows, err := conn.QueryContext(context.Background(),
		"SELECT SURNAME,PHONE_MOBILE FROM MEMBERS WHERE MEMBERNOLOCAL=1")
	if assert.Nil(t, err) {
		defer rows.Close()
		assert.True(t, rows.Next())
		var surname, mobile string
		err = rows.Scan(&surname, &mobile)
		assert.Nil(t, err)
		assert.Equal(t, "Fudd", strings.TrimSpace(surname))
		assert.Equal(t, "+61423456789", strings.TrimSpace(mobile))
		assert.False(t, rows.Next())
	}
}

func TestRunInsertStatements(t *testing.T) {
	dbObj := &linkActivationDB{
		ID: 42,
		Job: Job{
			StartTime: getTimeUTC(t, "2022-01-01T06:00:35Z"),
			SeaState:  "calm",
			Vessel: Vessel{
				ID:   2,
				Name: "MR2",
			},
		},
	}
	stmts, err := buildInsertStatements(dbObj)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(stmts))

	err = runStatements(context.Background(), realDB, stmts)
	assert.Nil(t, err)

	// Check that data in DB was updated correctly
	rows, err := realDB.QueryContext(context.Background(),
		"SELECT JOBSEAS FROM DUTYJOBS WHERE JOBTIMEOUT='2022-01-01 06:00:35' AND JOBDUTYVESSELNAME='MR2'")
	if assert.Nil(t, err) {
		defer rows.Close()
		assert.True(t, rows.Next())
		var seastate string
		err = rows.Scan(&seastate)
		assert.Nil(t, err)
		assert.Equal(t, "calm", strings.TrimSpace(seastate))
		assert.False(t, rows.Next())
	}
}
