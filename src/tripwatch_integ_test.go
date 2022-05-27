// +build integration

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
	list, err := listActivations(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, 2, len(list))
	assert.Equal(t, 86239, list[0].ID)
	assert.Equal(t, "MARINERESCUE1", list[0].Job.VMRVessel.Name)
	assert.Equal(t, 86297, list[1].ID)
	assert.Equal(t, "MARINERESCUE2", list[1].Job.VMRVessel.Name)
	assert.Equal(t, getTime(t, "2022-05-27T00:58:00Z"), time.Time(list[1].Job.StartTime))
}
