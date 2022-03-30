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

type tripwatchActivation struct {
	ID      int       `json:"id"`
	Created time.Time `json:"created_at"`
	Updated time.Time `json:"updated_at"`
}

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

func listActivations(ctx context.Context) ([]tripwatchActivation, error) {
	ids := []struct {
		ID int `json:"id"`
	}{}
	if resp, err := tripwatchCall(ctx, http.MethodGet, "/activations/recent", ""); err != nil {
		return []tripwatchActivation{}, errors.Wrapf(err, "list activations call")
	} else if body, err := ioutil.ReadAll(resp.Body); err != nil {
		return []tripwatchActivation{}, errors.Wrapf(err, "list activations body read")
	} else if resp.StatusCode == http.StatusTooManyRequests {
		return []tripwatchActivation{},
			errors.Errorf("list activations too many requests")
	} else if resp.StatusCode != http.StatusOK {
		return []tripwatchActivation{},
			errors.Errorf("list activations invalid status code %d", resp.StatusCode)
	} else if err := json.Unmarshal([]byte(body), &ids); err != nil {
		return []tripwatchActivation{}, errors.Wrapf(err, "list activations body parse")
	} else {
		actList := make([]tripwatchActivation, len(ids))
		for _, v := range ids {
			if a, err := getOneActivation(ctx, v.ID); err != nil {
				return []tripwatchActivation{},
					errors.Wrapf(err, "list activations get activation %d", v.ID)
			} else {
				actList = append(actList, a)
			}
		}
		return actList, nil
	}
}

func getOneActivation(ctx context.Context, id int) (tripwatchActivation, error) {
	activation := tripwatchActivation{}
	if resp, err := tripwatchCall(ctx, http.MethodGet, fmt.Sprintf("/activation/%d", id), ""); err != nil {
		return tripwatchActivation{}, errors.Wrapf(err, "get one activation call")
	} else if body, err := ioutil.ReadAll(resp.Body); err != nil {
		return tripwatchActivation{}, errors.Wrapf(err, "get one activation body read")
	} else if err := json.Unmarshal([]byte(body), &activation); err != nil {
		return tripwatchActivation{}, errors.Wrapf(err, "get one activation body parse")
	} else {
		return activation, nil
	}
}
