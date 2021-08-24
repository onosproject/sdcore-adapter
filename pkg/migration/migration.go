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
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
)

var log = logging.GetLogger("migration")

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

// BuildStepList builds a list of migration steps that starts with "fromVersion" and ends with "toVersion".
// Steps must form a contiguous list of migrations -- each step must yield the migration that is used
// by the following migration.
func (m *Migrator) BuildStepList(fromVersion string, toVersion string) ([]*MigrationStep, error) {
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

// RunStep runs a migration step
func (m *Migrator) RunStep(step *MigrationStep, fromTarget string, toTarget string) ([]*MigrationActions, error) {
	// fetch the old models
	srcVal, err := GetPath(context.Background(), "", fromTarget, m.AetherConfigAddr)
	if err != nil {
		return nil, err
	}
	// TODO: handle srcVal == nil

	// fetch the new models
	destVal, err := GetPath(context.Background(), "", toTarget, m.AetherConfigAddr)
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

// ExecuteActions executes a list of actions
func (m *Migrator) ExecuteActions(actions []*MigrationActions, fromTarget string, toTarget string) error {
	// do the updates in forward order
	for _, action := range actions {
		err := Update(context.Background(), action.UpdatePrefix, toTarget, m.AetherConfigAddr, action.Updates)
		if err != nil {
			return err
		}
	}

	// now do the deletes in reverse order
	for i := len(actions) - 1; i >= 0; i-- {
		action := actions[i]
		err := Delete(context.Background(), action.DeletePrefix, fromTarget, m.AetherConfigAddr, action.Deletes)
		if err != nil {
			return err
		}
	}
	return nil
}

// Migrate performs migration from one version to another
func (m *Migrator) Migrate(fromTarget string, fromVersion string, toTarget string, toVersion string) error {
	steps, err := m.BuildStepList(fromVersion, toVersion)
	if err != nil {
		return err
	}

	for _, step := range steps {
		actions, err := m.RunStep(step, fromTarget, toTarget)
		if err != nil {
			return err
		}

		err = m.ExecuteActions(actions, fromTarget, toTarget)
		if err != nil {
			return err
		}
	}

	return nil
}

// NewMigrator creates a new migrator
func NewMigrator(aetherConfigAddr string) *Migrator {
	m := &Migrator{
		AetherConfigAddr: aetherConfigAddr,
	}
	return m
}
