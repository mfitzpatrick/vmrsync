// +build integration

package main

import "log"

func init() {
	// For integration tests, we need to read the configuration file
	if err := openConfig(); err != nil {
		log.Fatalf("Config parsing failed: %v", err)
	}
}
