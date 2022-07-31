package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/nakagami/firebirdsql"
	"github.com/pkg/errors"
)

// Error type returned by the run() function in main.go
type runError struct {
	error
	activation *linkActivationDB
}

func (e runError) String() string {
	if e.activation == nil {
		return "Empty RunError"
	}
	return fmt.Sprintf("activation %d on %s at %s",
		e.activation.ID, e.activation.Job.VMRVessel.Name, e.activation.Job.StartTime)
}

func (e runError) Error() string {
	if e.activation == nil {
		return e.error.Error()
	}
	return fmt.Sprintf("%s: %s", e.String(), e.error.Error())
}

func (e runError) Unwrap() error {
	return e.error
}

var dbConnStr string

func setDBConnString(host string, port int, pass string) {
	dbConnStr = fmt.Sprintf("SYSDBA:%s@%s:%d", pass, host, port)
}

func openDB() (*sql.DB, error) {
	return sql.Open("firebirdsql", fmt.Sprintf("%s/firebird/data/VMRMEMBERS.FDB", dbConnStr))
}

func openConfig() error {
	cfgFileName := os.Getenv("CONFIG_FILE")
	if cfgFileName == "" {
		cfgFileName = ".config.yml"
	}
	return errors.Wrapf(parseConfig(cfgFileName), "parse config in main")
}

func run(db *sql.DB) []error {
	var errlist []error
	// Shouldn't take more than 60s to perform the whole update (read and write)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if activations, err := listActivations(ctx); err != nil {
		errlist = append(errlist, errors.Wrapf(err, "List TripWatch activations"))
	} else {
		for i, _ := range activations {
			if err := sendToDB(ctx, db, &activations[i]); err != nil {
				errlist = append(errlist, runError{
					error:      errors.Wrapf(err, "DB update"),
					activation: &activations[i],
				})
			}
		}
	}
	return errlist
}

func main() {
	if err := openConfig(); err != nil {
		log.Fatalf("Config parsing failed: %v", err)
	} else if db, err := openDB(); err != nil {
		log.Fatalf("Unable to open DB: %v", err)
	} else if err := db.Ping(); err != nil {
		log.Fatalf("No connection to DB: %v", err)
	} else {
		defer db.Close()

		// Run an infinite loop reading data from TripWatch and synchronising it with the
		// Firebird DB.
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
}
