package main

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
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
	ID             int               `firebird:"JOBDUTYVESSELNO" json:"activationsrvsequence"`
	Name           VMRVesselNameEnum `firebird:"JOBDUTYVESSELNAME,match" len:"30" json:"activationsrvvessel"`
	StartHoursPort IntString         `firebird:"JOBHOURSSTART" json:"activationsrvenginehours1start"`
	StartHoursStbd IntString         `json:"activationsrvenginehours2start"`
	EndHoursPort   IntString         `firebird:"JOBHOURSEND" json:"activationsrvenginehours1end"`
	EndHoursStbd   IntString         `json:"activationsrvenginehours2end"`
	Master         string            `json:"activationsrvmaster"`
	CrewList       StringList        `json:"activationsrvcrew"`
}

type AssistedVessel struct {
	Rego       string         `firebird:"JOBVESSELREGO" len:"10" json:"activationsdvvesselsregistration"`
	Name       string         `firebird:"JOBVESSELNAME" len:"30" json:"activationsdvvesselsname"`
	Length     LengthEnum     `firebird:"JOBLOA" len:"10" json:"activationsdvvesselslength"`
	Type       BoatTypeEnum   `firebird:"JOBVESSELTYPE" len:"20" json:"activationsdvvesselstype"`
	Propulsion PropulsionEnum `firebird:"JOBPROPULSION" len:"20" json:"activationsdvvesselsenginetype"`
	EngineQTY  int            `json:"activationsdvvesselsenginequantity"`
	NumAdults  int            `firebird:"JOBADULTS" json:"activationsdvpobadult"`
	NumKids    int            `firebird:"JOBCHILDREN" json:"activationsdvpobchildren"`
	Phone      IntString      `json:"activationsdvcontactnumber"`
	RadioChan  IntString      `json:"activationsdvradiochannel"`
}

type Emergency struct {
	Emergency         CustomBool     `firebird:"JOBEMERGENCY" len:"1"`
	PoliceNum         string         `firebird:"JOBQASNO" len:"10" json:"activationspoliceincidentnumber"`
	Notified          CustomBool     `firebird:"JOBPOLICE" len:"1" json:"activationspolicenotified"`
	PoliceName        string         `json:"activationspolicenotifiedcontact"`
	Time              CustomJSONTime `json:"activationspolicenotifiedtime"`
	AgenciesAttending StringList     `json:"activationsqasattending"`
}

type FirebirdGPS struct {
	Lat  float64 `firebird:"JOBLATDEC"`
	Long float64 `firebird:"JOBLONDEC"`

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
	DutyLogID   int             `firebird:"JOBDUTYSEQUENCE,id"` //NB: hacky use of the `id` label. Don't change order.
	ID          int             `firebird:"JOBJOBSEQUENCE,id"`
	Status      string          `json:"activationsstatus"`
	StartTime   CustomJSONTime  `firebird:"JOBTIMEOUT,match" json:"activationsrvdeparttime"`
	EndTime     CustomJSONTime  `firebird:"JOBTIMEIN" json:"activationsrvreturntime"`
	Type        JobType         `firebird:"JOBTYPE" len:"20" json:"activationstype"`
	Action      JobAction       `firebird:"JOBACTIONTAKEN" len:"20" json:"activationsdvactionrequested"`
	Purpose     string          `firebird:"JOBDETAILS" len:"96" json:"activationspurpose"`
	Comments    string          `firebird:"JOBDETAILS_LONG" len:"4096" json:"activationscomments"`
	Donation    IntString       `firebird:"JOBDONATION" json:"activationsdonationreceived"`
	Frequency   string          `firebird:"JOBFREQUENCY" len:"30"`
	WaterLimits WaterLimitsEnum `firebird:"JOBWATERLIMITS" len:"20" json:"activationsoperationsareaclassification"`
	SeaState    SeaStateEnum    `firebird:"JOBSEAS" len:"20" json:"activationsobservedseastate"`
	Commercial  CustomBool      `firebird:"JOBCOMMERCIALVESSEL" len:"1"`
	Pos         GPS             `json:"activationsposition"`
	ActivatedBy JobSource       `firebird:"JOBACTIVATION" len:"20" json:"activationssource"`
	Freq        JobFreq         `firebird:"JOBFREQUENCY" len:"30"`
	AssistNum   IntString       `firebird:"JOBASSISTNO" json:"activationsdonationreceiptnumber"`
	VMRVessel
	AssistedVessel
	Emergency
	FirebirdGPS
	Weather
}

type linkActivationDB struct {
	ID      int            `json:"id"`
	Created CustomJSONTime `json:"created_at"`
	Updated CustomJSONTime `json:"updated_at"`
	Job     `firebird:"DUTYJOBS"`
	Sitreps []Sitrep
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

// This matches the data returned by the /activationtransactions API.
type Sitrep struct {
	Updated CustomJSONTime `json:"updated_at"`
	Pos     GPS            `json:"activationstransactionscurrentposition"`
	Comment string         `json:"activationstransactionsnote"`
}

var sitrepNotFoundError = errors.New("sitrep not found")

func getEntryForComment(s []Sitrep, comment string) (Sitrep, error) {
	for _, sr := range s {
		if strings.HasPrefix(sr.Comment, comment) && !sr.Pos.IsZero() {
			return sr, nil
		}
	}
	return Sitrep{}, errors.Wrapf(sitrepNotFoundError, "get entry for comment %s", comment)
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

// Convert the time object to the UTC+10 timezone
func (tm CustomJSONTime) AEST() time.Time {
	tz := time.FixedZone("UTC+10", 10*60*60)
	return time.Time(tm).In(tz)
}

// Called before a CustomJSONTime object is written to the DB.
// As the DB requires the TS to be in the AEST timezone, this first calls the AEST() function to convert
// the time value appropriately.
func (tm CustomJSONTime) Value() (driver.Value, error) {
	return tm.AEST(), nil
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

func (i IntString) IsZero() bool {
	return i == IntString(0.0)
}

type LengthEnum string // Firebird enumerated length range

func (l *LengthEnum) UnmarshalJSON(bytes []byte) error {
	const FT_CONV_FACTOR = 0.3048
	rawString := strings.Trim(strings.TrimSpace(string(bytes)), "\"")
	if rawString == "null" {
		// Special case - ignore NULL
		*l = LengthEnum("")
		return nil
	}
	// Check for units indicator of ' or m trailing characters
	isFeet := false
	if strings.HasSuffix(rawString, "'") || strings.HasSuffix(rawString, "\u2019") ||
		strings.HasSuffix(rawString, "f") {
		isFeet = true
	}
	rawString = strings.TrimRight(rawString, "mf'\u2019 ") // Remove feet or metres indicator character
	if val, err := strconv.ParseFloat(rawString, 32); err != nil {
		return errors.Wrapf(err, "unmarshal LengthEnum %s", string(bytes))
	} else {
		if isFeet {
			val = val * FT_CONV_FACTOR
		}
		// Set length in metres to a string enum representing the range it lies in
		lenRange := ""
		switch length := val; {
		case length < 4.5:
			lenRange = "<4.5m"
		case length < 8:
			lenRange = "4.5m - 8m"
		case length < 10:
			lenRange = "8m - 10m"
		case length < 15:
			lenRange = "10m - 15m"
		case length < 25:
			lenRange = "15m - 25m"
		default:
			lenRange = "25m +"
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
		*w = WindSpeedEnum("0 - 10 knots")
	case speed <= 20:
		*w = WindSpeedEnum("10 - 20 knots")
	default:
		*w = WindSpeedEnum("20+ knots")
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

type JobType string

func (j *JobType) UnmarshalJSON(bytes []byte) error {
	var jt string
	if err := json.Unmarshal(bytes, &jt); err != nil {
		return errors.Wrapf(err, "JobType parse JSON '%s'", string(bytes))
	} else {
		switch jt {
		case "Medivac":
			*j = JobType("Medical")
		case "SAR":
			*j = JobType("Search")
		case "Assist":
			*j = JobType("Breakdown")
		case "Training":
			*j = JobType("Training/Patrol")
		case "Scattering of Ashes":
			*j = JobType("Dispersal")
		case "Public Service":
			*j = JobType("PR/Promo")
		case "MAYDAY", "PANPAN":
			*j = JobType("EPIRB")
		default:
			*j = JobType(jt)
		}
		return nil
	}
}

func (j JobType) ToJobAction() JobAction {
	var action JobAction
	switch j {
	case "Training/Patrol":
		action = JobAction("Training")
	case "Medical":
		action = JobAction("Medivac")
	default:
		action = JobAction("Other")
	}
	return action
}

func (j JobType) IsZero() bool {
	return j == JobType("")
}

type JobAction string

func (j *JobAction) UnmarshalJSON(bytes []byte) error {
	var ja string
	if err := json.Unmarshal(bytes, &ja); err != nil {
		return errors.Wrapf(err, "JobAction parse JSON '%s'", string(bytes))
	} else {
		lja := strings.ToLower(ja)
		switch {
		case strings.Contains(lja, "jump"):
			*j = JobAction("Jump Start")
		case strings.Contains(lja, "medivac"),
			strings.Contains(lja, "medevac"),
			strings.Contains(lja, "medical"):
			*j = JobAction("Medivac")
		case strings.Contains(lja, "nil"):
			*j = JobAction("Nil")
		case strings.Contains(lja, "pump"):
			*j = JobAction("Pump Out")
		case strings.Contains(lja, "search"),
			strings.Contains(lja, "sar"):
			*j = JobAction("Search & Rescue")
		case strings.Contains(lja, "fuel"):
			*j = JobAction("Supplied Fuel")
		case strings.Contains(lja, "tow"):
			*j = JobAction("Tow")
		case strings.Contains(lja, "train"):
			*j = JobAction("Training")
		case strings.Contains(lja, "unground"):
			*j = JobAction("Ungrounded")
		case strings.Contains(lja, "investigate"):
			*j = JobAction("Investigate")
		default:
			*j = JobAction("Other")
		}
		return nil
	}
}

func (j JobAction) IsZero() bool {
	return j == JobAction("")
}

type WaterLimitsEnum string

func (w *WaterLimitsEnum) UnmarshalJSON(bytes []byte) error {
	var wl string
	if err := json.Unmarshal(bytes, &wl); err != nil {
		return errors.Wrapf(err, "WaterLimitsEnum parse JSON '%s'", string(bytes))
	} else {
		switch wl {
		case "A", "B", "C":
			*w = WaterLimitsEnum("Open")
		case "D":
			*w = WaterLimitsEnum("Partially Smooth")
		case "E":
			*w = WaterLimitsEnum("Smooth")
		}
		return nil
	}
}

type VMRVesselNameEnum string

func (n *VMRVesselNameEnum) UnmarshalJSON(bytes []byte) error {
	var vn string
	if err := json.Unmarshal(bytes, &vn); err != nil {
		return errors.Wrapf(err, "VMRVesselNameEnum parse JSON '%s'", string(bytes))
	} else {
		switch vn {
		case "MARINERESCUE1":
			*n = VMRVesselNameEnum("Marine Rescue 1")
		case "MARINERESCUE2":
			*n = VMRVesselNameEnum("Marine Rescue 2")
		case "MARINERESCUE4":
			*n = VMRVesselNameEnum("Marine Rescue 4")
		case "MARINERESCUE5":
			*n = VMRVesselNameEnum("Marine Rescue 5")
		}
		return nil
	}
}

type BoatTypeEnum string

func (b *BoatTypeEnum) UnmarshalJSON(bytes []byte) error {
	var bn string
	if err := json.Unmarshal(bytes, &bn); err != nil {
		return errors.Wrapf(err, "BoatTypeEnum parse JSON '%s'", string(bytes))
	} else {
		bn = strings.ToLower(bn)
		switch {
		case strings.Contains(bn, "jet ski"), strings.Contains(bn, "jetski"):
			*b = "PWC"
		case strings.Contains(bn, "yacht"), strings.Contains(bn, "sail"),
			strings.Contains(bn, "ketch"), strings.Contains(bn, "schooner"):
			*b = "Sailing"
		case strings.Contains(bn, "kayak"), strings.Contains(bn, "paddle"):
			*b = "Paddle"
		default:
			if len(bn) > 0 {
				*b = "Speed/Motor Boat"
			}
		}
	}
	return nil
}

type PropulsionEnum string

func (p *PropulsionEnum) UnmarshalJSON(bytes []byte) error {
	if string(bytes) == "null" {
		// Special case - ignore NULL
		*p = PropulsionEnum("")
		return nil
	}
	var pn string
	if err := json.Unmarshal(bytes, &pn); err != nil {
		return errors.Wrapf(err, "PropulsionEnum parse JSON '%s'", string(bytes))
	} else {
		pn = strings.ToLower(pn)
		if len(strings.TrimSpace(pn)) == 0 {
			// Special case - ignore empty string
			*p = PropulsionEnum("")
			return nil
		}
		switch {
		case strings.Contains(pn, "outboard"):
			*p = "Single Outboard"
		case strings.Contains(pn, "inboard"):
			*p = "Single Inboard"
		case strings.Contains(pn, "paddle"), strings.Contains(pn, "oar"):
			*p = "Oars"
		case strings.Contains(pn, "wind"), strings.Contains(pn, "sail"):
			*p = "Sail"
		default:
			*p = "Single Outboard"
		}
	}
	return nil
}

func (p *PropulsionEnum) UpdateFromEngineQTY(qty int) error {
	var prefix string
	var plural string
	switch qty {
	case 1:
		prefix = "Single"
	default:
		prefix = "Twin"
		plural = "s"
	}
	var suffix string
	if strings.Contains(string(*p), "Outboard") {
		suffix = "Outboard"
	} else if strings.Contains(string(*p), "Inboard") {
		suffix = "Inboard"
	}
	if suffix != "" {
		*p = PropulsionEnum(fmt.Sprintf("%s %s%s", prefix, suffix, plural))
	}
	return nil
}

type JobSource string

func (j *JobSource) UnmarshalJSON(bytes []byte) error {
	var js string
	if err := json.Unmarshal(bytes, &js); err != nil {
		return errors.Wrapf(err, "JobSource parse JSON '%s'", string(bytes))
	} else {
		switch js {
		case "Water Police", "Land Police":
			*j = "Police"
		case "Ambulance Service":
			*j = "QAS"
		default:
			*j = "Base"
		}
	}
	return nil
}

func (j JobSource) ToJobFreq() JobFreq {
	var jf JobFreq
	switch j {
	case "QAS", "Police":
		jf = "Telephone"
	case "Base":
		jf = "Unit Counter Inquiry"
	}
	return jf
}

type JobFreq string

func (j JobFreq) IsZero() bool {
	return j == JobFreq("")
}
