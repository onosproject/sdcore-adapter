// SPDX-FileCopyrightText: 2021-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package subproxy

import (
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/gnmiclient"
	"github.com/openconfig/ygot/ygot"
	"time"
)

var (
	subscriberAPISuffix = "/api/subscriber/"
	log                 = logging.GetLogger("subscriber-proxy")
	clientHTTP          HTTPClient
)

type subscriberProxy struct {
	AetherConfigAddress string
	BaseWebConsoleURL   string
	AetherConfigTarget  string
	gnmiClient          gnmiclient.GnmiInterface
	PostTimeout         time.Duration
	retryInterval       time.Duration
	// Busy indicator, primarily used for unit testing. The channel length in and of itself
	// is not sufficient, as it does not include the potential update that is currently syncing.
	// >0 if the synchronizer has operations pending and/or in-progress
	busy int32

	// used for ease of mocking
	synchronizeDeviceFunc func(config ygot.ValidatedGoStruct) error
}
