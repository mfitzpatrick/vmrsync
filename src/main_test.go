package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Helper function to parse an RFC3339-formatted time string and automatically handle the error.
func getTime(t *testing.T, ts string) time.Time {
	tm, err := time.Parse(time.RFC3339, ts)
	assert.Nil(t, err)
	return tm.UTC()
}

// Same as getTime() but returning an empty timezone (instead of nil timezone)
func getTimeUTC(t *testing.T, ts string) time.Time {
	return getTime(t, ts).In(time.FixedZone("", 0))
}
