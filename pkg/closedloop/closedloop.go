// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package closedloop

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

/*
rules:
  - name: starbucks-low-utilization
    expr: sum(smf_pdu_session_profile{slice="starbucks_newyork_cameras",state="active"})<=3
    actions:
  	  - set-upstream: 300
  - name: starbucks-high-utilization
	expr: sum(smf_pdu_session_profile{slice="starbucks_newyork_cameras",state="active"})>5
	actions:
	  - set-upstream: 1000
  - name: starbucks-default
	actions:
	  - set-upstream: 500
*/

type Action struct {
	SetUpstream *uint64 `yaml:"set-upstream"`
}

type Rule struct {
	Name    string   `yaml:"name"`
	Expr    *string  `yaml:"expr"`
	Actions []Action `yaml:"actions"`
}

type ClosedLoopConfig struct {
	Rules []Rule `yaml:"rules"`
}

type ClosedLoopControl struct {
	Config *ClosedLoopConfig
}

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

func NewClosedLoopControl(config *ClosedLoopConfig) *ClosedLoopControl {
	return &ClosedLoopControl{Config: config}
}
