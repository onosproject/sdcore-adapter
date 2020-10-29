// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package gnmi implements a gnmi server to mock a device with YANG models.
package migration

import (
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
)

type MigrationStep struct {
	fromVersion string
	fromModels  *gnmi.Model
	toVersion   string
	toModels    *gnmi.Model
}

type Migrator struct {
	steps []MigrationStep
}
