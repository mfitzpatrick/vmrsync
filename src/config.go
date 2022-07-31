package main

import (
	"os"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

func parseConfig(fname string) error {
	cfg := struct {
		TripWatch struct {
			APIkey        string `yaml:"apikey"`
			URL           string `yaml:"url"`
			PollFrequency string `yaml:"poll"`
		} `yaml:"tripwatch"`
		Firebird struct {
			Host     string `yaml:"host"`
			Port     int    `yaml:"port"`
			Password string `yaml:"password"`
		} `yaml:"firebird"`
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
			if freq, err := time.ParseDuration(cfg.TripWatch.PollFrequency); err != nil {
				return errors.Wrapf(err, "parse config duration")
			} else {
				tripwatchPollFrequency = freq
			}
			setDBConnString(cfg.Firebird.Host, cfg.Firebird.Port, cfg.Firebird.Password)
		}
	}
	return nil
}
