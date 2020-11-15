package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type PodConfig struct {
	Pattern   string     `yaml:"pattern"`
	Container string     `yaml:"container"`
	Fields    [][]string `yaml:"fields-order,flow"`
	ShowAll   bool       `yaml:"show-all"`
}

type Config struct {
	Pods []PodConfig `yaml:"pods,flow"`
}

func ReadConfig(name string) (Config, error) {
	bt, err := ioutil.ReadFile(name)
	if err != nil {
		return Config{}, fmt.Errorf("couldn't open config file, %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(bt, &cfg); err != nil {
		return Config{}, fmt.Errorf("invalid config file %s, %w", name, err)
	}

	return cfg, nil
}
