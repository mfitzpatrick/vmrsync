package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

var tripwatchAPIkey string
var tripwatchURL string

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
		for _, v := range ids {
			if a, err := getOneActivation(ctx, v.ID); err != nil {
				return []linkActivationDB{},
					errors.Wrapf(err, "list activations get activation %d", v.ID)
			} else {
				actList = append(actList, a)
			}
		}
		return actList, nil
	}
}

func getOneActivation(ctx context.Context, id int) (linkActivationDB, error) {
	activation := linkActivationDB{}
	if resp, err := tripwatchCall(ctx, http.MethodGet, fmt.Sprintf("/activation/%d", id), ""); err != nil {
		return linkActivationDB{}, errors.Wrapf(err, "get one activation call")
	} else if body, err := ioutil.ReadAll(resp.Body); err != nil {
		return linkActivationDB{}, errors.Wrapf(err, "get one activation body read")
	} else if err := json.Unmarshal([]byte(body), &activation); err != nil {
		return linkActivationDB{}, errors.Wrapf(err, "get one activation body parse")
	} else {
		return activation, nil
	}
}
