// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizerv3 implements the version 3 synchronizer.
package synchronizerv3

import (
	"time"

	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	"github.com/onosproject/sdcore-adapter/pkg/synchronizer"
	"github.com/openconfig/ygot/ygot"
)

const (
	DefaultImsiFormat = "CCCNNNEEESSSSSS"
)

type Synchronizer struct {
	outputFileName string
	postEnable     bool
	postTimeout    time.Duration
	pusher         synchronizer.PusherInterface
	updateChannel  chan *SynchronizerUpdate
	retryInterval  time.Duration

	// Busy indicator, primarily used for unit testing. The channel length in and of itself
	// is not sufficient, as it does not include the potential update that is currently syncing.
	// >0 if the synchronizer has operations pending and/or in-progress
	busy int32

	// used for ease of mocking
	synchronizeDeviceFunc func(config ygot.ValidatedGoStruct) error
}

type SynchronizerUpdate struct {
	config       ygot.ValidatedGoStruct
	callbackType gnmi.ConfigCallbackType
}
