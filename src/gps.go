package main

import (
	"math"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type DMS struct {
	Hemisphere bool // True is north or east, false is south or west
	Deg        int
	Min        int
	Sec        float64
}

type GPS_DMS struct {
	Lat  DMS
	Long DMS
}

type GPS struct {
	Lat  float64
	Long float64
}

func (g *GPS) UnmarshalJSON(bytes []byte) error {
	// Strip out quotation marks from JSON if any are present
	rawString := strings.TrimSpace(string(bytes))
	if rawString[0] == '"' || rawString[0] == '\'' {
		if unquotedString, err := strconv.Unquote(rawString); err != nil {
			return errors.Wrapf(err, "StringList couldn't unquote string")
		} else {
			rawString = strings.TrimSpace(unquotedString)
		}
	}
	if rawString == "null" || rawString == "" {
		g.Lat = 0
		g.Long = 0
		return nil
	}

	// Split string into pieces based on spaces or commas
	if floats, err := pullFloatsFromString(rawString); err != nil {
		return errors.Wrapf(err, "unmarshal GPS float extraction from '%s'", rawString)
	} else if len(floats) == 1 && floats[0] == 0 {
		// Special case - a zero-value is input to ignore this field. Leave the lat and long
		// values as 0 and do no further actions.
	} else if len(floats) != 2 {
		return errors.Errorf("unmarshal GPS expected exactly 2 numbers, got %d from %s",
			len(floats), rawString)
	} else if math.Abs(floats[0]) > 90 || math.Abs(floats[1]) > 180 {
		return errors.Errorf("GPS position out of range (%f, %f)", floats[0], floats[1])
	} else {
		g.Lat = floats[0]
		g.Long = floats[1]
	}
	return nil
}

// Convert the GPS struct contents to Degrees, Minutes, and Decimal Seconds
func (g GPS) AsDMS() (GPS_DMS, error) {
	return GPS_DMS{
		Lat:  dmsFromDD(g.Lat),
		Long: dmsFromDD(g.Long),
	}, nil
}

func (g GPS) IsZero() bool {
	return (g.Lat == 0.0 && g.Long == 0.0)
}

// Convert a floating-point decimal-degrees value to a value represented by the DMS struct.
// The Hemisphere field is determined by the sign of the floating point value. True
// values = northern and eastern hemispheres, and false values = southern and western hemispheres.
func dmsFromDD(dd float64) DMS {
	deg, degf := math.Modf(math.Abs(dd))
	min, minf := math.Modf(degf * 60)
	return DMS{
		Hemisphere: (dd >= 0.0),
		Deg:        int(deg),
		Min:        int(min),
		Sec:        minf * 60,
	}
}

func pullFloatsFromString(in string) ([]float64, error) {
	var floats []float64
	splitStr := strings.Split(in, ",")
	if len(splitStr) == 1 {
		// Try splitting by other characters
		splitStr = strings.Split(in, ":")
		if len(splitStr) == 1 {
			// Finally, try splitting by space
			splitStr = strings.Split(in, " ")
		}
	}
	for _, v := range splitStr {
		v = strings.Trim(strings.TrimSpace(v), "\"")
		if len(v) > 0 {
			if float, err := strconv.ParseFloat(v, 64); err == nil {
				floats = append(floats, float)
			}
		}
	}
	return floats, nil
}
