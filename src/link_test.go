package main

import (
	"encoding/json"
	"reflect"
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
				Name: "MARINERESCUE1",
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
