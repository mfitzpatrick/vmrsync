package main

import (
	"database/sql/driver"
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

type VMRVessel struct {
	ID             int        `firebird:"JOBDUTYVESSELNO" json:"activationsrvsequence"`
	Name           string     `firebird:"JOBDUTYVESSELNAME,match" len:"30" json:"activationsrvvessel"`
	StartHoursPort IntString  `firebird:"JOBHOURSSTART" json:"activationsrvenginehours1start"`
	StartHoursStbd IntString  `json:"activationsrvenginehours2start"`
	EndHoursPort   IntString  `firebird:"JOBHOURSEND" json:"activationsrvenginehours1end"`
	EndHoursStbd   IntString  `json:"activationsrvenginehours2end"`
	Master         string     `json:"activationsrvmaster"`
	CrewList       StringList `json:"activationsrvcrew"`
}

type AssistedVessel struct {
	Rego       string     `firebird:"JOBVESSELREGO" len:"10" json:"activationsdvvesselsregistration"`
	Name       string     `firebird:"JOBVESSELNAME" len:"30" json:"activationsdvvesselsname"`
	Length     LengthEnum `firebird:"JOBLOA" len:"10" json:"activationsdvvesselslength"`
	Type       string     `firebird:"JOBVESSELTYPE" len:"20" json:"activationsdvvesselstype"`
	Propulsion string     `firebird:"JOBPROPULSION" len:"20" json:"activationsdvvesselsenginetype"`
	EngineQTY  int        `json:"activationsdvvesselsenginequantity"`
	NumAdults  int        `firebird:"JOBADULTS" json:"activationsdvpobadult"`
	NumKids    int        `firebird:"JOBCHILDREN" json:"activationsdvpobchildren"`
}

type Emergency struct {
	Emergency         CustomBool     `firebird:"JOBEMERGENCY" len:"1"`
	PoliceNum         string         `firebird:"JOBQASNO" len:"10" json:"activationspoliceincidentnumber"`
	Notified          CustomBool     `firebird:"JOBPOLICE" len:"1" json:"activationspolicenotified"`
	PoliceName        string         `json:"activationspolicenotifiedcontact"`
	Time              CustomJSONTime `json:"activationspolicenotifiedtime"`
	AgenciesAttending StringList     `json:"activationsqasattending"`
}

type GPS struct {
	TWLatLong string  `json:"activationsposition"`
	Lat       float64 `firebird:"JOBLATDEC"`
	Long      float64 `firebird:"JOBLONDEC"`

	// Breaking it down to DMS for Firebird
	LatD  int     `firebird:"JOBLATDEG"`
	LatM  int     `firebird:"JOBLATMIN"`
	LatS  float64 `firebird:"JOBLATSEC"`
	LongD int     `firebird:"JOBLONDEG"`
	LongM int     `firebird:"JOBLONMIN"`
	LongS float64 `firebird:"JOBLONSEC"`
}

type Weather struct {
	Forecast  string        `json:"activationsactivationweatherforecast"`
	WindSpeed WindSpeedEnum `firebird:"JOBWINDSPEED" len:"20"`
	WindDir   WindDirEnum   `firebird:"JOBWINDDIRECTION" len:"3"`
	RainState string        `firebird:"JOBWEATHER" len:"20"`
}

type Job struct {
	DutyLogID   int            `firebird:"JOBDUTYSEQUENCE"`
	ID          int            `firebird:"JOBJOBSEQUENCE,id"`
	StartTime   CustomJSONTime `firebird:"JOBTIMEOUT,match" json:"activationsrvdeparttime"`
	EndTime     CustomJSONTime `firebird:"JOBTIMEIN" json:"activationsrvreturntime"`
	Type        string         `firebird:"JOBTYPE" len:"20" json:"activationstype"`
	Action      string         `firebird:"JOBACTIONTAKEN" len:"20" json:"activationsdvactionrequested"`
	Purpose     string         `json:"activationspurpose"`
	Comments    string         `firebird:"JOBDETAILS" len:"96" json:"activationscomments"`
	Donation    IntString      `firebird:"JOBDONATION" json:"activationsdonationreceived"`
	Frequency   string         `firebird:"JOBFREQUENCY" len:"30"`
	WaterLimits string         `firebird:"JOBWATERLIMITS" len:"20" json:"activationsoperationsareaclassification"`
	SeaState    SeaStateEnum   `firebird:"JOBSEAS" len:"20" json:"activationsobservedseastate"`
	Commercial  CustomBool     `firebird:"JOBCOMMERCIALVESSEL" len:"1"`
	VMRVessel
	AssistedVessel
	Emergency
	GPS
	Weather
}

type linkActivationDB struct {
	ID      int            `json:"id"`
	Created CustomJSONTime `json:"created_at"`
	Updated CustomJSONTime `json:"updated_at"`
	Job     `firebird:"DUTYJOBS"`
}

type DutyLogTable struct {
	DutyLog struct {
		ID       int       `firebird:"DUTYSEQUENCE,id"`
		Date     time.Time `firebird:"DUTYDATE"`
		CrewName string    `firebird:"CREW" len:"10"`
	} `firebird:"DUTYLOG"`
}

type Member struct {
	ID     int    `firebird:"MEMBERNOLOCAL,id"`
	Email1 string `firebird:"EMAIL1" len:"96"`
	Email2 string `firebird:"EMAIL2" len:"96"`
}

type CrewOnDuty struct {
	ID       int `firebird:"DUTYSEQUENCE,id"`
	MemberNo int `firebird:"CREWMEMBER" join:"MEMBERS"`
	RankID   int `firebird:"CREWRANKING"`
}

type JobCrew struct {
	DutyCrewID int        `firebird:"CREWDUTYSEQUENCE,match" join:"DUTYCREWS"`
	JobID      int        `firebird:"CREWJOBSEQUENCE,match"`
	MemberID   int        `firebird:"CREWMEMBER,match" join:"MEMBERS"`
	RankID     int        `firebird:"CREWRANKING"`
	IsMaster   CustomBool `firebird:"SKIPPER" len:"1"`
	IsOnJob    CustomBool `firebird:"CREWONJOB" len:"1"`
	email      string     //Copy of the member's email address if read from DB
}

type crewInfo struct {
	Member     `firebird:"MEMBERS"`
	CrewOnDuty `firebird:"DUTYCREWS"`
}

type crewOnJob struct {
	Member     `firebird:"MEMBERS"`
	CrewOnDuty `firebird:"DUTYCREWS"`
	JobCrew    `firebird:"DUTYJOBSCREW"`
}

/*
 * The custom types here are used to wrap the data type received in JSON with the data type
 * expected by Firebase. Sometimes the data types of each are incompatible (i.e. an integer
 * wrapped in a string). So, for the purposes of conversion the custom data type's underlying
 * type is equivalent to Firebase's type, and the JSON unmarshalling function parses the type
 * received to create an object of Firebase's type.
 */

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

func (tm CustomJSONTime) String() string {
	return time.Time(tm).String()
}

func (tm CustomJSONTime) Value() (driver.Value, error) {
	return time.Time(tm), nil
}

type CustomBool string //TripWatch boolean contained as a string or normal bool

func (b *CustomBool) UnmarshalJSON(bytes []byte) error {
	rawString := strings.ToLower(strings.Trim(strings.TrimSpace(string(bytes)), "\""))
	var realBool bool
	if err := json.Unmarshal([]byte(rawString), &realBool); err == nil {
		if realBool {
			*b = CustomBool("Y")
		} else {
			*b = CustomBool("N")
		}
		return nil
	} else if rawString == "null" {
		*b = CustomBool("N")
		return nil
	} else {
		switch rawString {
		case "null", "false", "no":
			*b = CustomBool("N")
		case "true", "yes":
			*b = CustomBool("Y")
		default:
			return errors.Errorf("CustomBool JSON unmarshal of '%s' failed", rawString)
		}
		return nil
	}
}

func (b CustomBool) AsBool() bool {
	return (b == "Y")
}

type IntString float32 //TripWatch floating-point number contained in a string

func (i *IntString) UnmarshalJSON(bytes []byte) error {
	rawString := strings.Trim(strings.TrimSpace(string(bytes)), "\"")
	if rawString == "null" {
		// Special case - ignore NULL
		*i = IntString(0.0)
		return nil
	}
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

type LengthEnum string // Firebird enumerated length range

func (l *LengthEnum) UnmarshalJSON(bytes []byte) error {
	rawString := strings.Trim(strings.TrimSpace(string(bytes)), "\"")
	if rawString == "null" {
		// Special case - ignore NULL
		*l = LengthEnum("")
		return nil
	}
	if val, err := strconv.ParseFloat(rawString, 32); err != nil {
		return errors.Wrapf(err, "unmarshal LengthEnum %s", string(bytes))
	} else {
		// Set length in metres to a string enum representing the range it lies in
		lenRange := ""
		switch length := val; {
		case length < 8:
			lenRange = "0 - 8m"
		case length < 12:
			lenRange = "8 - 12m"
		default:
			lenRange = "> 12m"
		}
		*l = LengthEnum(lenRange)
	}
	return nil
}

type WindSpeedEnum string // Firebird enumerated speed range

func (w *WindSpeedEnum) UnmarshalJSON(bytes []byte) error {
	if val, err := strconv.ParseFloat(string(bytes), 32); err != nil {
		return errors.Wrapf(err, "unmarshal LengthEnum %s", string(bytes))
	} else {
		w.Set(int(val))
	}
	return nil
}

// Set speed in knots to a string enum representing the range it lies in
func (w *WindSpeedEnum) Set(knots int) {
	switch speed := knots; {
	case speed < 10:
		*w = WindSpeedEnum("0 - 10kt")
	case speed <= 20:
		*w = WindSpeedEnum("10 - 20kt")
	default:
		*w = WindSpeedEnum("> 20kt")
	}
}

type WindDirEnum string // Firebird enumerated direction

func (w *WindDirEnum) UnmarshalJSON(bytes []byte) error {
	// Set length in metres to a string enum representing the range it lies in
	enum := ""
	switch dir := strings.ToLower(string(bytes)); dir {
	case "south", "s":
		enum = "S"
	case "south-east", "south east", "se":
		enum = "SE"
	case "east", "e":
		enum = "E"
	case "north-east", "north east", "ne":
		enum = "NE"
	case "north", "n":
		enum = "N"
	case "north-west", "north west", "nw":
		enum = "NW"
	case "west", "w":
		enum = "W"
	case "south-west", "south west", "sw":
		enum = "SW"
	default:
		enum = "> 20kt"
	}
	*w = WindDirEnum(enum)
	return nil
}

type SeaStateEnum string // Firebird enumerated sea state

func (s *SeaStateEnum) UnmarshalJSON(bytes []byte) error {
	if string(bytes) == "null" {
		// Special case - ignore NULL
		*s = SeaStateEnum("")
		return nil
	}

	if stateID, err := strconv.ParseInt(string(bytes), 10, 32); err != nil {
		return errors.Wrapf(err, "SeaStateEnum JSON int parse failed")
	} else {
		switch id := stateID; {
		case id <= 3:
			*s = SeaStateEnum("Calm")
		case id == 4, id == 5:
			*s = SeaStateEnum("Moderate")
		default:
			*s = SeaStateEnum("Rough")
		}
	}
	return nil
}

// Contains a TripWatch JSON list that has been encoded as a single string
type StringList []string

func (s *StringList) UnmarshalJSON(bytes []byte) error {
	rawString := strings.TrimSpace(string(bytes))
	if rawString[0] == '"' || rawString[0] == '\'' {
		if unquotedString, err := strconv.Unquote(rawString); err != nil {
			return errors.Wrapf(err, "StringList couldn't unquote string")
		} else {
			rawString = unquotedString
		}
	}
	var sList []string
	if err := json.Unmarshal([]byte(rawString), &sList); err != nil {
		return errors.Wrapf(err, "StringList failed to parse JSON '%s'", rawString)
	} else {
		*s = StringList(sList)
		return nil
	}
}

func (s StringList) Has(email string) bool {
	for _, memberEmail := range s {
		if memberEmail == email {
			return true
		}
	}
	return false
}
