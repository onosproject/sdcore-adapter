// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package closedloop

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"

	"github.com/onosproject/sdcore-adapter/pkg/metrics"
)

type Action struct {
	SetUpstream *uint64 `yaml:"set-upstream"`
}

type Rule struct {
	Name    string   `yaml:"name"`
	Expr    *string  `yaml:"expr"`
	Source  *string  `yaml:"source"`
	Actions []Action `yaml:"actions"`
}

type Source struct {
	Name     string
	Endpoint string
}

type Vcs struct {
	Name  string
	Rules []Rule `yaml:"rules"`
}

type ClosedLoopConfig struct {
	Sources []Source `yaml:"sources"`
	Vcs     []Vcs    `yaml:"vcs"`
}

type ClosedLoopControl struct {
	Config  *ClosedLoopConfig
	Sources map[string]*metrics.MetricsFetcher
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

func (c *ClosedLoopConfig) GetSourceByName(name string) (*Source, error) {
	for _, src := range c.Sources {
		if src.Name == name {
			return &src, nil
		}
	}
	return nil, fmt.Errorf("Failed to find source %s", name)
}

func (c *ClosedLoopControl) GetFetcher(endpoint string) (*metrics.MetricsFetcher, error) {
	mf, okay := c.Sources[endpoint]
	if okay {
		return mf, nil
	}

	mf, err := metrics.NewMetricsFetcher(endpoint)
	if err != nil {
		return nil, err
	}
	c.Sources[endpoint] = mf
	return mf, nil
}

func (c *ClosedLoopControl) EvaluateRule(rule *Rule) ([]Action, error) {
	var err error
	var source *Source

	if rule.Expr == nil {
		// no expression; the rule always evaluates as True
		return rule.Actions, nil
	}

	if rule.Source != nil {
		source, err = c.Config.GetSourceByName(*rule.Source)
		if err != nil {
			return nil, err
		}
	} else {
		source, err = c.Config.GetSourceByName("default")
		if err != nil {
			return nil, err
		}
	}

	fetcher, err := c.GetFetcher(source.Endpoint)
	if err != nil {
		return nil, err
	}

	result, err := fetcher.GetSingleVector(*rule.Expr)
	if err != nil {
		return nil, err
	}

	if result != nil {
		// Match!!
		return rule.Actions, nil
	}

	return nil, nil
}

func NewClosedLoopControl(config *ClosedLoopConfig) *ClosedLoopControl {
	clc := &ClosedLoopControl{Config: config}
	clc.Sources = map[string]*metrics.MetricsFetcher{}
	return clc
}
