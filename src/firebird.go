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

type firebirdColHandler func(tableName, colName string, isKey bool, item interface{}) error

// Recursive function which will call the handler function for each item in the struct.
func forEachColumn(tableName string, obj reflect.Value, handler firebirdColHandler) error {
	if obj.Type().Kind() != reflect.Struct {
		return errors.Errorf("obj should be a struct, not %v", obj.Type().Kind())
	}
	for i := 0; i < obj.NumField(); i++ {
		structVal := obj.Field(i)
		structField := obj.Type().Field(i)
		isKey := false
		firebirdTag := structField.Tag.Get("firebird")
		if fbTags := strings.SplitN(firebirdTag, ",", 2); len(fbTags) > 1 && fbTags[1] == "match" {
			isKey = true
			firebirdTag = fbTags[0]
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
		} else if err := handler(tableName, firebirdTag, isKey, structVal.Interface()); err != nil {
			return errors.Wrapf(err, "firebird for each column (%s.%s) handler", tableName, firebirdTag)
		}
	}
	return nil
}

func buildInsertStatements(db *linkActivationDB) (map[string][]interface{}, error) {
	type tableData struct {
		isKey bool
		arg   interface{}
	}
	tables := make(map[string]map[string]tableData)
	dbObj := reflect.ValueOf(*db)
	if err := forEachColumn("parent", dbObj, func(tableName, colName string, isKey bool, item interface{}) error {
		if reflect.ValueOf(item).IsZero() {
			if isKey {
				return errors.Errorf("Key field cannot be zero")
			} else {
				// Don't include values that are the zero-value for that type
				return nil
			}
		} else if _, ok := tables[tableName]; !ok {
			tables[tableName] = make(map[string]tableData)
		}
		tables[tableName][colName] = tableData{
			isKey: isKey,
			arg:   item,
		}
		return nil
	}); err != nil {
		return make(map[string][]interface{}), errors.Wrapf(err, "build all insert statements column loop")
	}

	statements := make(map[string][]interface{}, len(tables))
	for table, columns := range tables {
		colList := make([]string, len(columns))
		valList := make([]interface{}, len(columns))
		valPlaceholder := make([]string, len(columns))
		valKeys := []string{}
		i := 0
		for column, value := range columns {
			colList[i] = column
			valList[i] = value.arg
			if value.isKey {
				valKeys = append(valKeys, column)
			}
			valPlaceholder[i] = fmt.Sprintf("?")
			i++
			if i > len(colList) || i > len(valList) {
				return make(map[string][]interface{}),
					errors.Errorf("Coding error: column list is smaller than column map")
			}
		}
		if len(valKeys) == 0 {
			return make(map[string][]interface{}), errors.Errorf("no keys specified for table %s", table)
		}
		stmt := fmt.Sprintf("UPDATE OR INSERT INTO %s (%s) VALUES (%s) MATCHING (%s)", table,
			strings.Join(colList, ","), strings.Join(valPlaceholder, ","),
			strings.Join(valKeys, ","))
		statements[stmt] = valList
	}

	return statements, nil
}

func runStatements(ctx context.Context, db *sql.DB, statements map[string][]interface{}) error {
	// Helper function to wrap an error and rollback the current SQL transaction
	rollbackWrapf := func(tx *sql.Tx, err error, format string, args ...interface{}) error {
		if rbErr := tx.Rollback(); rbErr != nil {
			rbstr := fmt.Sprintf("rollback error (%s)", err.Error())
			return errors.Wrapf(err, rbstr+format, args...)
		}
		return errors.Wrapf(err, format, args...)
	}

	if txn, err := db.BeginTx(ctx, &sql.TxOptions{}); err != nil {
		return errors.Wrapf(err, "run DB statements txn begin")
	} else {
		for stmt, args := range statements {
			if result, err := txn.ExecContext(ctx, stmt, args...); err != nil {
				return rollbackWrapf(txn, err, "run DB statement: '%s'", stmt)
			} else if rowCount, err := result.RowsAffected(); err != nil {
				return rollbackWrapf(txn, err, "run DB statements get result rows")
			} else if rowCount != 1 {
				return rollbackWrapf(txn, errors.Errorf("run DB statements should update or insert 1 row"), "")
			}
		}
		if txn.Commit(); err != nil {
			return rollbackWrapf(txn, err, "commit DB statements")
		}
	}
	return nil
}
