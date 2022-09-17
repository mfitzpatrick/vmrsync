//go:build linux || darwin || (windows && !service)

package main

import (
	"database/sql"
	"log"
	"time"

	"github.com/pkg/errors"
)

// Main run loop suitable for running on a system directly (and not as a Windows Service).
// Print statements are output directly to STDOUT in this mode.
func runLoop() {
	var db *sql.DB
	if fdb, closefunc, err := setup(); err != nil {
		log.Fatalf("Cannot connect to DB: %v", err)
	} else {
		defer closefunc()
		db = fdb
	}

	for {
		if errlist := run(db); len(errlist) > 0 {
			for _, err := range errlist {
				if errors.Is(err, matchFieldIsZero) {
					var runerr runError
					if ok := errors.As(err, &runerr); ok {
						log.Printf("Couldn't match field for %s",
							runerr.String())
					} else {
						log.Printf("Missing match field (and runError object)")
					}
				} else {
					log.Printf("Run loop failure: %+v", err)
				}
			}
		}
		time.Sleep(tripwatchPollFrequency)
	}
}
