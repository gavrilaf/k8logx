package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

type ContainerConfig struct {
	Pattern string     `yaml:"pattern"`
	Fields  [][]string `yaml:"fields-order,flow"`
	ShowAll bool       `yaml:"show-all"`
}

type PodConfig struct {
	Pattern    string            `yaml:"pattern"`
	Containers []ContainerConfig `yaml:"containers,flow"`
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

func (p PodConfig) IsPodMatched(name string) bool {
	return strings.HasPrefix(name, p.Pattern)
}

func (p PodConfig) GetContainerConfig(name string) *ContainerConfig {
	for _, c := range p.Containers {
		if strings.HasPrefix(name, c.Pattern) {
			return &c
		}
	}

	return nil
}
