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
	return fmt.Sprintf("DB error \ntable: %s\nstatement: %s",
		e.name, e.statement)
}

type column struct {
	name       string
	isMatch    bool
	isSequence bool
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

// For fields that are not automatically prefilled by data input from TripWatch, manually update
// them with aggregated data from other fields in the record.
func aggregateFields(data *linkActivationDB) error {
	if data == nil {
		return errors.Errorf("Data pointer cannot be nil")
	}

	data.Job.Emergency.Emergency = bool(data.Job.Emergency.Notified)
	data.Job.Commercial = strings.HasSuffix(data.Job.AssistedVessel.Rego, "C")
	if data.Job.Weather.Forecast != "" {
		if err := parseForecast(&data.Job.Weather); err != nil {
			return errors.Wrapf(err, "aggregateFields parsing forecast failed")
		}
	}

	return nil
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

func sendToDB(ctx context.Context, db *sql.DB, data *linkActivationDB) error {
	// Aggregate any field entries that it is possible to aggregate
	if err := aggregateFields(data); err != nil {
		return errors.Wrapf(err, "sendToDB failed to aggregate fields")
	}

	// Build a map of tables that contains the list of columns and associated data
	tables := make(map[string][]column)
	dbObj := reflect.ValueOf(*data)
	if err := forEachColumn("parent", dbObj, func(tableName string, col column) error {
		if !col.isSequence && reflect.ValueOf(col.value).IsZero() {
			if col.isMatch {
				return errors.Errorf("match field cannot be zero")
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
			return errors.Wrapf(err, "send to DB insert table %s", table)
		}
	}

	return nil
}
