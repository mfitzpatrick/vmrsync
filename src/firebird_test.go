package main

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type VesselTest struct {
	ID             int    `firebird:"JOBDUTYVESSELNO" json:"activationsrvsequence"`
	Name           string `firebird:"JOBDUTYVESSELNAME" json:"activationsrvvessel"`
	StartHoursPort int    `firebird:"JOBHOURSSTART" json:"activationsrvenginehours1start"`
	StartHoursStbd int    `json:"activationsrvenginehours2start"`
	EndHoursPort   int    `firebird:"JOBHOURSEND" json:"activationsrvenginehours1end"`
	EndHoursStbd   int    `json:"activationsrvenginehours2end"`
}

type JobTest struct {
	StartTime time.Time `firebird:"JOBTIMEOUT" json:"activationsrvdeparttime"`
	EndTime   time.Time `firebird:"JOBTIMEIN" json:"activationsrvreturntime"`
	VesselTest
}

type testParent struct {
	ID      int       `json:"id"`
	Created time.Time `json:"created_at"`
	Updated time.Time `json:"updated_at"`
	JobTest `firebird:"DUTYJOBS"`
}

func TestForEachCol(t *testing.T) {
	db := &linkActivationDB{
		ID: 42,
		Job: Job{
			StartTime: CustomJSONTime(time.Now()),
			VMRVessel: VMRVessel{
				ID:   1,
				Name: "MR1",
			},
		},
	}
	mainObj := reflect.ValueOf(*db)
	err := forEachColumn("parent", mainObj, func(tableName string, col column) error {
		if col.name == "JOBDUTYVESSELNAME" {
			assert.Equal(t, db.Job.VMRVessel.Name, col.value)
		} else if col.name == "JOBDUTYVESSELNO" {
			assert.Equal(t, db.Job.VMRVessel.ID, col.value)
		} else if col.name == "JOBTIMEOUT" {
			assert.Equal(t, db.Job.StartTime, col.value)
		}
		return nil
	})
	assert.Nil(t, err)

	// Test that a nested struct can do for each col as well
	colNames := []string{}
	coj := crewOnJob{}
	o := reflect.ValueOf(coj)
	err = forEachColumn("parent", o, func(tableName string, col column) error {
		if tableName == "DUTYJOBSCREW" {
			colNames = append(colNames, col.name)
		}
		return nil
	})
	assert.Nil(t, err)
	assert.Equal(t, []string{"CREWDUTYSEQUENCE", "CREWJOBSEQUENCE", "CREWMEMBER",
		"CREWRANKING", "SKIPPER", "CREWONJOB"},
		colNames)
}

func TestAggregateFields(t *testing.T) {
	// Check emergency flag
	data := &linkActivationDB{
		Job: Job{
			Emergency: Emergency{
				Notified: "Y",
			},
		},
	}
	err := aggregateFields(data)
	assert.Nil(t, err)
	assert.Equal(t, "Y", string(data.Job.Emergency.Emergency))
	assert.Equal(t, "N", string(data.Job.Commercial))

	// Check commercial flag
	data = &linkActivationDB{
		Job: Job{
			AssistedVessel: AssistedVessel{
				Rego: "ABC123QC",
			},
		},
	}
	err = aggregateFields(data)
	assert.Nil(t, err)
	assert.Equal(t, "", string(data.Job.Emergency.Emergency))
	assert.Equal(t, "Y", string(data.Job.Commercial))

	// Check forecast parsing
	data = &linkActivationDB{
		Job: Job{
			Weather: Weather{
				Forecast: "[Byron Coast: Point Danger to Wooli, Winds:" +
					"  Southeasterly 10 to 15 knots, reaching up to 20 knots" +
					" north of Yamba in the evening.\r\n" +
					" Seas:  Around 1 metre, increasing to 1 to 1.5 metres offshore north of Cape Byron.\r\n" +
					" Swell1:  Southerly around 1 metre inshore, increasing to 1.5 metres offshore.\r\n" +
					" Swell2:  Easterly 1.5 metres.\r\n Weather:  Mostly clear.\r\n] " +
					"[Moreton Bay, Winds:  South to southeasterly 15 to 20 knots.\r\n" +
					" Seas:  Around 1 metre, increasing to 1 to 1.5 metres in the northern bay.\r\n" +
					" Weather:  Partly cloudy. 40% chance of showers over the" +
					" eastern bay during the evening.\r\n] " +
					"[Gold Coast Waters: Cape Moreton to Point Danger," +
					" Winds:  South to southeasterly 15 to 20 knots.\r\n" +
					" Seas:  1.5 metres.\r\n" +
					" Swell1:  Easterly around 1 metre inshore, increasing to 1.5 metres offshore.\r\n" +
					" Swell2:  Southerly below 1 metre inshore, increasing to 1 to 1.5 metres offshore.\r\n" +
					" Weather:  Partly cloudy. 40% chance of showers offshore.\r\n]",
			},
		},
	}
	err = aggregateFields(data)
	assert.Nil(t, err)
	assert.Equal(t, "SE", string(data.Job.Weather.WindDir))
	assert.Equal(t, "10 - 20 knots", string(data.Job.Weather.WindSpeed))
	assert.Equal(t, "Clear", data.Job.Weather.RainState)

	// Check GPS parsing
	data = &linkActivationDB{
		Job: Job{
			GPS: GPS{
				TWLatLong: "-27.5,153.7",
			},
		},
	}
	err = aggregateFields(data)
	assert.Nil(t, err)
	assert.Equal(t, -27.5, data.Job.GPS.Lat)
	assert.Equal(t, 153.7, data.Job.GPS.Long)
	assert.Equal(t, 27, data.Job.GPS.LatD)
	assert.Equal(t, 30, data.Job.GPS.LatM)
	assert.InDelta(t, 0.0, data.Job.GPS.LatS, 0.1)
	assert.Equal(t, 153, data.Job.GPS.LongD)
	assert.Equal(t, 41, data.Job.GPS.LongM)
	assert.InDelta(t, 59.9, data.Job.GPS.LongS, 0.1)
	data = &linkActivationDB{
		Job: Job{
			GPS: GPS{
				TWLatLong: "-27.5 153.7",
			},
		},
	}
	err = aggregateFields(data)
	assert.Nil(t, err)
	assert.Equal(t, -27.5, data.Job.GPS.Lat)
	assert.Equal(t, 153.7, data.Job.GPS.Long)
	assert.Equal(t, 27, data.Job.GPS.LatD)
	assert.Equal(t, 30, data.Job.GPS.LatM)
	assert.InDelta(t, 0.0, data.Job.GPS.LatS, 0.1)
	assert.Equal(t, 153, data.Job.GPS.LongD)
	assert.Equal(t, 41, data.Job.GPS.LongM)
	assert.InDelta(t, 59.9, data.Job.GPS.LongS, 0.1)

	// Check Vessel Propulsion
	data = &linkActivationDB{
		Job: Job{
			AssistedVessel: AssistedVessel{
				Type:       "Speed/Motor Boat",
				Propulsion: "Single Outboard",
				EngineQTY:  1,
			},
		},
	}
	err = aggregateFields(data)
	assert.Nil(t, err)
	assert.Equal(t, PropulsionEnum("Single Outboard"), data.Job.AssistedVessel.Propulsion)
	data = &linkActivationDB{
		Job: Job{
			AssistedVessel: AssistedVessel{
				Type:       "Speed/Motor Boat",
				Propulsion: "Single Outboard",
				EngineQTY:  2,
			},
		},
	}
	err = aggregateFields(data)
	assert.Nil(t, err)
	assert.Equal(t, PropulsionEnum("Twin Outboard"), data.Job.AssistedVessel.Propulsion)
	data = &linkActivationDB{
		Job: Job{
			AssistedVessel: AssistedVessel{
				Type:       "Speed/Motor Boat",
				Propulsion: "Single Inboard",
				EngineQTY:  2,
			},
		},
	}
	err = aggregateFields(data)
	assert.Nil(t, err)
	assert.Equal(t, PropulsionEnum("Twin Inboard"), data.Job.AssistedVessel.Propulsion)
	data = &linkActivationDB{
		Job: Job{
			AssistedVessel: AssistedVessel{
				Type:       "Speed/Motor Boat",
				Propulsion: "Sail",
				EngineQTY:  6,
			},
		},
	}
	err = aggregateFields(data)
	assert.Nil(t, err)
	assert.Equal(t, PropulsionEnum("Sail"), data.Job.AssistedVessel.Propulsion)
	data = &linkActivationDB{
		Job: Job{
			AssistedVessel: AssistedVessel{
				Type:       "Kayak",
				Propulsion: "Oars",
				EngineQTY:  3,
			},
		},
	}
	err = aggregateFields(data)
	assert.Nil(t, err)
	assert.Equal(t, PropulsionEnum("Oars"), data.Job.AssistedVessel.Propulsion)
}
