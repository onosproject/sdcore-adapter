// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Utility functions for synchronizer
package synchronizer

import (
        "time"

	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
        "github.com/openconfig/ygot/ygot"
)

type SynchronizerInterface interface {
     Synchronize(config ygot.ValidatedGoStruct, callbackType gnmi.ConfigCallbackType) error
     GetModels() *gnmi.Model
     SetOutputFileName(fileName string)
     SetPostEnable(postEnable bool)
     SetPostTimeout(postTimeout time.Duration)
     Start()
}
