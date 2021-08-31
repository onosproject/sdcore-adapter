// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

/*
 * The main entry point for the migration engine. Provides functions to create new migrators
 * and steps, and to execute migration.
 */

package migration

import (
	"context"
	"fmt"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	"github.com/onosproject/sdcore-adapter/pkg/gnmiclient"
)

// AddMigrationStep adds a migration step to the list
func (m *Migrator) AddMigrationStep(fromVersion string, fromModels *gnmi.Model, toVersion string, toModels *gnmi.Model, migrationFunc MigrationFunction) {
	step := MigrationStep{
		FromVersion:   fromVersion,
		FromModels:    fromModels,
		ToVersion:     toVersion,
		ToModels:      toModels,
		MigrationFunc: migrationFunc,
		Migrator:      m,
	}
	m.steps = append(m.steps, &step)
}

// buildStepList builds a list of migration steps that starts with "fromVersion" and ends with "toVersion".
// Steps must form a contiguous list of migrations -- each step must yield the migration that is used
// by the following migration.
func (m *Migrator) buildStepList(fromVersion string, toVersion string) ([]*MigrationStep, error) {
	steps := []*MigrationStep{}
	currentVersion := fromVersion

	if currentVersion == toVersion {
		//  there's nothing to do!
		return steps, nil
	}

	for _, step := range m.steps {
		if step.FromVersion == currentVersion {
			steps = append(steps, step)
			currentVersion = step.ToVersion
		}
		if currentVersion == toVersion {
			break
		}
	}

	if len(steps) == 0 {
		return nil, fmt.Errorf("Unable to find a step that started with version %s", fromVersion)
	}

	if currentVersion != toVersion {
		return nil, fmt.Errorf("Unable to find a step that ended with version %s", toVersion)
	}

	return steps, nil
}

// runStep runs a migration step
func (m *Migrator) runStep(step *MigrationStep, fromTarget string, toTarget string) ([]*MigrationActions, error) {
	// fetch the old models
	srcVal, err := m.Gnmi.GetPath(context.Background(), "", fromTarget, m.Gnmi.Address())
	if err != nil {
		return nil, err
	}
	// TODO: handle srcVal == nil

	// fetch the new models
	destVal, err := m.Gnmi.GetPath(context.Background(), "", toTarget, m.Gnmi.Address())
	if err != nil {
		return nil, err
	}
	// TODO: handle destVal == nil

	// execute the function to migrate items from old to new
	actions, err := step.MigrationFunc(step, fromTarget, toTarget, srcVal, destVal)
	if err != nil {
		return nil, err
	}

	return actions, nil
}

// executeActions executes a list of actions
func (m *Migrator) executeActions(actions []*MigrationActions, fromTarget string, toTarget string) error {
	// do the updates in forward order
	for _, action := range actions {
		if len(action.Updates) == 0 {
			if len(action.Deletes) == 0 {
				return errors.NewInternal("action has neither Updates nor Deletes",
					action.UpdatePrefix.String(), action.DeletePrefix.String())
			}
			continue
		}
		err := m.Gnmi.Update(context.Background(), action.UpdatePrefix, toTarget, m.Gnmi.Address(), action.Updates)
		if err != nil {
			return err
		}
	}

	// now do the deletes in reverse order
	for i := len(actions) - 1; i >= 0; i-- {
		action := actions[i]
		if len(action.Deletes) == 0 {
			continue
		}
		err := m.Gnmi.Delete(context.Background(), action.DeletePrefix, fromTarget, m.Gnmi.Address(), action.Deletes)
		if err != nil {
			return err
		}
	}
	return nil
}

// Migrate performs migration from one version to another
func (m *Migrator) Migrate(fromTarget string, fromVersion string, toTarget string, toVersion string) error {
	steps, err := m.buildStepList(fromVersion, toVersion)
	if err != nil {
		return err
	}

	for _, step := range steps {
		actions, err := m.runStep(step, fromTarget, toTarget)
		if err != nil {
			return err
		}

		err = m.executeActions(actions, fromTarget, toTarget)
		if err != nil {
			return err
		}
	}

	return nil
}

// NewMigrator creates a new migrator
func NewMigrator(gnmiClient gnmiclient.GnmiInterface) *Migrator {
	m := &Migrator{
		AetherConfigAddr: "",
		Gnmi:             gnmiClient,
	}

	return m
}

// SupportedVersions - list the versions supported
func (m *Migrator) SupportedVersions() []string {
	stepsSupported := make([]string, 0)
	for _, s := range m.steps {
		stepsSupported = append(stepsSupported, fmt.Sprintf("%s to %s", s.FromVersion, s.ToVersion))
	}
	return stepsSupported
}
