// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package gnmi implements a gnmi server to mock a device with YANG models.
package migration

import (
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
)

type MigrationFunction func(MigrationStep, string, *gpb.TypedValue, *gpb.TypedValue) error

type MigrationStep struct {
	FromVersion   string            // verion of source models
	FromModels    *gnmi.Model       // source models
	ToVersion     string            // version of destination models
	ToModels      *gnmi.Model       // destination models
	MigrationFunc MigrationFunction // function that executes the migration
	Migrator      *Migrator         // link to Migrator
}

type Migrator struct {
	steps            []MigrationStep
	AetherConfigAddr string
}
