package main

import (
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
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

// Get a UTC timestamp from a time in the UTC+10 timezone
func getTimeFromAEST(t *testing.T, ts string) time.Time {
	tz := time.FixedZone("UTC+10", 10*60*60)
	tm, err := time.ParseInLocation(time.RFC3339, ts, tz)
	assert.Nil(t, err)
	return tm.UTC()
}

func TestGetTimeFromAEST(t *testing.T) {
	assert.Equal(t, getTime(t, "2020-01-01T03:15:00Z"),
		getTimeFromAEST(t, "2020-01-01T13:15:00+10:00"))
}

func TestRunError(t *testing.T) {
	record := &linkActivationDB{
		ID: 42,
		Job: Job{
			VMRVessel: VMRVessel{
				ID:   1,
				Name: "MR1",
			},
		},
	}
	chain := errors.Wrapf(errors.Wrapf(errors.Wrapf(matchFieldIsZero, "layer 1"), "layer 2"), "layer 3")
	chainedRunerr := runError{
		error:      errors.Wrapf(chain, "runError wrapping"),
		activation: record,
	}
	higherLevelErr := errors.Wrapf(chainedRunerr, "higher level")
	assert.True(t, errors.Is(chainedRunerr, matchFieldIsZero))
	assert.True(t, errors.Is(higherLevelErr, matchFieldIsZero))
	assert.True(t, strings.HasPrefix(chainedRunerr.Error(), "activation 42 on MR1"))
	assert.True(t, strings.Contains(higherLevelErr.Error(), "activation 42 on MR1"))
	assert.True(t, strings.Contains(higherLevelErr.Error(), "runError wrapping"))
	assert.True(t, strings.Contains(higherLevelErr.Error(), "layer 2"))
}
