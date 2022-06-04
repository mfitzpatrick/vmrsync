package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomBoolUnmarshal(t *testing.T) {
	var b CustomBool
	err := (&b).UnmarshalJSON([]byte("false"))
	assert.Nil(t, err)
	assert.Equal(t, "f", string(b))

	err = (&b).UnmarshalJSON([]byte("\"Yes\" "))
	assert.Nil(t, err)
	assert.Equal(t, "t", string(b))
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
	assert.Equal(t, "0 - 8m", string(l))

	err = (&l).UnmarshalJSON([]byte(" \"15"))
	assert.Nil(t, err)
	assert.Equal(t, "> 12m", string(l))

	err = (&l).UnmarshalJSON([]byte("null"))
	assert.Nil(t, err)
	assert.Equal(t, "", string(l))
}

func TestWindSpeedEnumUnmarshal(t *testing.T) {
	var w WindSpeedEnum
	err := (&w).UnmarshalJSON([]byte("15"))
	assert.Nil(t, err)
	assert.Equal(t, "10 - 20kt", string(w))
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
