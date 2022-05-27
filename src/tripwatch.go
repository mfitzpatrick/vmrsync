package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var tripwatchAPIkey string
var tripwatchURL string
var tripwatchPollFrequency time.Duration

var (
	twNotFound = errors.Errorf("TripWatch item not found")
)

func tripwatchCall(ctx context.Context, method, url, body string) (*http.Response, error) {
	if req, err := http.NewRequestWithContext(ctx,
		method, tripwatchURL+url, strings.NewReader(body),
	); err != nil {
		return &http.Response{}, errors.Wrapf(err, "tripwatch call new request")
	} else {
		req.Header.Add("Authorization", "Bearer "+tripwatchAPIkey)
		c := http.Client{}
		if resp, err := c.Do(req); err != nil {
			return &http.Response{}, errors.Wrapf(err, "tripwatch call execute")
		} else if resp.StatusCode == 404 {
			return &http.Response{}, errors.Wrapf(twNotFound, "tripwatch call")
		} else {
			return resp, nil
		}
	}
}

func listActivations(ctx context.Context) ([]linkActivationDB, error) {
	ids := []struct {
		ID int `json:"id"`
	}{}
	if resp, err := tripwatchCall(ctx, http.MethodGet, "/activations/recent", ""); err != nil {
		return []linkActivationDB{}, errors.Wrapf(err, "list activations call")
	} else if body, err := ioutil.ReadAll(resp.Body); err != nil {
		return []linkActivationDB{}, errors.Wrapf(err, "list activations body read")
	} else if resp.StatusCode == http.StatusTooManyRequests {
		return []linkActivationDB{},
			errors.Errorf("list activations too many requests")
	} else if resp.StatusCode != http.StatusOK {
		return []linkActivationDB{},
			errors.Errorf("list activations invalid status code %d", resp.StatusCode)
	} else if err := json.Unmarshal([]byte(body), &ids); err != nil {
		return []linkActivationDB{}, errors.Wrapf(err, "list activations body parse")
	} else {
		actList := make([]linkActivationDB, len(ids))
		for i, v := range ids {
			if a, err := getOneActivation(ctx, v.ID); err != nil {
				return []linkActivationDB{},
					errors.Wrapf(err, "list activations get activation %d", v.ID)
			} else {
				actList[i] = a
			}
		}
		return actList, nil
	}
}

func getOneActivation(ctx context.Context, id int) (linkActivationDB, error) {
	activation := linkActivationDB{}
	if resp, err := tripwatchCall(ctx, http.MethodGet, fmt.Sprintf("/activations/%d", id), ""); err != nil {
		return linkActivationDB{}, errors.Wrapf(err, "get one activation call for ID %d", id)
	} else if body, err := ioutil.ReadAll(resp.Body); err != nil {
		return linkActivationDB{}, errors.Wrapf(err, "get one activation body read for ID %d", id)
	} else if err := json.Unmarshal([]byte(body), &activation); err != nil {
		return linkActivationDB{}, errors.Wrapf(err, "get one activation body parse for ID %d '%s'", id, body)
	} else {
		return activation, nil
	}
}
