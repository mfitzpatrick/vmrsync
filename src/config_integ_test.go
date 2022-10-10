//go:build integration

package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {
	assert.Equal(t, "no-api-key", tripwatchAPIkey)
	assert.Equal(t, time.Duration(60*time.Second), tripwatchPollFrequency)
}
