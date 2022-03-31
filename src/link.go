package main

import (
	"time"
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
	ID             int    `firebird:"JOBDUTYVESSELNO" json:"activationsrvsequence"`
	Name           string `firebird:"JOBDUTYVESSELNAME" json:"activationsrvvessel"`
	StartHoursPort int    `firebird:"JOBHOURSSTART" json:"activationsrvenginehours1start"`
	StartHoursStbd int    `json:"activationsrvenginehours2start"`
	EndHoursPort   int    `firebird:"JOBHOURSEND" json:"activationsrvenginehours1end"`
	EndHoursStbd   int    `json:"activationsrvenginehours2end"`
}

type Job struct {
	StartTime time.Time `firebird:"JOBTIMEOUT" json:"activationsrvdeparttime"`
	EndTime   time.Time `firebird:"JOBTIMEIN" json:"activationsrvreturntime"`
	Vessel
}

type linkActivationDB struct {
	ID      int       `json:"id"`
	Created time.Time `json:"created_at"`
	Updated time.Time `json:"updated_at"`
	Job     `firebird:"DUTYJOBS"`
}
