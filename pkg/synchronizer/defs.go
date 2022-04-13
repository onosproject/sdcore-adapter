// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

// Package synchronizer implements the synchronizer.
package synchronizer

import (
	"time"

	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
)

const (
	// DefaultImsiFormat is the default format for an IMSI string
	DefaultImsiFormat = "CCCNNNEEESSSSSS"

	// DefaultPostTimeout is the default timeout for post operations
	DefaultPostTimeout = time.Second * 10

	// DefaultPartialUpdateEnable is the default partial update setting
	DefaultPartialUpdateEnable = true
)

// Synchronizer is a Version 3 synchronizer.
type Synchronizer struct {
	outputFileName      string
	postEnable          bool
	postTimeout         time.Duration
	pusher              PusherInterface
	updateChannel       chan *ConfigUpdate
	retryInterval       time.Duration
	partialUpdateEnable bool

	// Busy indicator, primarily used for unit testing. The channel length in and of itself
	// is not sufficient, as it does not include the potential update that is currently syncing.
	// >0 if the synchronizer has operations pending and/or in-progress
	busy int32

	// used for ease of mocking
	synchronizeDeviceFunc func(config gnmi.ConfigForest) (int, error)

	// cache of previously synchronized updates
	cache map[string]interface{}
}

// ConfigUpdate holds the configuration for a particular synchronization request
type ConfigUpdate struct {
	config       gnmi.ConfigForest
	callbackType gnmi.ConfigCallbackType
	target       string
}

// SynchronizerOption is for options passed when creating a new synchronizer
type SynchronizerOption func(c *Synchronizer) // nolint

// AetherScope is used within the synchronizer to convey the scope we're working at within the
// tree. Contexts were considered for this implementation, but rejected due to the lack of
// static checking.
type AetherScope struct {
	Enterprise   *RootDevice // Each enterprise is a configuration tree
	Generation   *string     // "4G" or "5G"
	CoreEndpoint *string
	Site         *Site
	Slice        *Slice
}
