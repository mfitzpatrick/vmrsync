package main

import (
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

// Example nested struct tag parsing
func TestStructTagParsing(t *testing.T) {
	parent := testParent{}
	mainObj := reflect.TypeOf(parent)
	assert.Equal(t, reflect.Struct, mainObj.Kind())
	tableMap, err := getFirebirdStructTags("parent", mainObj)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(tableMap))
	expect := make(map[string]map[string]reflect.Type)
	expect["DUTYJOBS"] = map[string]reflect.Type{
		"JOBTIMEOUT":        reflect.ValueOf(time.Time{}).Type(),
		"JOBTIMEIN":         reflect.ValueOf(time.Time{}).Type(),
		"JOBDUTYVESSELNO":   reflect.ValueOf(0).Type(),
		"JOBDUTYVESSELNAME": reflect.ValueOf("").Type(),
		"JOBHOURSSTART":     reflect.ValueOf(0).Type(),
		"JOBHOURSEND":       reflect.ValueOf(0).Type(),
	}
	assert.Equal(t, expect, tableMap)
}

func TestFirebirdGetRequests(t *testing.T) {
	db := linkActivationDB{}
	err := firebirdGet(&db)
	assert.Nil(t, err)
}
