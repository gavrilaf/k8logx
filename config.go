package main

import (
	"fmt"
	"io/ioutil"
	"regexp"

	"gopkg.in/yaml.v2"
)

type ContainerConfig struct {
	Pattern string     `yaml:"pattern"`
	Fields  [][]string `yaml:"fields-order,flow"`
}

type PodConfig struct {
	Pattern    string            `yaml:"pattern"`
	Containers []ContainerConfig `yaml:"containers,flow"`
}

type Config struct {
	Namespace     string            `yaml:"namespace"`
	SecondsBefore int               `yaml:"seconds-before"`
	Mapping       map[string]string `yaml:"mapping"`
	Pods          []PodConfig       `yaml:"pods,flow"`
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

func (c Config) GetPodConfig(podName string) *PodConfig {
	if len(c.Pods) == 0 { // empty pods list in config, means 'accept all'
		return &PodConfig{}
	}

	for _, p := range c.Pods {
		if p.isPodMatched(podName) {
			return &p
		}
	}

	return nil
}

// PodConfig

func (p PodConfig) isPodMatched(name string) bool {
	match, _ := regexp.MatchString(p.Pattern, name)
	return match
}

func (p PodConfig) GetContainerConfig(containerName string) *ContainerConfig {
	if len(p.Containers) == 0 { // empty containers list in config, means 'accept all'
		return &ContainerConfig{}
	}

	for _, c := range p.Containers {
		if match, _ := regexp.MatchString(c.Pattern, containerName); match {
			return &c
		}
	}

	return nil
}
