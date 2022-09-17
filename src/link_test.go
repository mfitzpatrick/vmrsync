package main

import (
	"encoding/json"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Check that we can unmarshal the TripWatch data into nested structs properly
func TestNestedStructJSONUnmarshal(t *testing.T) {
	link := linkActivationDB{}
	err := json.Unmarshal([]byte(`{"id":1,`+
		`"created_at":"2022-03-12T12:30:31.000000Z",`+
		`"updated_at":"2022-03-12T12:50:15.000000Z",`+
		`"activationsrvdeparttime":"2022-03-12T12:35:00.000000Z",`+
		`"activationsrvvessel":"MARINERESCUE1"}`), &link)
	assert.Nil(t, err)
	assert.Equal(t, 1, link.ID)
	assert.Equal(t, linkActivationDB{
		ID:      1,
		Created: CustomJSONTime(getTime(t, "2022-03-12T12:30:31.000000Z")),
		Updated: CustomJSONTime(getTime(t, "2022-03-12T12:50:15.000000Z")),
		Job: Job{
			VMRVessel: VMRVessel{
				Name: "Marine Rescue 1",
			},
			StartTime: CustomJSONTime(getTime(t, "2022-03-12T12:35:00.000000Z")),
		},
	}, link)
}

func TestCustomTimeZero(t *testing.T) {
	tm := getTimeUTC(t, "2000-11-11T13:14:15Z")
	ctm := CustomJSONTime(tm)
	assert.False(t, tm.IsZero())
	assert.False(t, time.Time(ctm).IsZero())
	assert.False(t, reflect.ValueOf(tm).IsZero())
	assert.False(t, reflect.ValueOf(ctm).IsZero())
}

// Test case to check that I know how strings are indexed.
// They are indexed a byte at a time (so unicode characters will be indexed 1 byte at a time, and
// not as a whole character).
func TestSubstringAssumption(t *testing.T) {
	s := "This is a string"
	assert.Equal(t, "his is a strin", s[1:len(s)-1])
	assert.Equal(t, 'g', rune(s[len(s)-1]))
	s = "\"Hello'"
	assert.Equal(t, '"', rune(s[0]))
	assert.Equal(t, '\'', rune(s[len(s)-1]))
}

// Helper function for TestStrlenField below
func findAllStrings(t *testing.T, obj reflect.Value) {
	for i := 0; i < obj.NumField(); i++ {
		structVal := obj.Field(i)
		structField := obj.Type().Field(i)
		firebirdTag := structField.Tag.Get("firebird")
		lenTag := structField.Tag.Get("len")
		if structVal.Kind() == reflect.String && firebirdTag != "" {
			if assert.NotEqual(t, "", lenTag, "Field %s", firebirdTag) {
				length, err := strconv.ParseInt(lenTag, 10, 32)
				assert.Nil(t, err, "Field %s", firebirdTag)
				assert.Less(t, 0, int(length), "Field %s", firebirdTag)
			}
		}
		if structVal.Kind() == reflect.Struct && structVal.Type() != reflect.TypeOf(CustomJSONTime{}) {
			findAllStrings(t, structVal)
		}
	}
}

func TestStrlenField(t *testing.T) {
	obj := reflect.ValueOf(linkActivationDB{})
	findAllStrings(t, obj)
}

func TestSitrepGet(t *testing.T) {
	s := []Sitrep{{Pos: GPS{-27, 153}, Comment: "Fake"}}
	_, err := getEntryForComment(s, "RV has arrived at target")
	assert.ErrorIs(t, err, sitrepNotFoundError)

	s = []Sitrep{
		{Pos: GPS{-27, 153}, Comment: "Fake"},
		{Pos: GPS{-27, 153}, Comment: "RV has arrived at target -> DMS"},
	}
	sr, err := getEntryForComment(s, "RV has arrived at target")
	assert.Nil(t, err)
	assert.Equal(t, GPS{-27, 153}, sr.Pos)
}
