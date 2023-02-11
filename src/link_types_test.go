package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCustomJSONTime(t *testing.T) {
	var v CustomJSONTime
	err := (&v).UnmarshalJSON([]byte(`"2020-01-01T03:15:00Z"`))
	assert.Nil(t, err)
	assert.Equal(t, getTime(t, "2020-01-01T03:15:00Z"), time.Time(v))

	// Parse non-standard timestamp
	err = (&v).UnmarshalJSON([]byte(`"2020-01-01 03:15:00"`))
	assert.Nil(t, err)
	assert.Equal(t, getTime(t, "2020-01-01T03:15:00Z"), time.Time(v))
}

func TestCustomBoolUnmarshal(t *testing.T) {
	var b CustomBool
	err := (&b).UnmarshalJSON([]byte("false"))
	assert.Nil(t, err)
	assert.Equal(t, "N", string(b))

	err = (&b).UnmarshalJSON([]byte("true"))
	assert.Nil(t, err)
	assert.Equal(t, "Y", string(b))

	err = (&b).UnmarshalJSON([]byte("\"Yes\" "))
	assert.Nil(t, err)
	assert.Equal(t, "Y", string(b))
}

func TestIntStringUnmarshal(t *testing.T) {
	var i IntString
	err := (&i).UnmarshalJSON([]byte("156"))
	assert.Nil(t, err)
	assert.Equal(t, 156, int(i))

	err = (&i).UnmarshalJSON([]byte("3665489.5351867"))
	assert.Nil(t, err)
	assert.Equal(t, float32(3665489.5351867), float32(i))

	err = (&i).UnmarshalJSON([]byte("null"))
	assert.Nil(t, err)
	assert.Equal(t, float32(0.0), float32(i))
}

func TestLengthEnumUnmarshal(t *testing.T) {
	var l LengthEnum
	err := (&l).UnmarshalJSON([]byte("5"))
	assert.Nil(t, err)
	assert.Equal(t, "4.5m - 8m", string(l))

	err = (&l).UnmarshalJSON([]byte("15m"))
	assert.Nil(t, err)
	assert.Equal(t, "15m - 25m", string(l))

	err = (&l).UnmarshalJSON([]byte("15'"))
	assert.Nil(t, err)
	assert.Equal(t, "4.5m - 8m", string(l))

	err = (&l).UnmarshalJSON([]byte("30  '"))
	assert.Nil(t, err)
	assert.Equal(t, "8m - 10m", string(l))

	err = (&l).UnmarshalJSON([]byte(" \"15"))
	assert.Nil(t, err)
	assert.Equal(t, "15m - 25m", string(l))

	err = (&l).UnmarshalJSON([]byte("null"))
	assert.Nil(t, err)
	assert.Equal(t, "", string(l))
}

func TestWindSpeedEnumUnmarshal(t *testing.T) {
	var w WindSpeedEnum
	err := (&w).UnmarshalJSON([]byte("15"))
	assert.Nil(t, err)
	assert.Equal(t, "10 - 20 knots", string(w))
}

func TestWindDirectionEnumUnmarshal(t *testing.T) {
	var w WindDirEnum
	err := (&w).UnmarshalJSON([]byte("South-East"))
	assert.Nil(t, err)
	assert.Equal(t, "SE", string(w))

	err = (&w).UnmarshalJSON([]byte("north"))
	assert.Nil(t, err)
	assert.Equal(t, "N", string(w))

	err = (&w).UnmarshalJSON([]byte("norTH"))
	assert.Nil(t, err)
	assert.Equal(t, "N", string(w))
}

func TestSeaStateEnumUnmarshal(t *testing.T) {
	var s SeaStateEnum
	err := (&s).UnmarshalJSON([]byte("3"))
	assert.Nil(t, err)
	assert.Equal(t, "Calm", string(s))

	err = (&s).UnmarshalJSON([]byte("5"))
	assert.Nil(t, err)
	assert.Equal(t, "Moderate", string(s))

	err = (&s).UnmarshalJSON([]byte("9"))
	assert.Nil(t, err)
	assert.Equal(t, "Rough", string(s))

	err = (&s).UnmarshalJSON([]byte("null"))
	assert.Nil(t, err)
	assert.Equal(t, "", string(s))
}

func TestStringListUnmarshal(t *testing.T) {
	var s StringList
	err := (&s).UnmarshalJSON([]byte(`["s1", "s2"]`))
	assert.Nil(t, err)
	assert.Equal(t, []string{"s1", "s2"}, []string(s))

	err = (&s).UnmarshalJSON([]byte(` "[\"s1\", \"s2\"]" `))
	assert.Nil(t, err)
	assert.Equal(t, []string{"s1", "s2"}, []string(s))
}

func TestJobType(t *testing.T) {
	var j JobType
	err := (&j).UnmarshalJSON([]byte(`"Assist"`))
	assert.Nil(t, err)
	assert.Equal(t, "Breakdown", string(j))

	err = (&j).UnmarshalJSON([]byte(`"SAR"`))
	assert.Nil(t, err)
	assert.Equal(t, "Search", string(j))

	err = (&j).UnmarshalJSON([]byte(`"Training"`))
	assert.Nil(t, err)
	assert.Equal(t, "Training/Patrol", string(j))

	err = (&j).UnmarshalJSON([]byte(`"Other type"`))
	assert.Nil(t, err)
	assert.Equal(t, "Other type", string(j))
}

func TestJobTypeToAction(t *testing.T) {
	assert.Equal(t, "Training", string(JobType("Training/Patrol").ToJobAction()))
	assert.Equal(t, "Medivac", string(JobType("Medical").ToJobAction()))
}

func TestJobAction(t *testing.T) {
	var j JobAction
	err := (&j).UnmarshalJSON([]byte(`"Tow"`))
	assert.Nil(t, err)
	assert.Equal(t, "Tow", string(j))

	err = (&j).UnmarshalJSON([]byte(`"go tOWards the light"`))
	assert.Nil(t, err)
	assert.Equal(t, "Tow", string(j))

	err = (&j).UnmarshalJSON([]byte(`"search for boat with cops"`))
	assert.Nil(t, err)
	assert.Equal(t, "Search & Rescue", string(j))

	err = (&j).UnmarshalJSON([]byte(`"ungrounding of boat"`))
	assert.Nil(t, err)
	assert.Equal(t, "Ungrounded", string(j))

	err = (&j).UnmarshalJSON([]byte(`"Miscellaneous"`))
	assert.Nil(t, err)
	assert.Equal(t, "Other", string(j))

	err = (&j).UnmarshalJSON([]byte(`"medical emergency"`))
	assert.Nil(t, err)
	assert.Equal(t, "Medivac", string(j))

	err = (&j).UnmarshalJSON([]byte(`"Broadwater Training"`))
	assert.Nil(t, err)
	assert.Equal(t, "Training", string(j))
}

func TestWaterLimitsEnum(t *testing.T) {
	var w WaterLimitsEnum
	err := (&w).UnmarshalJSON([]byte(`"E"`))
	assert.Nil(t, err)
	assert.Equal(t, "Smooth", string(w))

	err = (&w).UnmarshalJSON([]byte(`"D"`))
	assert.Nil(t, err)
	assert.Equal(t, "Partially Smooth", string(w))

	err = (&w).UnmarshalJSON([]byte(`"C"`))
	assert.Nil(t, err)
	assert.Equal(t, "Open", string(w))

	err = (&w).UnmarshalJSON([]byte(`"B"`))
	assert.Nil(t, err)
	assert.Equal(t, "Open", string(w))

	err = (&w).UnmarshalJSON([]byte(`"A"`))
	assert.Nil(t, err)
	assert.Equal(t, "Open", string(w))
}

func TestVMRVesselNameEnum(t *testing.T) {
	var n VMRVesselNameEnum
	err := (&n).UnmarshalJSON([]byte(`"MARINERESCUE1"`))
	assert.Nil(t, err)
	assert.Equal(t, "Marine Rescue 1", string(n))

	err = (&n).UnmarshalJSON([]byte(`"MARINERESCUE2"`))
	assert.Nil(t, err)
	assert.Equal(t, "Marine Rescue 2", string(n))

	err = (&n).UnmarshalJSON([]byte(`"MARINERESCUE4"`))
	assert.Nil(t, err)
	assert.Equal(t, "Marine Rescue 4", string(n))

	err = (&n).UnmarshalJSON([]byte(`"MARINERESCUE5"`))
	assert.Nil(t, err)
	assert.Equal(t, "Marine Rescue 5", string(n))
}

func TestBoatTypeEnum(t *testing.T) {
	var b BoatTypeEnum
	err := (&b).UnmarshalJSON([]byte(`"jetski"`))
	assert.Nil(t, err)
	assert.Equal(t, "PWC", string(b))

	err = (&b).UnmarshalJSON([]byte(`"Sailing Catamaran"`))
	assert.Nil(t, err)
	assert.Equal(t, "Sailing", string(b))

	err = (&b).UnmarshalJSON([]byte(`"double-masted YachT"`))
	assert.Nil(t, err)
	assert.Equal(t, "Sailing", string(b))

	err = (&b).UnmarshalJSON([]byte(`"KETCH"`))
	assert.Nil(t, err)
	assert.Equal(t, "Sailing", string(b))

	err = (&b).UnmarshalJSON([]byte(`"Schooner"`))
	assert.Nil(t, err)
	assert.Equal(t, "Sailing", string(b))

	err = (&b).UnmarshalJSON([]byte(`"The most common type"`))
	assert.Nil(t, err)
	assert.Equal(t, "Speed/Motor Boat", string(b))

	err = (&b).UnmarshalJSON([]byte(`"kaYak"`))
	assert.Nil(t, err)
	assert.Equal(t, "Paddle", string(b))

	// Check that a miscellaneous field is set as a speed boat
	err = (&b).UnmarshalJSON([]byte(`"*"`))
	assert.Nil(t, err)
	assert.Equal(t, "Speed/Motor Boat", string(b))

	// Check that a NULL or empty field is ignored
	b = BoatTypeEnum("")
	err = (&b).UnmarshalJSON([]byte(`null`))
	assert.Nil(t, err)
	assert.Equal(t, "", string(b))
	err = (&b).UnmarshalJSON([]byte(`""`))
	assert.Nil(t, err)
	assert.Equal(t, "", string(b))
}

func TestPropulsionTypeEnum(t *testing.T) {
	var p PropulsionEnum
	err := (&p).UnmarshalJSON([]byte(`"sail"`))
	assert.Nil(t, err)
	assert.Equal(t, "Sail", string(p))

	err = (&p).UnmarshalJSON([]byte(`"Single OUTBOARD"`))
	assert.Nil(t, err)
	assert.Equal(t, "Single Outboard", string(p))

	// Expected because we will set quantity from the TripWatch quantity field
	err = (&p).UnmarshalJSON([]byte(`"Double OUTBOARD"`))
	assert.Nil(t, err)
	assert.Equal(t, "Single Outboard", string(p))

	err = (&p).UnmarshalJSON([]byte(`"inboARD"`))
	assert.Nil(t, err)
	assert.Equal(t, "Single Inboard", string(p))

	err = (&p).UnmarshalJSON([]byte(`"Paddles"`))
	assert.Nil(t, err)
	assert.Equal(t, "Oars", string(p))

	err = (&p).UnmarshalJSON([]byte(`"Oars"`))
	assert.Nil(t, err)
	assert.Equal(t, "Oars", string(p))

	err = (&p).UnmarshalJSON([]byte(`"WIND"`))
	assert.Nil(t, err)
	assert.Equal(t, "Sail", string(p))

	err = (&p).UnmarshalJSON([]byte(`"SAILING"`))
	assert.Nil(t, err)
	assert.Equal(t, "Sail", string(p))

	// Check that a NULL or empty field is ignored
	p = PropulsionEnum("")
	err = (&p).UnmarshalJSON([]byte(`null`))
	assert.Nil(t, err)
	assert.Equal(t, "", string(p))
	err = (&p).UnmarshalJSON([]byte(`""`))
	assert.Nil(t, err)
	assert.Equal(t, "", string(p))
	err = (&p).UnmarshalJSON([]byte(`"   "`))
	assert.Nil(t, err)
	assert.Equal(t, "", string(p))
}

func TestPropulsionUpdateFromEngineQTY(t *testing.T) {
	p := PropulsionEnum("Single Inboard")
	qty := 1
	assert.Nil(t, p.UpdateFromEngineQTY(qty))
	assert.Equal(t, PropulsionEnum("Single Inboard"), p)
	qty = 2
	assert.Nil(t, p.UpdateFromEngineQTY(qty))
	assert.Equal(t, PropulsionEnum("Twin Inboards"), p)
	qty = 6
	assert.Nil(t, p.UpdateFromEngineQTY(qty))
	assert.Equal(t, PropulsionEnum("Twin Inboards"), p)
	p = PropulsionEnum("Inboard")
	qty = 2
	assert.Nil(t, p.UpdateFromEngineQTY(qty))
	assert.Equal(t, PropulsionEnum("Twin Inboards"), p)

	p = PropulsionEnum("Single Outboard")
	qty = 1
	assert.Nil(t, p.UpdateFromEngineQTY(qty))
	assert.Equal(t, PropulsionEnum("Single Outboard"), p)
	qty = 2
	assert.Nil(t, p.UpdateFromEngineQTY(qty))
	assert.Equal(t, PropulsionEnum("Twin Outboards"), p)
	p = PropulsionEnum("Single Outboard")
	qty = 6
	assert.Nil(t, p.UpdateFromEngineQTY(qty))
	assert.Equal(t, PropulsionEnum("Twin Outboards"), p)

	p = PropulsionEnum("Sail")
	qty = 19
	assert.Nil(t, p.UpdateFromEngineQTY(qty))
	assert.Equal(t, PropulsionEnum("Sail"), p)

	p = PropulsionEnum("Kayak")
	qty = 2
	assert.Nil(t, p.UpdateFromEngineQTY(qty))
	assert.Equal(t, PropulsionEnum("Kayak"), p)
}

func TestJobSource(t *testing.T) {
	var j JobSource
	err := (&j).UnmarshalJSON([]byte(`"VMR"`))
	assert.Nil(t, err)
	assert.Equal(t, "Base", string(j))

	err = (&j).UnmarshalJSON([]byte(`"Water Police"`))
	assert.Nil(t, err)
	assert.Equal(t, "Police", string(j))

	err = (&j).UnmarshalJSON([]byte(`"Land Police"`))
	assert.Nil(t, err)
	assert.Equal(t, "Police", string(j))

	err = (&j).UnmarshalJSON([]byte(`"Ambulance Service"`))
	assert.Nil(t, err)
	assert.Equal(t, "QAS", string(j))
}

func TestJobSourceToJobFreq(t *testing.T) {
	assert.Equal(t, JobFreq("Telephone"), JobSource("QAS").ToJobFreq())
	assert.Equal(t, JobFreq("Unit Counter Inquiry"), JobSource("Base").ToJobFreq())

	// Ensure no change is made if there's nothing pre-filled
	assert.Equal(t, JobFreq(""), JobSource("QFES").ToJobFreq())
}
