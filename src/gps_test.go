package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGPS(t *testing.T) {
	var g GPS
	err := (&g).UnmarshalJSON([]byte(`"-27.475458084334857 153.15326141723338 128.55"`))
	assert.NotNil(t, err)
	err = (&g).UnmarshalJSON([]byte(`"-27.475458084334857"`))
	assert.NotNil(t, err)
	err = (&g).UnmarshalJSON([]byte(`"-95.475458084334857 153.15326141723338"`))
	assert.NotNil(t, err)
	err = (&g).UnmarshalJSON([]byte(`"-27.475458084334857 190.15326141723338"`))
	assert.NotNil(t, err)

	err = (&g).UnmarshalJSON([]byte(`"-27.475458084334857 153.15326141723338"`))
	assert.Nil(t, err)
	assert.Equal(t, GPS{Lat: -27.475458084334857, Long: 153.15326141723338}, g)

	err = (&g).UnmarshalJSON([]byte(`" -27.475458084334857   153.15326141723338  "`))
	assert.Nil(t, err)
	assert.Equal(t, GPS{Lat: -27.475458084334857, Long: 153.15326141723338}, g)

	err = (&g).UnmarshalJSON([]byte(`"-27.475458084334857,153.15326141723338"`))
	assert.Nil(t, err)
	assert.Equal(t, GPS{Lat: -27.475458084334857, Long: 153.15326141723338}, g)

	err = (&g).UnmarshalJSON([]byte(`"-27.475458084334857 ,   153.15326141723338"`))
	assert.Nil(t, err)
	assert.Equal(t, GPS{Lat: -27.475458084334857, Long: 153.15326141723338}, g)

	err = (&g).UnmarshalJSON([]byte(`"\"-27.475458084334857 153.15326141723338\""`))
	assert.Nil(t, err)
	assert.Equal(t, GPS{Lat: -27.475458084334857, Long: 153.15326141723338}, g)

	err = (&g).UnmarshalJSON([]byte(`"\"-27.475458084334857\",\"153.15326141723338\""`))
	assert.Nil(t, err)
	assert.Equal(t, GPS{Lat: -27.475458084334857, Long: 153.15326141723338}, g)

	// Test that this can be converted to DMS easily
	dms, err := g.AsDMS()
	assert.Nil(t, err)
	assert.Equal(t, GPS_DMS{
		Lat: DMS{
			Hemisphere: false,
			Deg:        27,
			Min:        28,
			Sec:        31.649103605485323,
		},
		Long: DMS{
			Hemisphere: true,
			Deg:        153,
			Min:        9,
			Sec:        11.741102040170972,
		},
	}, dms)

	// Check zero-value handling
	g = GPS{}
	err = (&g).UnmarshalJSON([]byte(`"0"`))
	assert.Nil(t, err)
	assert.Equal(t, GPS{Lat: 0, Long: 0}, g)
	err = (&g).UnmarshalJSON([]byte(`"0 0  "`))
	assert.Nil(t, err)
	assert.Equal(t, GPS{Lat: 0, Long: 0}, g)
	err = (&g).UnmarshalJSON([]byte(`"  0   "`))
	assert.Nil(t, err)
	assert.Equal(t, GPS{Lat: 0, Long: 0}, g)

	// Check NULL-value handling
	err = (&g).UnmarshalJSON([]byte(`null`))
	assert.Nil(t, err)
	assert.Equal(t, GPS{Lat: 0, Long: 0}, g)
}

func TestDMSFromDD(t *testing.T) {
	assert.Equal(t, DMS{
		Hemisphere: true,
		Deg:        27,
		Min:        50,
		Sec:        33.5039999999978,
	}, dmsFromDD(27.84264))
	assert.Equal(t, DMS{
		Hemisphere: false,
		Deg:        27,
		Min:        50,
		Sec:        33.5039999999978,
	}, dmsFromDD(-27.84264))
}

func TestPullFloatsFromString(t *testing.T) {
	floats, err := pullFloatsFromString("-27.1 153.1")
	assert.Nil(t, err)
	assert.Equal(t, []float64{-27.1, 153.1}, floats)

	floats, err = pullFloatsFromString("-27 153")
	assert.Nil(t, err)
	assert.Equal(t, []float64{-27, 153.0}, floats)

	floats, err = pullFloatsFromString("-27.2:153.6")
	assert.Nil(t, err)
	assert.Equal(t, []float64{-27.2, 153.6}, floats)

	floats, err = pullFloatsFromString("-27.2,153.6")
	assert.Nil(t, err)
	assert.Equal(t, []float64{-27.2, 153.6}, floats)

	floats, err = pullFloatsFromString("-27.2, 153.6")
	assert.Nil(t, err)
	assert.Equal(t, []float64{-27.2, 153.6}, floats)

	floats, err = pullFloatsFromString("  -27.2   153.6    ")
	assert.Nil(t, err)
	assert.Equal(t, []float64{-27.2, 153.6}, floats)

	floats, err = pullFloatsFromString("\"-27.2\", \"153.6\"")
	assert.Nil(t, err)
	assert.Equal(t, []float64{-27.2, 153.6}, floats)
}
