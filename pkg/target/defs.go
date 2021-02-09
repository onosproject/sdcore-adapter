// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package target

import (
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	pb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
)

type target struct {
	*gnmi.Server
	Model        *gnmi.Model
	configStruct ygot.ValidatedGoStruct
	UpdateChann  chan *pb.Update
}

var log = logging.GetLogger("target")
