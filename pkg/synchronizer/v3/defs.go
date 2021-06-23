// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package gnmi implements a gnmi server to mock a device with YANG models.
package synchronizerv3

import (
	"time"

	"github.com/onosproject/sdcore-adapter/pkg/synchronizer"
)

type Synchronizer struct {
	outputFileName string
	postEnable     bool
	postTimeout    time.Duration
	pusher         synchronizer.PusherInterface
}
