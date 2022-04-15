package main

import (
	"log"
	"reflect"
	"time"

	"github.com/pkg/errors"
)

// Recursive function which will generate a map of tables to a list of struct fields
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
