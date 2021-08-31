// SPDX-FileCopyrightText: 2021-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package imsirange

import (
	"github.com/onosproject/onos-lib-go/pkg/logging"
)

// Imsi contains the range name and first and last imsi within the range
type Imsi struct {
	firstImsi uint64
	lastImsi  uint64
	name      string
}

//ImsiRange contain the Target and Address
type ImsiRange struct {
	AetherConfigAddress string
	AetherConfigTarget  string
}

var log = logging.GetLogger("imsirange")
