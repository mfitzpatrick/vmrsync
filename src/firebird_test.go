package main

import (
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type VesselTest struct {
	ID             int    `firebird:"JOBDUTYVESSELNO" json:"activationsrvsequence"`
	Name           string `firebird:"JOBDUTYVESSELNAME" json:"activationsrvvessel"`
	StartHoursPort int    `firebird:"JOBHOURSSTART" json:"activationsrvenginehours1start"`
	StartHoursStbd int    `json:"activationsrvenginehours2start"`
	EndHoursPort   int    `firebird:"JOBHOURSEND" json:"activationsrvenginehours1end"`
	EndHoursStbd   int    `json:"activationsrvenginehours2end"`
}

type JobTest struct {
	StartTime time.Time `firebird:"JOBTIMEOUT" json:"activationsrvdeparttime"`
	EndTime   time.Time `firebird:"JOBTIMEIN" json:"activationsrvreturntime"`
	VesselTest
}

type testParent struct {
	ID      int       `json:"id"`
	Created time.Time `json:"created_at"`
	Updated time.Time `json:"updated_at"`
	JobTest `firebird:"DUTYJOBS"`
}

func TestForEachCol(t *testing.T) {
	db := &linkActivationDB{
		ID: 42,
		Job: Job{
			StartTime: CustomJSONTime(time.Now()),
			Vessel: Vessel{
				ID:   1,
				Name: "MR1",
			},
		},
	}
	mainObj := reflect.ValueOf(*db)
	err := forEachColumn("parent", mainObj, func(tableName string, col column) error {
		log.Printf("item for %s.%s is %v", tableName, col.name, col.value)
		return nil
	})
	assert.Nil(t, err)
}
