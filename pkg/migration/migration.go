// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package gnmi implements a gnmi server to mock a device with YANG models.
package migration

import (
	"context"
	"fmt"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
)

var log = logging.GetLogger("migration")

func (m *Migrator) AddMigrationStep(fromVersion string, fromModels *gnmi.Model, toVersion string, toModels *gnmi.Model, migrationFunc MigrationFunction) {
	step := MigrationStep{
		FromVersion:   fromVersion,
		FromModels:    fromModels,
		ToVersion:     toVersion,
		ToModels:      toModels,
		MigrationFunc: migrationFunc,
		Migrator:      m,
	}
	m.steps = append(m.steps, step)
}

/*
 * BuildStepList
 *
 * Build a list of migration steps that starts with "fromVersion" and ends with "toVersion". Steps
 * must form a contiguous list of migrations -- each step must yield the migration that is used
 * by the following migration.
 */

func (m *Migrator) BuildStepList(fromVersion string, toVersion string) (*[]MigrationStep, error) {
	steps := []MigrationStep{}
	currentVersion := fromVersion

	if currentVersion == toVersion {
		//  there's nothing to do!
		return &steps, nil
	}

	for _, step := range m.steps {
		if currentVersion == step.ToVersion {
			break
		}
		if step.FromVersion == currentVersion {
			steps = append(steps, step)
			currentVersion = step.ToVersion
		}
	}

	if len(steps) == 0 {
		return nil, fmt.Errorf("Unable to find a step that started with version %s", fromVersion)
	}

	if currentVersion != toVersion {
		return nil, fmt.Errorf("Unable to find a step that ended with version %s", toVersion)
	}

	return &steps, nil
}

func (m *Migrator) RunStep(step MigrationStep, fromTarget string, toTarget string) error {
	srcVal, err := GetPath("", fromTarget, m.AetherConfigAddr, context.Background())
	if err != nil {
		return err
	}

	destVal, err := GetPath("", toTarget, m.AetherConfigAddr, context.Background())
	if err != nil {
		return err
	}

	err = step.MigrationFunc(step, toTarget, srcVal, destVal)
	if err != nil {
		return err
	}

	return nil
}

func (m *Migrator) Migrate(fromTarget string, fromVersion string, toTarget string, toVersion string) error {
	steps, err := m.BuildStepList(fromVersion, toVersion)
	if err != nil {
		return err
	}

	for _, step := range *steps {
		err := m.RunStep(step, fromTarget, toTarget)
		if err != nil {
			return err
		}
	}

	return nil
}

func NewMigrator(aetherConfigAddr string) *Migrator {
	m := &Migrator{
		AetherConfigAddr: aetherConfigAddr,
	}
	return m
}
