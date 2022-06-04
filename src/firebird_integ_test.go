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

func TestSendToDB_ExistingRecord(t *testing.T) {
	dbObj := &linkActivationDB{
		ID: 42,
		Job: Job{
			StartTime: CustomJSONTime(getTimeUTC(t, "2022-01-01T06:00:35Z")),
			SeaState:  "calm",
			VMRVessel: VMRVessel{
				ID:   2,
				Name: "MR2",
			},
		},
	}
	err := sendToDB(context.Background(), realDB, dbObj)
	assert.Nil(t, err)

	// Check that data in DB was updated correctly
	rows, err := realDB.QueryContext(context.Background(),
		"SELECT JOBDUTYSEQUENCE,JOBSEAS FROM DUTYJOBS"+
			" WHERE JOBTIMEOUT='2022-01-01 06:00:35' AND JOBDUTYVESSELNAME='MR2'")
	if assert.Nil(t, err) {
		defer rows.Close()
		assert.True(t, rows.Next())
		var seq int
		var seastate string
		err = rows.Scan(&seq, &seastate)
		assert.Nil(t, err)
		assert.Equal(t, "calm", strings.TrimSpace(seastate))
		assert.Equal(t, 1, seq)
		assert.False(t, rows.Next())
	}
}

func TestSendToDB_NewRecord(t *testing.T) {
	dbObj := &linkActivationDB{
		ID: 482,
		Job: Job{
			StartTime: CustomJSONTime(getTimeUTC(t, "2022-02-07T13:50:12Z")),
			SeaState:  "moderate",
			VMRVessel: VMRVessel{
				ID:   3,
				Name: "MR4",
			},
		},
	}
	err := sendToDB(context.Background(), realDB, dbObj)
	assert.Nil(t, err)

	// Check that data in DB was updated correctly
	rows, err := realDB.QueryContext(context.Background(),
		"SELECT JOBDUTYSEQUENCE,JOBSEAS FROM DUTYJOBS"+
			" WHERE JOBTIMEOUT='2022-02-07 13:50:12' AND JOBDUTYVESSELNAME='MR4'")
	if assert.Nil(t, err) {
		defer rows.Close()
		assert.True(t, rows.Next())
		var seq int
		var seastate string
		err = rows.Scan(&seq, &seastate)
		assert.Nil(t, err)
		assert.Equal(t, "moderate", strings.TrimSpace(seastate))
		assert.Equal(t, 3, seq)
		assert.False(t, rows.Next())
	}

	// Check it again with a different object
	dbObj = &linkActivationDB{
		ID: 882,
		Job: Job{
			StartTime: CustomJSONTime(getTimeUTC(t, "2022-02-12T16:01:56Z")),
			SeaState:  "rough",
			VMRVessel: VMRVessel{
				ID:   2,
				Name: "MARINERESCUE2",
			},
		},
	}
	err = sendToDB(context.Background(), realDB, dbObj)
	assert.Nil(t, err)

	// Check that data in DB was updated correctly
	rows, err = realDB.QueryContext(context.Background(),
		"SELECT JOBDUTYSEQUENCE,JOBSEAS FROM DUTYJOBS"+
			" WHERE JOBTIMEOUT='2022-02-12 16:01:56' AND JOBDUTYVESSELNAME='MARINERESCUE2'")
	if assert.Nil(t, err) {
		defer rows.Close()
		assert.True(t, rows.Next())
		var seq int
		var seastate string
		err = rows.Scan(&seq, &seastate)
		assert.Nil(t, err)
		assert.Equal(t, "rough", strings.TrimSpace(seastate))
		assert.Equal(t, 4, seq)
		assert.False(t, rows.Next())
	}

	// And again with everything that's failing in another test
	dbObj = &linkActivationDB{
		ID: 22,
		Job: Job{
			StartTime:   CustomJSONTime(getTimeUTC(t, "2022-01-16T06:09:32Z")),
			EndTime:     CustomJSONTime(getTimeUTC(t, "2022-01-16T08:00:00Z")),
			Type:        "Assist",
			Action:      "Tow, refloat, medical assist",
			Comments:    "This is the comments field.",
			Donation:    IntString(200),
			WaterLimits: "E",
			SeaState:    "Calm",
			AssistedVessel: AssistedVessel{
				Rego:       "AB123Q",
				Name:       "Dummy II",
				Length:     LengthEnum("0-8m"),
				Type:       "Party Pontoon",
				Propulsion: "Outboard",
				NumAdults:  1,
				NumKids:    3,
			},
			Emergency: Emergency{
				PoliceNum: "987654321",
				Notified:  "t",
			},
			GPS: GPS{
				TWLatLong: "-27.5,153.7",
			},
			VMRVessel: VMRVessel{
				ID:             2,
				Name:           "MARINERESCUE2",
				StartHoursPort: IntString(56),
				EndHoursPort:   IntString(58),
			},
			Weather: Weather{
				WindSpeed: WindSpeedEnum("10 - 20kt"),
				WindDir:   WindDirEnum("SE"),
				RainState: "Clear",
			},
		},
	}
	err = sendToDB(context.Background(), realDB, dbObj)
	assert.Nil(t, err)

	// Check that data in DB was updated correctly
	rows, err = realDB.QueryContext(context.Background(),
		"SELECT JOBDUTYSEQUENCE,JOBSEAS FROM DUTYJOBS"+
			" WHERE JOBTIMEOUT='2022-01-16 06:09:32' AND JOBDUTYVESSELNAME='MARINERESCUE2'")
	if assert.Nil(t, err) {
		defer rows.Close()
		assert.True(t, rows.Next())
		var seq int
		var seastate string
		err = rows.Scan(&seq, &seastate)
		assert.Nil(t, err)
		assert.Equal(t, "Calm", strings.TrimSpace(seastate))
		assert.Equal(t, 5, seq)
		assert.False(t, rows.Next())
	}
}
