// +build integration

package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegManualTripwatchGet(t *testing.T) {
	req, err := http.NewRequestWithContext(context.Background(),
		http.MethodGet, tripwatchURL+"/activations/recent", strings.NewReader(""))
	assert.Nil(t, err)
	req.Header.Add("Authorization", "Bearer "+tripwatchAPIkey)
	c := http.Client{}
	resp, err := c.Do(req)
	assert.Nil(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.NotEqual(t, "", string(body))
}

func TestIntegTripwatchCallHelper(t *testing.T) {
	resp, err := tripwatchCall(context.Background(), http.MethodGet, "/activations/recent", "")
	assert.Nil(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.NotEqual(t, "", string(body))
}

func TestIntegTripwatchListActivations(t *testing.T) {
	_, err := listActivations(context.Background())
	assert.Nil(t, err)
	// assert.Less(t, 0, len(list))
}
