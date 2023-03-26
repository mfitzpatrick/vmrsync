//go:build integration

package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIntegManualTripwatchGet(t *testing.T) {
	req, err := http.NewRequestWithContext(context.Background(),
		http.MethodGet, tripwatchURL+"/activations/recent", strings.NewReader(""))
	assert.Nil(t, err)
	req.Header.Add("Authorization", "Bearer "+tripwatchAPIkey)
	c := http.Client{}
	resp, err := c.Do(req)
	if assert.Nil(t, err) {
		body, err := ioutil.ReadAll(resp.Body)
		assert.Nil(t, err)
		assert.NotEqual(t, "", string(body))
	}
}

func TestIntegTripwatchCallHelper(t *testing.T) {
	resp, err := tripwatchCall(context.Background(), http.MethodGet, "/activations/recent", "")
	if assert.Nil(t, err) {
		body, err := ioutil.ReadAll(resp.Body)
		assert.Nil(t, err)
		assert.NotEqual(t, "", string(body))
	}
}

func TestIntegTripwatchListActivations(t *testing.T) {
	list, err := listActivations(context.Background(), now())
	assert.Nil(t, err)
	assert.Equal(t, 0, len(list))

	lastUpdatedTS = getTimeUTC(t, "2022-05-27T01:04:00Z")
	setNow(getTimeUTC(t, "2022-05-27T01:08:00Z"))
	list, err = listActivations(context.Background(), lastUpdatedTS)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(list))
	assert.Equal(t, 86297, list[0].ID)
	assert.Equal(t, "Marine Rescue 2", string(list[0].Job.VMRVessel.Name))
	assert.Equal(t, getTime(t, "2022-05-27T00:58:00Z"), time.Time(list[0].Job.StartTime))
	assert.Equal(t, getTimeUTC(t, "2022-05-27T01:08:00Z"), now())

	lastUpdatedTS = getTimeUTC(t, "2022-03-21T09:37:01Z")
	setNow(getTimeUTC(t, "2022-03-21T09:40:01Z"))
	list, err = listActivations(context.Background(), lastUpdatedTS)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(list))
	assert.Equal(t, 86239, list[0].ID)
	assert.Equal(t, "Marine Rescue 1", string(list[0].Job.VMRVessel.Name))
}

func TestIntegTripwatchGetOneActivation(t *testing.T) {
	a, err := getOneActivation(context.Background(), 86359)
	assert.Nil(t, err)
	assert.Equal(t, "InProgress", a.Job.Status)
	assert.Equal(t, 3, len(a.Sitreps))
	assert.Equal(t, GPS{-27.475458084334857, 153.15326141723338}, a.Sitreps[0].Pos)
	// Test GPS field aggregation
	err = aggregateFields(&a)
	assert.Nil(t, err)
	assert.Equal(t, -27.475458084334857, a.Job.FirebirdGPS.Lat)
	assert.Equal(t, 153.15326141723338, a.Job.FirebirdGPS.Long)
}
