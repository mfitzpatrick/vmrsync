package main

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

func parseConfig(fname string) error {
	cfg := struct {
		Firebird struct {
			Host     string `yaml:"host"`
			Port     int    `yaml:"port"`
			Password string `yaml:"password"`
			Path     string `yaml:"path"`
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
			setDBConnString(cfg.Firebird.Host, cfg.Firebird.Port, cfg.Firebird.Password,
				cfg.Firebird.Path)
		}
	}
	return nil
}
