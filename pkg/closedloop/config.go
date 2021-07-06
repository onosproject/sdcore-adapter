// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package closedloop

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// ClosedLoopConfig structs are all based on yaml configuration. The yaml
// field names are implied and do not need to be specified if they are
// merely the lowercased versions of the Golang field names.

type Action struct {
	Operation string
	Field     *string
	Value     *uint32
}

type Rule struct {
	Name        string
	Expr        *string
	Source      *string
	Destination *string
	Actions     []Action
	Debug       *bool
	Continue    *bool
}

type Source struct {
	Name     string
	Endpoint string
}

type Destination struct {
	Name     string
	Endpoint string
	Target   string
}

type Vcs struct {
	Name  string
	Rules []Rule
}

type ClosedLoopConfig struct {
	Sources      []Source
	Destinations []Destination
	Vcs          []Vcs
}

// Load a ClosedLoopConfig from a YAML File
func (c *ClosedLoopConfig) LoadFromYamlFile(fn string) error {
	yamlFile, err := ioutil.ReadFile(fn)
	if err != nil {
		return fmt.Errorf("Failed to read yaml file: %v", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal yaml: %v", err)
	}
	return nil
}

// Given the name of a source, fetch it from the config tree
func (c *ClosedLoopConfig) GetSourceByName(name string) (*Source, error) {
	for _, src := range c.Sources {
		if src.Name == name {
			return &src, nil
		}
	}
	return nil, fmt.Errorf("Failed to find source %s", name)
}

// Given the name of a destination, fetch it from the config tree
func (c *ClosedLoopConfig) GetDestinationByName(name string) (*Destination, error) {
	for _, dst := range c.Destinations {
		if dst.Name == name {
			return &dst, nil
		}
	}
	return nil, fmt.Errorf("Failed to find destination %s", name)
}
