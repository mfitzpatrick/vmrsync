package main

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

/*
 * Design Philosophy:
 * The purpose of this file is to link the VMR Firebird database with the TripWatch JSON
 * API. This is done using struct tags. When data is received from TripWatch, it is sorted
 * into fields in this structure using the appropriately-named struct tags. Then the same
 * data is sent to the corresponding firebird DB table also by using the firebird struct
 * tags.
 */

type Vessel struct {
	ID             int       `firebird:"JOBDUTYVESSELNO" json:"activationsrvsequence"`
	Name           string    `firebird:"JOBDUTYVESSELNAME,match" json:"activationsrvvessel"`
	StartHoursPort IntString `firebird:"JOBHOURSSTART" json:"activationsrvenginehours1start"`
	StartHoursStbd IntString `json:"activationsrvenginehours2start"`
	EndHoursPort   IntString `firebird:"JOBHOURSEND" json:"activationsrvenginehours1end"`
	EndHoursStbd   IntString `json:"activationsrvenginehours2end"`
}

type Job struct {
	StartTime CustomJSONTime `firebird:"JOBTIMEOUT,match" json:"activationsrvdeparttime"`
	EndTime   CustomJSONTime `firebird:"JOBTIMEIN" json:"activationsrvreturntime"`
	SeaState  string         `firebird:"JOBSEAS" json:"activationsobservedseastate"`
	Vessel
}

type linkActivationDB struct {
	ID      int            `json:"id"`
	Created CustomJSONTime `json:"created_at"`
	Updated CustomJSONTime `json:"updated_at"`
	Job     `firebird:"DUTYJOBS"`
}

type CustomJSONTime time.Time

// Create a custom unmarshaler for timestamps because TripWatch provides timestamps in multiple different
// formats which are not RFC3339-compatible (which is required by the default unmarshaler).
func (t *CustomJSONTime) UnmarshalJSON(bytes []byte) error {
	var outtime time.Time
	if err := json.Unmarshal(bytes, &outtime); err == nil {
		// default parser worked. Assign the time and get out of here
		*t = CustomJSONTime(outtime)
		return nil
	} else if tm, err := time.Parse("2006-01-02 15:04:05", strings.Trim(string(bytes), "\"")); err == nil {
		// this parser worked. Assign the time and get out of here
		*t = CustomJSONTime(tm)
		return nil
	} else {
		return errors.Wrapf(err, "custom unmarshaler failed with time string '%s'", string(bytes))
	}
}

type IntString float32 //TripWatch floating-point number contained in a string

func (i *IntString) UnmarshalJSON(bytes []byte) error {
	rawString := strings.Trim(string(bytes), "\"")
	if strings.ToLower(rawString) == "null" {
		*i = IntString(float32(0.0))
		return nil
	} else if val, err := strconv.ParseFloat(rawString, 32); err != nil {
		return errors.Wrapf(err, "unmarshal intstring")
	} else {
		*i = IntString(float32(val))
		return nil
	}
}
