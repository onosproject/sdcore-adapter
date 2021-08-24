// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizer is basic declarations and utilities for the synchronizer
package synchronizer

import (
	"time"

	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	"github.com/openconfig/ygot/ygot"
)

// SynchronizerInterface defines the interface that all synchronizers should have.
type SynchronizerInterface interface { //nolint
	Synchronize(config ygot.ValidatedGoStruct, callbackType gnmi.ConfigCallbackType) error
	GetModels() *gnmi.Model
	SetOutputFileName(fileName string)
	SetPostEnable(postEnable bool)
	SetPostTimeout(postTimeout time.Duration)
	SetPusher(pusher PusherInterface)
	Start()
}

// PusherInterface is an interface to a pusher, which pushes json to underlying services.
type PusherInterface interface {
	PushUpdate(endpoint string, data []byte) error
}
