// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizerv2 implements the v2 synchronizer
package synchronizerv2

import (
	"time"
)

// Synchronizer is a version 2 Synchronizer.
type Synchronizer struct {
	outputFileName string
	postEnable     bool
	postTimeout    time.Duration
}
