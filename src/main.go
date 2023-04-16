package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"strings"
	"time"

	_ "github.com/nakagami/firebirdsql"
	"github.com/pkg/errors"
)

// When booting, initialise the 'last updated time' to the previous 12 hours. This
// is a means of ensuring we don't pull all old jobs across if the system is crashing
// for some significant time and then repaired.
var lastUpdatedTS time.Time = now().UTC().Add(-12 * time.Hour)
var now = time.Now
var configFilePath string
var printVersion bool
var Version string = "development"

func init() {
	flag.StringVar(&configFilePath, "config-file", ".config.yml", "Configuration YAML file")
	flag.BoolVar(&printVersion, "version", false, "Print version information and exit")
}

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

func setDBConnString(host string, port int, pass, path string) {
	dbConnStr = fmt.Sprintf("SYSDBA:%s@%s:%d/%s", pass, host, port, path)
}

func openDB() (*sql.DB, error) {
	return sql.Open("firebirdsql", dbConnStr)
}

func setup() (*sql.DB, func(), error) {
	if err := parseConfig(configFilePath); err != nil {
		return nil, nil, errors.Wrapf(err, "Config parsing failed")
	} else if db, err := openDB(); err != nil {
		return nil, nil, errors.Wrapf(err, "Unable to open DB")
	} else if err := db.Ping(); err != nil {
		return nil, nil, errors.Wrapf(err, "No connection to DB")
	} else {
		return db, func() { db.Close() }, nil
	}
}

// Primary execution cycle. This retrieves data from TripWatch and sends it to the Firebird DB.
func run(db *sql.DB) []error {
	var errlist []error
	// Shouldn't take more than 60s to perform the whole update (read and write)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if activations, err := listActivations(ctx, lastUpdatedTS.Add(-60*time.Second)); err != nil {
		errlist = append(errlist, errors.Wrapf(err, "List TripWatch activations"))
	} else {
		for i, activation := range activations {
			if strings.ToLower(activations[i].Job.Status) == "cancelled" {
				// Don't synchronise cancelled activations. Skip over them.
				continue
			} else if err := sendToDB(ctx, db, &activations[i]); err != nil {
				errlist = append(errlist, runError{
					error:      errors.Wrapf(err, "DB update for activation %d", activation.ID),
					activation: &activations[i],
				})
			}
		}
	}
	lastUpdatedTS = now().UTC()
	return errlist
}

func main() {
	flag.Parse()
	if printVersion {
		fmt.Println(Version)
		return
	}
	// Run an infinite loop reading data from TripWatch and synchronising it with the
	// Firebird DB.
	// NB: this function is conditionally linked due to tags issued at build time.
	runLoop()
}
