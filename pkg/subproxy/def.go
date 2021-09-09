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

	// used for ease of mocking
	synchronizeDeviceFunc func(config ygot.ValidatedGoStruct) error
}
