// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizer implements the synchronizer.
package synchronizer

import (
	"time"

	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	"github.com/openconfig/ygot/ygot"
)

const (
	// DefaultImsiFormat is the default format for an IMSI string
	DefaultImsiFormat = "CCCNNNEEESSSSSS"
)

// Synchronizer is a Version 3 synchronizer.
type Synchronizer struct {
	outputFileName string
	postEnable     bool
	postTimeout    time.Duration
	pusher         PusherInterface
	updateChannel  chan *ConfigUpdate
	retryInterval  time.Duration

	// Busy indicator, primarily used for unit testing. The channel length in and of itself
	// is not sufficient, as it does not include the potential update that is currently syncing.
	// >0 if the synchronizer has operations pending and/or in-progress
	busy int32

	// used for ease of mocking
	synchronizeDeviceFunc func(config ygot.ValidatedGoStruct) (int, error)
}

// ConfigUpdate holds the configuration for a particular synchronization request
type ConfigUpdate struct {
	config       ygot.ValidatedGoStruct
	callbackType gnmi.ConfigCallbackType
}
