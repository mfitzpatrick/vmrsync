package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/nakagami/firebirdsql"
	"github.com/pkg/errors"
)

var dbConnStr string

func setDBConnString(host string, port int, pass, path string) {
	dbConnStr = fmt.Sprintf("SYSDBA:%s@%s:%d/%s", pass, host, port, path)
}

func openDB() (*sql.DB, error) {
	return sql.Open("firebirdsql", dbConnStr)
}

func openConfig() error {
	cfgFileName := os.Getenv("CONFIG_FILE")
	if cfgFileName == "" {
		cfgFileName = ".config.yml"
	}
	return errors.Wrapf(parseConfig(cfgFileName), "parse config in main")
}

func autoGenerateMRQEmails(db *sql.DB) error {
	// Create new column for MRQ email address, if such a column doesn't already exist.
	db.ExecContext(context.Background(),
		"ALTER TABLE MEMBERS ADD EMAILMRQ CHAR(96)",
	)

	// Read all existing active member records which don't have an MRQ email already set.
	if rows, err := db.QueryContext(context.Background(),
		"SELECT MEMBERNOLOCAL,SURNAME,FIRSTNAME,EMAIL1,EMAIL2 FROM MEMBERS"+
			" WHERE CURRENTCREW IS NOT NULL AND EMAILMRQ IS NULL",
	); err != nil {
		return errors.Wrapf(err, "Failed to read active members")
	} else {
		defer rows.Close()

		for rows.Next() {
			var id int
			var surname, firstname sql.NullString
			var email1, email2 sql.NullString
			if err := rows.Scan(&id, &surname, &firstname, &email1, &email2); err != nil {
				return errors.Wrapf(err, "Read failure for member")
			} else {
				const MRQ_DOMAIN string = "mrq.org.au"
				var email sql.NullString
				// Formulate an MRQ email address from the first and surname entry.
				// First, see if an MRQ email address is already in use. If so, use that.
				// Otherwise formulate our own.
				if email1.Valid && strings.Contains(email1.String, MRQ_DOMAIN) {
					email.Valid = true
					email.String = strings.TrimSpace(strings.ToLower(email1.String))
				} else if email2.Valid && strings.Contains(email2.String, MRQ_DOMAIN) {
					email.Valid = true
					email.String = strings.TrimSpace(strings.ToLower(email2.String))
				} else if surname.Valid && !firstname.Valid {
					email.Valid = true
					email.String = fmt.Sprintf("%s@%s",
						strings.TrimSpace(strings.ToLower(
							strings.ReplaceAll(surname.String, " ", ""))),
						MRQ_DOMAIN)
				} else if surname.Valid && firstname.Valid {
					email.Valid = true
					email.String = fmt.Sprintf("%s.%s@%s",
						strings.TrimSpace(strings.ToLower(
							strings.ReplaceAll(firstname.String, " ", ""))),
						strings.TrimSpace(strings.ToLower(
							strings.ReplaceAll(surname.String, " ", ""))),
						MRQ_DOMAIN)
				}
				// Perform DB update
				if _, err := db.ExecContext(context.Background(),
					"UPDATE MEMBERS SET EMAILMRQ=? WHERE MEMBERNOLOCAL=?",
					email, id,
				); err != nil {
					return errors.Wrapf(err, "Email-write of %s failure for member %d",
						email.String, id)
				}
			}
		}
	}

	return nil
}

// Update just 1 member's email address using the member's local ID number as a table key.
func updateSingleEmail(db *sql.DB, userID int, email string) error {
	if _, err := db.ExecContext(context.Background(),
		"UPDATE MEMBERS SET EMAILMRQ=? WHERE MEMBERNOLOCAL=?",
		email, userID,
	); err != nil {
		return errors.Wrapf(err, "update single email for user %d to %s failed: %v",
			userID, email, err)
	}

	return nil
}

func main() {
	argEmail := flag.String("email", "", "Email address to set on member record")
	argUserID := flag.Int("id", 0, "User ID to set email address field for")
	flag.Parse()

	if err := openConfig(); err != nil {
		log.Fatalf("Config parsing failed: %v", err)
	} else if db, err := openDB(); err != nil {
		log.Fatalf("Unable to open DB: %v", err)
	} else if err := db.Ping(); err != nil {
		log.Fatalf("No connection to DB: %v", err)
	} else {
		defer db.Close()

		if argEmail != nil && *argEmail != "" && argUserID != nil && *argUserID != 0 {
			if err := updateSingleEmail(db, *argUserID, *argEmail); err != nil {
				log.Printf("updating single email for user %d to %s: %v",
					*argUserID, *argEmail, err)
			}
		} else if err := autoGenerateMRQEmails(db); err != nil {
			log.Printf("auto-generating all emails: %v", err)
		}
	}
}
