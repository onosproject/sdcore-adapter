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

func (m *Migrator) AddMigrationStep(fromVersion string, fromModels *gnmi.Model, toVersion string, toModels *gnmi.Model) {
	step := MigrationStep{
		fromVersion: fromVersion,
		fromModels:  fromModels,
		toVersion:   toVersion,
		toModels:    toModels,
	}
	m.steps = append(m.steps, step)
}

func (m *Migrator) BuildStepList(fromVersion string, toVersion string) (*[]MigrationStep, error) {
	steps := []MigrationStep{}
	currentVersion := fromVersion

	if currentVersion == toVersion {
		//  there's nothing to do!
		return &steps, nil
	}

	for _, step := range m.steps {
		if currentVersion == step.toVersion {
			break
		}
		if step.fromVersion == currentVersion {
			steps = append(steps, step)
			currentVersion = step.toVersion
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
	err := GetPath("", fromTarget, m.aetherConfigAddr, context.Background())
	if err != nil {
		return err
	}
	_ = toTarget
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
		aetherConfigAddr: aetherConfigAddr,
	}
	return m
}
