// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Definitions of data structures and interfaces used by Migration.

package migration

import (
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
)

// MigrationFunction is the service-supplied function that migrates the models. It's provided with the
// definition of the step that's being executed, the from_target and to_target devices, and the
// set of models currently present on the from_target and to_target devices. It returns a set of
// actions to be executed in order.
type MigrationFunction func(*MigrationStep, string, string, *gpb.TypedValue, *gpb.TypedValue) ([]*MigrationActions, error) //nolint

// MigrationStep is the definition of a migration step. It defines which models are the "input"
// to the step and which models are output by the step. It provides a link to the function
// that executes the migration.
type MigrationStep struct { //nolint
	FromVersion   string            // verion of source models
	FromModels    *gnmi.Model       // source models
	ToVersion     string            // version of destination models
	ToModels      *gnmi.Model       // destination models
	MigrationFunc MigrationFunction // function that executes the migration
	Migrator      *Migrator         // link to Migrator
}

// MigrationActions is a set of actions that are returned by a MigrationFunction. There are
// two sets of actions, the "updates" that create the migrated object, and the deletes that
// remove the obsolete objects.
type MigrationActions struct { //nolint
	UpdatePrefix *gpb.Path
	Updates      []*gpb.Update
	DeletePrefix *gpb.Path
	Deletes      []*gpb.Path
}

// Migrator is the Migration Service.
type Migrator struct {
	steps            []*MigrationStep
	AetherConfigAddr string
}
