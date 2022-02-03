// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package closedloop

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// ClosedLoopConfig structs are all based on yaml configuration. The yaml
// field names are implied and do not need to be specified if they are
// merely the lowercased versions of the Golang field names.

// Action is an action to take when a rule matches
type Action struct {
	Operation string
	Field     *string
	Value     *uint32
}

// Rule is an expression that when matched will trigger actions
type Rule struct {
	Name        string
	Expr        *string
	Source      *string
	Destination *string
	Actions     []Action
	Debug       *bool
	Continue    *bool
}

// Source specifies the source of expressions
type Source struct {
	Name     string
	Endpoint string
}

// Destination specifies the destionation of actions
type Destination struct {
	Name     string
	Endpoint string
	Target   string
}

// Vcs is a virtual cellular service
type Vcs struct {
	Name  string
	Rules []Rule
}

// ClosedLoopConfig holds the configuration for the closed loop.
type ClosedLoopConfig struct { //nolint
	Sources      []Source
	Destinations []Destination
	Vcs          []Vcs
}

// LoadFromYamlFile loads a ClosedLoopConfig from a YAML File
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

// GetSourceByName given the name of a source, fetch it from the config tree
func (c *ClosedLoopConfig) GetSourceByName(name string) (*Source, error) {
	for _, src := range c.Sources {
		if src.Name == name {
			return &src, nil
		}
	}
	return nil, fmt.Errorf("Failed to find source %s", name)
}

// GetDestinationByName given the name of a destination, fetch it from the config tree
func (c *ClosedLoopConfig) GetDestinationByName(name string) (*Destination, error) {
	for _, dst := range c.Destinations {
		if dst.Name == name {
			return &dst, nil
		}
	}
	return nil, fmt.Errorf("Failed to find destination %s", name)
}
