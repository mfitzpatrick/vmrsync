// +build integration

package main

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	// For integration tests, we need to read the configuration file
	if err := parseConfig(*configFilePath); err != nil {
		log.Fatalf("Config parsing failed: %v", err)
	}
}

func TestParseConfig(t *testing.T) {
	assert.Equal(t, "no-api-key", tripwatchAPIkey)
	assert.Equal(t, time.Duration(60*time.Second), tripwatchPollFrequency)
}
