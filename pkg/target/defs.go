// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	pb "github.com/openconfig/gnmi/proto/gnmi"
)

type target struct {
	*gnmi.Server
	Model       *gnmi.Model
	UpdateChann chan *pb.Update
}

var log = logging.GetLogger("target")
