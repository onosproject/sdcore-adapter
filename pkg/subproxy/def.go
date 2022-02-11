// SPDX-FileCopyrightText: 2021-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package subproxy

import (
	"context"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/gnmiclient"
	"github.com/openconfig/ygot/ygot"
	"time"
)

var (
	log        = logging.GetLogger("subscriber-proxy")
	clientHTTP HTTPClient
)

type subscriberProxy struct {
	AetherConfigAddress string
	BaseWebConsoleURL   string
	AetherConfigTarget  string
	token               string
	gnmiClient          gnmiclient.GnmiInterface
	gnmiContext         context.Context
	PostTimeout         time.Duration
	retryInterval       time.Duration

	// used for ease of mocking
	synchronizeDeviceFunc func(config ygot.ValidatedGoStruct) error
}
