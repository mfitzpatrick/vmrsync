package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

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

// Recursive function which will generate a map of tables to a map of columns containing the type of the column.
func getFirebirdStructTags(tableName string, obj reflect.Type) (map[string]map[string]reflect.Type, error) {
	fields := make(map[string]reflect.Type)
	tmap := make(map[string]map[string]reflect.Type)
	for i := 0; i < obj.NumField(); i++ {
		structField := obj.Field(i)
		firebirdTag := structField.Tag.Get("firebird")
		if structField.Type.Kind() == reflect.Struct && structField.Type != reflect.TypeOf(time.Time{}) {
			nestedTable := tableName
			if firebirdTag != "" {
				nestedTable = firebirdTag
			}
			if nested, err := getFirebirdStructTags(nestedTable, structField.Type); err != nil {
				return tmap, errors.Wrapf(err, "firebird get struct tags")
			} else {
				for k, v := range nested {
					tmap[k] = v
				}
			}
		} else if firebirdTag == "" {
			continue //Nothing here matches with the firebird DB
		} else {
			fields[firebirdTag] = structField.Type
		}
	}
	if v, ok := tmap[tableName]; ok {
		// Extend field list
		for k, field := range v {
			fields[k] = field
		}
	}
	if len(fields) > 0 {
		tmap[tableName] = fields
	}
	return tmap, nil
}

func firebirdGet(db *linkActivationDB) error {
	mainObj := reflect.TypeOf(*db)
	if mainObj.Kind() != reflect.Struct {
		return errors.Errorf("db kind %v is not Struct (%v)", mainObj.Kind(), reflect.Struct)
	}
	if tableMap, err := getFirebirdStructTags("parent", mainObj); err != nil {
		return errors.Wrapf(err, "firebird get struct tags")
	} else {
		log.Printf("table map: %v", tableMap)
	}
	return nil
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
		if structVal.Kind() == reflect.Struct && structVal.Type() != reflect.TypeOf(time.Time{}) {
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
	colList := make([]string, len(columns))
	valList := make([]interface{}, len(columns))
	keyCol := []string{}
	keyVal := []interface{}{}
	i := 0
	for _, column := range columns {
		colList[i] = column.name
		valList[i] = column.value
		if column.isMatch {
			keyCol = append(keyCol, column.name)
			keyVal = append(keyVal, column.value)
		}
		i++
		if i > len(colList) || i > len(valList) {
			return errors.Errorf("Coding error: column list is smaller than column map")
		}
	}
	if len(keyCol) == 0 {
		return errors.Errorf("no keys specified for table %s", tableName)
	}
	if len(colList) == 0 {
		return errors.Errorf("no columns specified for table %s", tableName)
	}
	stmt := fmt.Sprintf("UPDATE %s SET %s WHERE %s", tableName,
		strings.Join(colList, "=?,")+"=?", strings.Join(keyCol, "=? AND")+"=?")
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
				cols:      columns,
				statement: idStmt,
			}, "insert getting next sequence number")
		} else if !rows.Next() {
			return 0, errors.Errorf("insert tx max id rows failed for table %s", tableName)
		} else if err := rows.Scan(&maxID); err != nil {
			return 0, errors.Errorf("insert tx max ID scan failed for table %s", tableName)
		}
		return maxID, nil
	}

	colList := make([]string, len(columns), len(columns)+1)
	valList := make([]interface{}, len(columns), len(columns)+1)
	keyCol := []string{}
	keyVal := []interface{}{}
	seqCol := ""
	i := 0
	for _, column := range columns {
		colList[i] = column.name
		valList[i] = column.value
		if column.isMatch {
			keyCol = append(keyCol, column.name)
			keyVal = append(keyVal, column.value)
		}
		if column.isSequence {
			seqCol = column.name
		}
		i++
		if i > len(colList) || i > len(valList) {
			return errors.Errorf("Coding error: column list is smaller than column map")
		}
	}
	if len(keyCol) == 0 {
		return errors.Errorf("no keys specified for table %s", tableName)
	}
	if len(colList) == 0 {
		return errors.Errorf("no columns specified for table %s", tableName)
	}
	if seqCol != "" {
		// Get the next logical sequence number for the table
		if maxID, err := getMaxID(ctx, tableName, seqCol); err != nil {
			return errors.Wrapf(err, "insert failed to get max sequence number")
		} else {
			colList = append(colList, seqCol)
			valList = append(valList, maxID+1)
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
			name:      tableName,
			cols:      columns,
			statement: insertStmt,
		}, "trying insert no rows affected")
	}
	return nil
}

func sendToDB(ctx context.Context, db *sql.DB, data *linkActivationDB) error {
	// Build a map of tables that contains the list of columns and associated data
	tables := make(map[string][]column)
	dbObj := reflect.ValueOf(*data)
	if err := forEachColumn("parent", dbObj, func(tableName string, col column) error {
		if reflect.ValueOf(col.value).IsZero() {
			if col.isMatch {
				return errors.Errorf("match field cannot be zero")
			} else {
				// Don't include values that are the zero-value for that type
				return nil
			}
		} else if _, ok := tables[tableName]; !ok {
			tables[tableName] = make([]column, 0, 12)
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
		} else if err := tryInsert(ctx, db, table, columns); err != nil {
			return errors.Wrapf(err, "send to DB insert table %s", table)
		}
	}

	return nil
}
