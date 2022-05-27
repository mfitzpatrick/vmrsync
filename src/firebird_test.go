package main

import (
	"log"
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
		log.Printf("item for %s.%s is %v", tableName, col.name, col.value)
		return nil
	})
	assert.Nil(t, err)
}

func TestAggregateFields(t *testing.T) {
	// Check emergency flag
	data := &linkActivationDB{
		Job: Job{
			Emergency: Emergency{
				Notified: true,
			},
		},
	}
	err := aggregateFields(data)
	assert.Nil(t, err)
	assert.True(t, data.Job.Emergency.Emergency)
	assert.False(t, data.Job.Commercial)

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
	assert.False(t, data.Job.Emergency.Emergency)
	assert.True(t, data.Job.Commercial)

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
	assert.False(t, data.Job.Emergency.Emergency)
	assert.Equal(t, "SE", string(data.Job.Weather.WindDir))
	assert.Equal(t, "10 - 20kt", string(data.Job.Weather.WindSpeed))
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
}
