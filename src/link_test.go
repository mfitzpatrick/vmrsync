package main

import (
	"encoding/json"
	"testing"

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
	assert.Equal(t, "MARINERESCUE1", link.Job.Vessel.Name)
}
