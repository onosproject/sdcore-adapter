// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package closedloop

import (
	"context"
	"fmt"

	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/metrics"
	"github.com/onosproject/sdcore-adapter/pkg/migration"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
)

var log = logging.GetLogger("closedloop")

// CloseLoopControl object. Contains the configuration, as well as a
// list of sources and a cache of the last rules pplied.

type ClosedLoopControl struct {
	Config   *ClosedLoopConfig
	Sources  map[string]*metrics.MetricsFetcher
	LastRule map[string]string
}

// Retrieve a MetricsFetcher from the cached list of metrics fetcher. If it doesn't
// exist, then create a new one.
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

// Evaluate a rule. If the rule matches, return its set of actions.
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

// Evaluate the set of rules for a VCS, stopping at the first rule that matches. If a rule
// matches, then execute its actions.
func (c *ClosedLoopControl) EvaluateVcs(vcs *Vcs) error {
	for _, rule := range vcs.Rules {
		rule := rule // Linter: Make a shadow copy of the range variable
		actions, err := c.EvaluateRule(&rule)
		if err != nil {
			return err
		}
		if actions != nil {
			// successful match, we're done.
			log.Infof("Vcs %s Rule %s matched", vcs.Name, rule.Name)

			lastRule, okay := c.LastRule[vcs.Name]
			if okay && (lastRule == rule.Name) {
				// TODO: This assumes nobody manually changes the variable. Maybe eventually
				// we want to verify this and/or occasionally throw out the cache values.
				log.Infof("Rule %s is already applied", rule.Name)
				return nil
			}

			var destination *Destination
			if rule.Destination != nil {
				destination, err = c.Config.GetDestinationByName(*rule.Destination)
				if err != nil {
					return err
				}
			} else {
				destination, err = c.Config.GetDestinationByName("default")
				if err != nil {
					return err
				}
			}

			err := c.ExecuteActions(vcs, destination, rule.Actions)
			if err != nil {
				return err
			}

			c.LastRule[vcs.Name] = rule.Name

			return nil
		}
	}
	return nil
}

// Evaluate all rules for all VCSes
func (c *ClosedLoopControl) Evaluate() error {
	for _, vcs := range c.Config.Vcs {
		vcs := vcs // Linter: Make a shadow copy of range variable
		err := c.EvaluateVcs(&vcs)
		if err != nil {
			return err
		}
	}
	return nil
}

// Execute a list of actions
func (c *ClosedLoopControl) ExecuteActions(vcs *Vcs, destination *Destination, actions []Action) error {
	updates := []*gpb.Update{}
	for _, action := range actions {
		switch action.Operation {
		case "set":
			if action.Field == nil {
				return fmt.Errorf("Set action must contain a non-nil Field")
			}

			updates = migration.AddUpdate(updates, migration.UpdateUInt32(*action.Field, destination.Target, action.Value))
		default:
			return fmt.Errorf("Unknown action operation %s", action.Operation)
		}
	}

	prefixStr := fmt.Sprintf("vcs/vcs[id=%s]", vcs.Name)
	prefix := migration.StringToPath(prefixStr, destination.Target)

	log.Infof("Executing target=%s:%s, endpoint=%s, updates=%+v", destination.Target, prefixStr, destination.Endpoint, updates)

	err := migration.Update(prefix, destination.Target, destination.Endpoint, updates, context.Background())
	if err != nil {
		return fmt.Errorf("Error executing actions: %v", err)
	}

	return nil
}

// Create a new ClosedLoopControl.
func NewClosedLoopControl(config *ClosedLoopConfig) *ClosedLoopControl {
	clc := &ClosedLoopControl{Config: config}
	clc.Sources = map[string]*metrics.MetricsFetcher{}
	clc.LastRule = map[string]string{}
	return clc
}
