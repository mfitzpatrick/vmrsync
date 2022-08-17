package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
