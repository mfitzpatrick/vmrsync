package main

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

func parseConfig(fname string) error {
	cfg := struct {
		TripWatch struct {
			APIkey string `yaml:"apikey"`
			URL    string `yaml:"url"`
		} `yaml:"tripwatch"`
	}{}
	if file, err := os.Open(fname); err != nil {
		return errors.Wrapf(err, "parse config file opening")
	} else {
		defer file.Close()
		if err := yaml.NewDecoder(file).Decode(&cfg); err != nil {
			return errors.Wrapf(err, "parse config YAML unmarshalling")
		} else {
			// safe to set any global variables now
			tripwatchAPIkey = cfg.TripWatch.APIkey
			tripwatchURL = cfg.TripWatch.URL
		}
	}
	return nil
}
