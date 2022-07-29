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

func TestDBInsertReturningClause(t *testing.T) {
	// Sadly, not supported by Firebird 2.5
	_, err := realDB.QueryContext(context.Background(),
		"UPDATE JOBLOA='8.5m' WHERE JOBDUTYSEQUENCE=1 RETURNING JOBDUTYSEQUENCE")
	assert.NotNil(t, err)
}

func TestGetLatestDutyLogEntry(t *testing.T) {
	table, err := getLatestDutyLogEntry(context.Background(), realDB)
	assert.Nil(t, err)
	assert.Equal(t, 2, table.DutyLog.ID)
	assert.Equal(t, "WHITE", strings.TrimSpace(table.DutyLog.CrewName))
}

func TestFindMemberForEmail(t *testing.T) {
	mbr, err := findMemberForEmail(context.Background(), realDB, "bugs.bunny@mrq.org.au")
	assert.Nil(t, err)
	assert.Equal(t, 3, mbr.ID)
}

func TestFindRankingForMember(t *testing.T) {
	rank, err := findRankingForMember(context.Background(), realDB, 2)
	assert.Nil(t, err)
	assert.Equal(t, 12, rank)
}

func TestPullMemberRecordsByEmail(t *testing.T) {
	member, err := pullMemberRecordsByEmail(context.Background(), realDB, "marvin.the.martian@mrq.org.au")
	assert.Nil(t, err)
	assert.Equal(t, 12, member.CrewOnDuty.RankID, "Record found %+v", member)

	member, err = pullMemberRecordsByEmail(context.Background(), realDB, "bugs.bunny@mrq.org.au")
	assert.Nil(t, err)
	assert.Equal(t, 3, member.CrewOnDuty.RankID, "Record found %+v", member)
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
				// CrewList: StringList{
				// "bugs.bunny@mrq.org.au",
				// },
			},
		},
	}
	err := sendToDB(context.Background(), realDB, dbObj)
	assert.Nil(t, err)

	// Check that data in DB was updated correctly
	var jobID int
	rows, err := realDB.QueryContext(context.Background(),
		"SELECT JOBJOBSEQUENCE,JOBSEAS FROM DUTYJOBS"+
			" WHERE JOBTIMEOUT='2022-01-01 06:00:35' AND JOBDUTYVESSELNAME='MR2'")
	if assert.Nil(t, err) {
		defer rows.Close()
		assert.True(t, rows.Next())
		var seastate string
		err = rows.Scan(&jobID, &seastate)
		assert.Nil(t, err)
		assert.Equal(t, "calm", strings.TrimSpace(seastate))
		assert.Equal(t, 1, jobID)
		assert.False(t, rows.Next())
	}

	// Update the crew list
	dbObj = &linkActivationDB{
		ID: 88,
		Job: Job{
			StartTime: CustomJSONTime(getTimeUTC(t, "2022-01-01T13:10:00Z")),
			VMRVessel: VMRVessel{
				ID:   4,
				Name: "MR5",
				CrewList: StringList{
					"bugs.bunny@mrq.org.au",
				},
			},
		},
	}
	jobID = 3
	err = sendToDB(context.Background(), realDB, dbObj)
	assert.Nil(t, err)
	rows, err = realDB.QueryContext(context.Background(),
		"SELECT CREWMEMBER,CREWRANKING FROM DUTYJOBSCREW"+
			" WHERE CREWJOBSEQUENCE=?", jobID)
	if assert.Nil(t, err) {
		defer rows.Close()
		assert.True(t, rows.Next())
		var memberID, rank int
		err = rows.Scan(&memberID, &rank)
		assert.Nil(t, err)
		assert.Equal(t, 3, memberID)
		assert.Equal(t, 3, rank)
		assert.False(t, rows.Next())
	}
}

func TestSendToDB_NewRecord(t *testing.T) {
	const MAX_PRELOADED_SEQUENCE = 3
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
		"SELECT JOBDUTYSEQUENCE,JOBJOBSEQUENCE,JOBSEAS FROM DUTYJOBS"+
			" WHERE JOBTIMEOUT='2022-02-07 13:50:12' AND JOBDUTYVESSELNAME='MR4'")
	if assert.Nil(t, err) {
		defer rows.Close()
		assert.True(t, rows.Next())
		var dutyseq, seq int
		var seastate string
		err = rows.Scan(&dutyseq, &seq, &seastate)
		assert.Nil(t, err)
		assert.Equal(t, "moderate", strings.TrimSpace(seastate))
		assert.Equal(t, 2, dutyseq)
		assert.Equal(t, MAX_PRELOADED_SEQUENCE+1, seq)
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
				CrewList: StringList{
					"bugs.bunny@mrq.org.au",
				},
			},
		},
	}
	err = sendToDB(context.Background(), realDB, dbObj)
	assert.Nil(t, err)

	// Check that data in DB was updated correctly
	rows, err = realDB.QueryContext(context.Background(),
		"SELECT JOBDUTYSEQUENCE,JOBJOBSEQUENCE,JOBSEAS FROM DUTYJOBS"+
			" WHERE JOBTIMEOUT='2022-02-12 16:01:56' AND JOBDUTYVESSELNAME='MARINERESCUE2'")
	if assert.Nil(t, err) {
		defer rows.Close()
		assert.True(t, rows.Next())
		var dutyseq, seq int
		var seastate string
		err = rows.Scan(&dutyseq, &seq, &seastate)
		assert.Nil(t, err)
		assert.Equal(t, "rough", strings.TrimSpace(seastate))
		assert.Equal(t, MAX_PRELOADED_SEQUENCE+2, seq)
		assert.Equal(t, 2, dutyseq)
		assert.False(t, rows.Next())
	}
	rows, err = realDB.QueryContext(context.Background(),
		"SELECT CREWMEMBER,CREWRANKING FROM DUTYJOBSCREW WHERE CREWJOBSEQUENCE=?", MAX_PRELOADED_SEQUENCE+2)
	if assert.Nil(t, err) {
		defer rows.Close()
		assert.True(t, rows.Next())
		var memberNo, rankID int
		err = rows.Scan(&memberNo, &rankID)
		assert.Nil(t, err)
		assert.Equal(t, 3, memberNo)
		assert.Equal(t, 3, rankID)
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
		"SELECT JOBJOBSEQUENCE,JOBSEAS FROM DUTYJOBS"+
			" WHERE JOBTIMEOUT='2022-01-16 06:09:32' AND JOBDUTYVESSELNAME='MARINERESCUE2'")
	if assert.Nil(t, err) {
		defer rows.Close()
		assert.True(t, rows.Next())
		var seq int
		var seastate string
		err = rows.Scan(&seq, &seastate)
		assert.Nil(t, err)
		assert.Equal(t, "Calm", strings.TrimSpace(seastate))
		assert.Equal(t, MAX_PRELOADED_SEQUENCE+3, seq)
		assert.False(t, rows.Next())
	}
}
