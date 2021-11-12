// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizer is basic declarations and utilities for the synchronizer
package synchronizer

import (
	"time"

	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	pb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
)

// SynchronizerInterface defines the interface that all synchronizers should have.
type SynchronizerInterface interface { //nolint
	Synchronize(config ygot.ValidatedGoStruct, callbackType gnmi.ConfigCallbackType, path *pb.Path) error
	GetModels() *gnmi.Model
	SetOutputFileName(fileName string)
	SetPostEnable(postEnable bool)
	SetPostTimeout(postTimeout time.Duration)
	SetPusher(pusher PusherInterface)
	Start()
}

// PusherInterface is an interface to a pusher, which pushes json to underlying services.
//go:generate mockgen -destination=../test/mocks/mock_pusher.go -package=mocks github.com/onosproject/sdcore-adapter/pkg/synchronizer PusherInterface
type PusherInterface interface {
	PushUpdate(endpoint string, data []byte) error
	PushDelete(endpoint string) error
}
