package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/nakagami/firebirdsql"
	"github.com/pkg/errors"
)

func openDB() (*sql.DB, error) {
	return sql.Open("firebirdsql", "SYSDBA:vmrdbpass@localhost:3050/firebird/data/VMRMEMBERS.FDB")
}

func openConfig() error {
	cfgFileName := os.Getenv("CONFIG_FILE")
	if cfgFileName == "" {
		cfgFileName = ".config.yml"
	}
	return errors.Wrapf(parseConfig(cfgFileName), "parse config in main")
}

func main() {
	conn, _ := openDB()
	defer conn.Close()
	if err := openConfig(); err != nil {
		log.Fatalf("Config parsing failed: %v", err)
	}
}
