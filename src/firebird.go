package main

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var matchFieldIsZero = errors.Errorf("Zero Key Value Error")
var dbZeroRowsErr = errors.Errorf("No DB Rows Returned")

type dbError struct {
	error
	name      string
	cols      []column
	statement string // Type of statement (insert, update, etc.)
}

func (e dbError) Unwrap() error {
	return e.error
}

func (e dbError) String() string {
	return fmt.Sprintf("DB error on table %s with statement:\n\t%s\n\t",
		e.name, e.statement)
}

func (e dbError) Error() string {
	return fmt.Sprintf("%s: %s", e.String(), e.error.Error())
}

type column struct {
	name       string
	isMatch    bool
	isSequence bool
	maxStrlen  int
	value      interface{}
}

type firebirdColHandler func(tableName string, col column) error

// Recursive function which will call the handler function for each item in the struct.
func forEachColumn(tableName string, obj reflect.Value, handler firebirdColHandler) error {
	if obj.Type().Kind() != reflect.Struct {
		return errors.Errorf("obj should be a struct, not %v", obj.Type().Kind())
	}
	for i := 0; i < obj.NumField(); i++ {
		structVal := obj.Field(i)
		structField := obj.Type().Field(i)
		isMatch := false
		isSequence := false
		maxStrlen := 0
		firebirdTag := structField.Tag.Get("firebird")
		if fbTags := strings.Split(firebirdTag, ","); len(fbTags) > 1 {
			firebirdTag = fbTags[0]
			for _, tag := range fbTags {
				switch tag {
				case "match":
					isMatch = true
				case "id":
					isSequence = true
				}
			}
		}
		if structVal.Kind() == reflect.String && firebirdTag != "" {
			if l, err := strconv.ParseInt(structField.Tag.Get("len"), 10, 32); err != nil {
				return errors.Wrapf(err, "firebird length not parseable in field %s", firebirdTag)
			} else {
				maxStrlen = int(l)
			}
		}
		if structVal.Kind() == reflect.Struct && structVal.Type() != reflect.TypeOf(CustomJSONTime{}) {
			// This references a nested struct. Call this function recursively.
			nestedTable := tableName
			if firebirdTag != "" {
				nestedTable = firebirdTag
			}
			if err := forEachColumn(nestedTable, structVal, handler); err != nil {
				return errors.Wrapf(err, "firebird for each col recursion (tbl %s)", tableName)
			}
		} else if firebirdTag == "" {
			continue //Nothing here matches with the firebird DB
		} else if err := handler(tableName, column{
			name:       firebirdTag,
			isMatch:    isMatch,
			isSequence: isSequence,
			maxStrlen:  maxStrlen,
			value:      structVal.Interface(),
		}); err != nil {
			return errors.Wrapf(err, "firebird for each column (%s.%s) handler", tableName, firebirdTag)
		}
	}
	return nil
}

// Create the update statement and try to execute it against the DB.
func tryUpdate(ctx context.Context, db *sql.DB, tableName string, columns []column) error {
	colList := make([]string, 0, len(columns))
	valList := make([]interface{}, 0, len(columns))
	keyCol := []string{}
	keyVal := []interface{}{}
	for _, column := range columns {
		if column.isSequence {
			// Don't use the sequence numbers in update statements
			continue
		}
		colList = append(colList, column.name)
		valList = append(valList, column.value)
		if column.isMatch {
			keyCol = append(keyCol, column.name)
			keyVal = append(keyVal, column.value)
		}
	}
	if len(keyCol) == 0 {
		return errors.Errorf("no keys specified for table %s", tableName)
	}
	if len(colList) == 0 {
		return errors.Errorf("no columns specified for table %s", tableName)
	}
	stmt := fmt.Sprintf("UPDATE %s SET %s WHERE %s", tableName,
		strings.Join(colList, "=?,")+"=?", strings.Join(keyCol, "=? AND ")+"=?")
	if result, err := db.ExecContext(ctx, stmt, append(valList, keyVal...)...); err != nil {
		return errors.Wrapf(dbError{
			error:     err,
			name:      tableName,
			cols:      columns,
			statement: stmt,
		}, "update errored for table %s", tableName)
	} else if rowCount, err := result.RowsAffected(); err != nil {
		return errors.Wrapf(dbError{
			error:     err,
			name:      tableName,
			cols:      columns,
			statement: stmt,
		}, "trying update can't fetch row count affected")
	} else if rowCount == int64(0) {
		// Update failed - row likely doesn't exist yet.
		return errors.Wrapf(dbError{
			error:     errors.Errorf("RowsAffected is 0"),
			name:      tableName,
			cols:      columns,
			statement: stmt,
		}, "trying update no rows affected")
	}
	return nil
}

// Create the insert statement and try to execute it against the DB.
func tryInsert(ctx context.Context, db *sql.DB, tableName string, columns []column) error {
	getMaxID := func(ctx context.Context, tableName, colName string) (int, error) {
		maxID := 0
		// Statement to get the maximum sequence number from the current DB table
		idStmt := fmt.Sprintf("SELECT MAX(%s) FROM %s", colName, tableName)
		if rows, err := db.QueryContext(ctx, idStmt); err != nil {
			return 0, errors.Wrapf(dbError{
				error:     err,
				name:      tableName,
				cols:      []column{{name: colName}},
				statement: idStmt,
			}, "insert getting next sequence number")
		} else if !rows.Next() {
			return 0, errors.Errorf("insert tx max id rows failed for table %s", tableName)
		} else if err := rows.Scan(&maxID); err != nil {
			return 0, errors.Errorf("insert tx max ID scan failed for table %s", tableName)
		}
		return maxID, nil
	}

	colList := make([]string, 0, len(columns))
	valList := make([]interface{}, 0, len(columns))
	var seqCol string
	var seqIDX int
	for _, column := range columns {
		colList = append(colList, column.name)
		valList = append(valList, column.value)
		if column.isSequence {
			seqCol = column.name
			seqIDX = len(valList) - 1
		}
	}
	if len(colList) == 0 {
		return errors.Errorf("no columns specified for table %s", tableName)
	}
	if seqCol != "" {
		// Get the next logical sequence number for the table
		if maxID, err := getMaxID(ctx, tableName, seqCol); err != nil {
			return errors.Wrapf(err, "insert failed to get max sequence number")
		} else if maxID == 0 {
			return errors.Wrapf(dbError{
				error: errors.Errorf("max sequence number is 0"),
				name:  tableName,
				cols:  []column{{name: seqCol}},
			}, "maxID cannot be 0")
		} else {
			valList[seqIDX] = maxID + 1
		}
	}
	// Statement to insert the new record
	insertStmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName,
		strings.Join(colList, ","),
		strings.TrimRight(strings.Repeat("?,", len(valList)), ","))
	// Run DB statements
	if result, err := db.ExecContext(ctx, insertStmt, valList...); err != nil {
		return errors.Wrapf(dbError{
			error:     err,
			name:      tableName,
			cols:      columns,
			statement: insertStmt,
		}, "insert errored for table %s", tableName)
	} else if rowCount, err := result.RowsAffected(); err != nil {
		return errors.Wrapf(dbError{
			error:     err,
			name:      tableName,
			cols:      columns,
			statement: insertStmt,
		}, "trying insert can't fetch row count affected")
	} else if rowCount == int64(0) {
		// Update failed - row likely doesn't exist yet.
		return errors.Wrapf(dbError{
			error:     errors.Errorf("RowsAffected is 0"),
			name:      tableName,
			cols:      columns,
			statement: insertStmt,
		}, "trying insert no rows affected")
	}
	return nil
}

func getLatestDutyLogEntry(ctx context.Context, db *sql.DB) (DutyLogTable, error) {
	stmt := "SELECT DUTYSEQUENCE,MAX(DUTYDATE),CREW FROM DUTYLOG GROUP BY DUTYSEQUENCE,CREW"
	if rows, err := db.QueryContext(ctx, stmt); err != nil {
		return DutyLogTable{}, errors.Wrapf(dbError{
			error:     err,
			name:      "DUTYLOG",
			statement: stmt,
		}, "latest duty log entry running DB query")
	} else {
		defer rows.Close()
		entry := DutyLogTable{}
		for rows.Next() {
			var crewName sql.NullString
			if err := rows.Scan(
				&entry.DutyLog.ID,
				&entry.DutyLog.Date,
				&crewName,
			); err != nil {
				return DutyLogTable{}, errors.Wrapf(dbError{
					error:     err,
					name:      "DUTYLOG",
					statement: stmt,
				}, "latest duty log entry scanning table columns")
			} else {
				entry.DutyLog.CrewName = crewName.String
			}
		}
		return entry, nil
	}
}

func findMemberForEmail(ctx context.Context, db *sql.DB, email string) (Member, error) {
	stmt := "SELECT MEMBERNOLOCAL FROM MEMBERS WHERE LOWER(EMAILMRQ)=?"
	if rows, err := db.QueryContext(ctx, stmt, email); err != nil {
		return Member{}, errors.Wrapf(dbError{
			error:     err,
			name:      "MEMBERS",
			statement: stmt,
		}, "find member for email %s", email)
	} else {
		defer rows.Close()
		mbr := Member{}
		for rows.Next() {
			if err := rows.Scan(&mbr.ID); err != nil {
				return Member{}, errors.Wrapf(dbError{
					error:     err,
					name:      "MEMBERS",
					statement: stmt,
				}, "find member for email %s reading rows", email)
			}
		}
		return mbr, nil
	}
}

func findRankingForMember(ctx context.Context, db *sql.DB, id int) (int, error) {
	stmt := "SELECT FIRST 1 CREWRANKING FROM DUTYCREWS WHERE CREWMEMBER=?" +
		" ORDER BY DUTYSEQUENCE DESC"
	if rows, err := db.QueryContext(ctx, stmt, id); err != nil {
		return 0, errors.Wrapf(dbError{
			error:     err,
			name:      "DUTYCREWS",
			statement: stmt,
		}, "find ranking for member %d", id)
	} else {
		defer rows.Close()
		var rank int
		for rows.Next() {
			if err := rows.Scan(&rank); err != nil {
				return 0, errors.Wrapf(dbError{
					error:     err,
					name:      "DUTYCREWS",
					statement: stmt,
				}, "find ranking for member %d reading rows", id)
			}
		}
		return rank, nil
	}
}

func pullMemberRecordsByEmail(ctx context.Context, db *sql.DB, dutyCrewID int, email string) (crewInfo, error) {
	stmt := "SELECT M.MEMBERNOLOCAL,M.EMAILMRQ,C.DUTYSEQUENCE,C.CREWMEMBER,C.CREWRANKING" +
		" FROM MEMBERS M INNER JOIN DUTYCREWS C ON M.MEMBERNOLOCAL=C.CREWMEMBER" +
		" WHERE LOWER(M.EMAILMRQ)=? AND C.DUTYSEQUENCE=?"
	if rows, err := db.QueryContext(ctx, stmt, email, dutyCrewID); err != nil {
		return crewInfo{}, errors.Wrapf(dbError{
			error:     err,
			name:      "MEMBERS & DUTYCREWS",
			statement: stmt,
		}, "trying to fetch member records")
	} else {
		defer rows.Close()
		crew := crewInfo{}

		if !rows.Next() {
			return crewInfo{}, errors.Wrapf(dbZeroRowsErr,
				"member records for duty log %d don't contain member %s",
				dutyCrewID, email)
		} else {
			if err := rows.Scan(&crew.Member.ID, &crew.Member.Email2,
				&crew.CrewOnDuty.ID, &crew.CrewOnDuty.MemberNo, &crew.CrewOnDuty.RankID,
			); err != nil {
				return crewInfo{}, errors.Wrapf(dbError{
					error:     err,
					name:      "MEMBERS & DUTYCREWS",
					statement: stmt,
				}, "pulling data from row")
			} else {
				crew.Member.Email2 = strings.TrimSpace(crew.Member.Email2)
			}
		}
		return crew, nil
	}
}

func pullMembersOnJob(ctx context.Context, db *sql.DB, jobID int) ([]JobCrew, error) {
	if rows, err := db.QueryContext(ctx,
		"SELECT CREWDUTYSEQUENCE,CREWJOBSEQUENCE,CREWMEMBER,CREWRANKING,SKIPPER,EMAILMRQ FROM DUTYJOBSCREW"+
			" INNER JOIN MEMBERS ON CREWMEMBER=MEMBERNOLOCAL"+
			" WHERE CREWJOBSEQUENCE=?", jobID); err != nil {
		return []JobCrew{}, errors.Wrapf(err, "pullMembersOnJob for job ID %d", jobID)
	} else {
		defer rows.Close()
		members := make([]JobCrew, 0, 10)
		for i := 0; rows.Next(); i++ {
			crew := JobCrew{}
			if err := rows.Scan(
				&crew.DutyCrewID, &crew.JobID, &crew.MemberID,
				&crew.RankID, &crew.IsMaster,
				&crew.email,
			); err != nil {
				return []JobCrew{}, errors.Wrapf(err, "pullMembersOnJob DB row %d scan", i)
			}
			crew.IsMaster = CustomBool(strings.TrimSpace(string(crew.IsMaster)))
			crew.email = strings.TrimSpace(crew.email)
			members = append(members, crew)
		}
		return members, nil
	}
}

func (crew JobCrew) rmFromDB(ctx context.Context, db *sql.DB) error {
	stmt := "DELETE FROM DUTYJOBSCREW WHERE" +
		" CREWDUTYSEQUENCE=? AND CREWJOBSEQUENCE=? AND CREWMEMBER=?"
	if crew.DutyCrewID == 0 || crew.JobID == 0 || crew.MemberID == 0 {
		return errors.Errorf("IDs cannot be 0: %d, %d, %d",
			crew.DutyCrewID, crew.JobID, crew.MemberID)
	}
	if result, err := db.ExecContext(ctx, stmt,
		crew.DutyCrewID, crew.JobID, crew.MemberID,
	); err != nil {
		return errors.Wrapf(dbError{
			error:     err,
			name:      "DUTYJOBSCREW",
			statement: stmt,
		}, "rmMember DB exec")
	} else if rowCount, err := result.RowsAffected(); err != nil {
		return errors.Wrapf(dbError{
			error:     err,
			name:      "DUTYJOBSCREW",
			statement: stmt,
		}, "rmMember DB exec result")
	} else if rowCount != 1 {
		return errors.Wrapf(dbError{
			error:     err,
			name:      "DUTYJOBSCREW",
			statement: stmt,
		}, "rmMember rows deleted is %d", rowCount)
	}
	return nil
}

func getJobID(ctx context.Context, db *sql.DB, job Job) (int, error) {
	query := fmt.Sprintf("SELECT JOBJOBSEQUENCE FROM DUTYJOBS" +
		" WHERE JOBTIMEOUT=? AND JOBDUTYVESSELNAME=?")
	if rows, err := db.QueryContext(ctx, query, job.StartTime, job.VMRVessel.Name); err != nil {
		return 0, errors.Wrapf(err, "fetch job ID for time %s and vessel %s",
			job.StartTime, job.VMRVessel.Name)
	} else {
		defer rows.Close()
		if rows.Next() {
			if err := rows.Scan(&job.ID); err != nil {
				return 0, errors.Wrapf(err, "fetch job ID row scan")
			}
		}
		if rows.Next() {
			return job.ID, errors.Errorf("getJobID returned multiple rows")
		}
		return job.ID, nil
	}
}

// Add relevant crew to the crew table, linked to the job record
func addCrewForJob(ctx context.Context, db *sql.DB, job Job) error {
	const TBL = "DUTYJOBSCREW"
	addCrew := func(email string, isMaster bool) error {
		if crew, err := pullMemberRecordsByEmail(ctx, db, job.DutyLogID, email); err != nil &&
			errors.Is(err, dbZeroRowsErr) {
			// No row found, but this isn't considered an error so just return nil with no action.
			// We will silently ignore cases where the TripWatch member doesn't exist in Firebird.
			return nil
		} else if err != nil {
			return errors.Wrapf(err, "member records for user '%s'", email)
		} else {
			jc := JobCrew{
				DutyCrewID: crew.CrewOnDuty.ID,
				JobID:      job.ID,
				MemberID:   crew.CrewOnDuty.MemberNo,
				RankID:     crew.CrewOnDuty.RankID,
				IsMaster:   CustomBool("N"),
				IsOnJob:    CustomBool("Y"),
			}
			if isMaster {
				jc.IsMaster = CustomBool("Y")
			}
			columns := []column{}
			o := reflect.ValueOf(crewOnJob{JobCrew: jc})
			if err := forEachColumn("parent", o, func(tableName string, col column) error {
				if tableName == TBL {
					columns = append(columns, col)
				}
				return nil
			}); err != nil {
				return errors.Wrapf(err, "fetch col names for table %s", TBL)
			}
			var dberr dbError
			if err := tryUpdate(ctx, db, TBL, columns); err == nil {
				// This worked. Fall out of the statement chain
			} else if !errors.As(err, &dberr) {
				return errors.Wrapf(err, "tryUpdate returned a coding error")
			} else if err := tryInsert(ctx, db, TBL, columns); err != nil {
				return errors.Wrapf(err, "insert member records for job %d user '%s'",
					job.ID, email)
			}
		}
		return nil
	}
	if job.ID == 0 {
		if jobID, err := getJobID(ctx, db, job); err != nil {
			return errors.Wrapf(err, "addCrewForJob ID not found and cannot be 0")
		} else {
			job.ID = jobID
		}
	}
	// Add or update crew records for job
	for _, email := range job.VMRVessel.CrewList {
		if err := addCrew(email, false); err != nil {
			return errors.Wrapf(err, "addCrew for crew list")
		}
	}
	if job.VMRVessel.Master != "" {
		// A master has been designated - add or update them
		if err := addCrew(job.VMRVessel.Master, true); err != nil {
			return errors.Wrapf(err, "addCrew for master")
		}
	}
	// Check if any crew rows need to be removed & remove them if needed
	if members, err := pullMembersOnJob(ctx, db, job.ID); err != nil {
		return errors.Wrapf(err, "list all crew")
	} else if len(members) == len(job.VMRVessel.CrewList) {
		// Length of members list is equivalent, so no need to remove anything. Drop out of chain.
	} else {
		excessMembers := make([]JobCrew, 0, len(members))
		for _, member := range members {
			if member.email != job.VMRVessel.Master &&
				!job.VMRVessel.CrewList.Has(member.email) {
				excessMembers = append(excessMembers, member)
			}
		}
		for _, member := range excessMembers {
			if err := member.rmFromDB(ctx, db); err != nil {
				return errors.Wrapf(err, "rm excess crew: %s from job %d",
					member.email, member.JobID)
			}
		}
	}
	return nil
}

// For fields that are not automatically prefilled by data input from TripWatch, manually update
// them with aggregated data from other fields in the record.
func aggregateFields(data *linkActivationDB) error {
	if data == nil {
		return errors.Errorf("Data pointer cannot be nil")
	}

	data.Job.Emergency.Emergency = data.Job.Emergency.Notified
	if strings.HasSuffix(data.Job.AssistedVessel.Rego, "C") {
		data.Job.Commercial = "Y"
	} else {
		data.Job.Commercial = "N"
	}
	if data.Job.Weather.Forecast != "" {
		if err := parseForecast(&data.Job.Weather); err != nil {
			return errors.Wrapf(err, "aggregateFields parsing forecast failed")
		}
	}
	if !data.Job.Pos.IsZero() || len(data.Sitreps) > 0 {
		if err := setGPS(data); err != nil {
			return errors.Wrapf(err, "aggregateFields parsing latlong")
		}
	}

	if data.Job.AssistedVessel.Type != "" && data.Job.AssistedVessel.Propulsion != "" {
		if err := aggregatePropulsion(&data.Job.AssistedVessel); err != nil {
			return errors.Wrapf(err, "aggregateFields for vessel propulsion")
		}
	}

	// Automate actions and job source/frequency
	if (data.Job.Action == JobAction("") || data.Job.Action == JobAction("Other")) &&
		data.Job.Type != JobType("") {
		data.Job.Action = data.Job.Type.ToJobAction()
	}
	if data.Job.Type == JobType("Training/Patrol") {
		data.Job.ActivatedBy = JobSource("Base")
		data.Job.Freq = data.Job.ActivatedBy.ToJobFreq()
	} else if data.Job.Type == JobType("Medical") {
		data.Job.ActivatedBy = JobSource("QAS")
		data.Job.Freq = data.Job.ActivatedBy.ToJobFreq()
	}

	return nil
}

func aggregatePropulsion(vessel *AssistedVessel) error {
	return vessel.Propulsion.UpdateFromEngineQTY(vessel.EngineQTY)
}

func parseForecast(weather *Weather) error {
	if weather.Forecast == "" {
		return errors.Errorf("parseForecast string cannot be empty")
	}

	gcwaters := strings.Split(weather.Forecast, "Gold Coast Waters:")
	if len(gcwaters) < 2 {
		return errors.Errorf("parseForecast couldn't find GC forecast: %d", len(gcwaters))
	}
	for _, line := range strings.Split(gcwaters[1], "\n") {
		line = strings.ToLower(strings.TrimSpace(line))
		if strings.Contains(line, "winds:") {
			// Find mentions of wind directions
			if strings.Contains(line, "southeasterly") {
				weather.WindDir = WindDirEnum("SE")
			} else if strings.Contains(line, "southerly") {
				weather.WindDir = WindDirEnum("S")
			} else if strings.Contains(line, "southwesterly") {
				weather.WindDir = WindDirEnum("SW")
			} else if strings.Contains(line, "westerly") {
				weather.WindDir = WindDirEnum("W")
			} else if strings.Contains(line, "northwesterly") {
				weather.WindDir = WindDirEnum("NW")
			} else if strings.Contains(line, "northerly") {
				weather.WindDir = WindDirEnum("N")
			} else if strings.Contains(line, "northeasterly") {
				weather.WindDir = WindDirEnum("NE")
			} else if strings.Contains(line, "easterly") {
				weather.WindDir = WindDirEnum("E")
			}

			// Parse wind speed
			knotSplit := strings.Split(line, "knots")
			fieldsList := strings.Fields(knotSplit[0])
			speed := fieldsList[len(fieldsList)-1]
			if val, err := strconv.ParseInt(speed, 10, 32); err != nil {
				return errors.Wrapf(err, "parse wind speed %s", speed)
			} else {
				weather.WindSpeed.Set(int(val))
			}
		} else if strings.Contains(line, "weather:") {
			if strings.Contains(line, "sunny") || strings.Contains(line, "partly cloudy") {
				weather.RainState = "Clear"
			} else {
				weather.RainState = "Rain"
			}
		}
	}

	return nil
}

// Select a GPS value to set in Firebird by searching through the list of job Sitreps as well as
// the manually-entered GPS value. The job sitreps will be taken as precedence, and any sitrep
// where the RV arrives at a target is the main preference.
func setGPS(data *linkActivationDB) error {
	set := func(gps *FirebirdGPS, pos GPS) error {
		if dms, err := pos.AsDMS(); err != nil {
			return errors.Wrapf(err, "setGPS internal set func")
		} else {
			gps.Lat = pos.Lat
			gps.Long = pos.Long
			gps.LatD = dms.Lat.Deg
			gps.LatM = dms.Lat.Min
			gps.LatS = dms.Lat.Sec
			gps.LongD = dms.Long.Deg
			gps.LongM = dms.Long.Min
			gps.LongS = dms.Long.Sec
			return nil
		}
	}

	if sr, err := getEntryForComment(data.Sitreps, "RV has arrived at target"); err == nil {
		if err := set(&data.Job.FirebirdGPS, sr.Pos); err == nil {
			return nil
		}
	}
	if sr, err := getEntryForComment(data.Sitreps, "Target vessel in tow"); err == nil {
		if err := set(&data.Job.FirebirdGPS, sr.Pos); err == nil {
			return nil
		}
	}
	if len(data.Sitreps) > 0 {
		if err := set(&data.Job.FirebirdGPS, data.Sitreps[0].Pos); err == nil {
			return nil
		}
	}
	if err := set(&data.Job.FirebirdGPS, data.Job.Pos); err != nil {
		return errors.Wrapf(err, "parse GPS setting from overall job pos")
	}

	return nil
}

func sendToDB(ctx context.Context, db *sql.DB, data *linkActivationDB) error {
	// Aggregate any field entries that it is possible to aggregate
	if err := aggregateFields(data); err != nil {
		return errors.Wrapf(err, "sendToDB failed to aggregate fields")
	}

	// Map this to an existing DutyLog table entry
	if dl, err := getLatestDutyLogEntry(ctx, db); err != nil {
		return errors.Wrapf(err, "sendToDB failed to get duty log entry")
	} else {
		data.Job.DutyLogID = dl.DutyLog.ID
	}

	// Build a map of tables that contains the list of columns and associated data
	tables := make(map[string][]column)
	dbObj := reflect.ValueOf(*data)
	if err := forEachColumn("parent", dbObj, func(tableName string, col column) error {
		// Truncate string lengths if required
		if s, ok := col.value.(string); ok {
			if len(s) > col.maxStrlen {
				col.value = s[:col.maxStrlen]
			}
		}
		if !col.isSequence && reflect.ValueOf(col.value).IsZero() {
			if col.isMatch {
				return errors.Wrapf(matchFieldIsZero, "match field cannot be zero")
			} else {
				// Don't include values that are the zero-value for that type
				return nil
			}
		} else if _, ok := tables[tableName]; !ok {
			tables[tableName] = make([]column, 0)
		}
		tables[tableName] = append(tables[tableName], col)
		return nil
	}); err != nil {
		return errors.Wrapf(err, "build all insert statements column loop")
	}

	// For each table, synchronise the data with the firebird DB
	for table, columns := range tables {
		var dberr dbError
		// First try an SQL update statement, then if that fails try an SQL INSERT statement.
		if err := tryUpdate(ctx, db, table, columns); err == nil {
			// This worked. Move on to the next DB table
			continue
		} else if !errors.As(err, &dberr) {
			return errors.Wrapf(err, "tryUpdate returned a coding error")
		} else if inserr := tryInsert(ctx, db, table, columns); inserr != nil {
			return errors.Wrapf(inserr, "send to DB insert table %s", table)
		}
	}

	if err := addCrewForJob(ctx, db, data.Job); err != nil {
		return errors.Wrapf(err, "update job add crew rows")
	}

	return nil
}
