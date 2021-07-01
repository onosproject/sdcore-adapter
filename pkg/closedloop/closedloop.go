// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package closedloop

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"

	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/metrics"
)

var log = logging.GetLogger("closedloop")

type Action struct {
	SetUpstream *uint64 `yaml:"set-upstream"`
}

type Rule struct {
	Name     string   `yaml:"name"`
	Expr     *string  `yaml:"expr"`
	Source   *string  `yaml:"source"`
	Actions  []Action `yaml:"actions"`
	Debug    *bool    `yaml:"debug"`
	Continue *bool    `yaml:"continue"`
}

type Source struct {
	Name     string
	Endpoint string
}

type Vcs struct {
	Name  string `yaml:"name"`
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

	if rule.Debug != nil && *rule.Debug {
		if result != nil {
			log.Infof("Debug: %s: %f", rule.Name, *result)
		} else {
			log.Infof("Debug: %s: <nil>", rule.Name)
		}
	}

	if rule.Continue != nil && *rule.Continue {
		return nil, nil
	}

	if result != nil {
		// Match!!
		return rule.Actions, nil
	}

	return nil, nil
}

func (c *ClosedLoopControl) EvaluateVcs(vcs *Vcs) error {
	for _, rule := range vcs.Rules {
		actions, err := c.EvaluateRule(&rule)
		if err != nil {
			return err
		}
		if actions != nil {
			// successful match, we're done.
			log.Infof("Vcs %s Rule %s matched", vcs.Name, rule.Name)
			return nil
		}
	}
	return nil
}

func (c *ClosedLoopControl) Evaluate() error {
	for _, vcs := range c.Config.Vcs {
		err := c.EvaluateVcs(&vcs)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewClosedLoopControl(config *ClosedLoopConfig) *ClosedLoopControl {
	clc := &ClosedLoopControl{Config: config}
	clc.Sources = map[string]*metrics.MetricsFetcher{}
	return clc
}
