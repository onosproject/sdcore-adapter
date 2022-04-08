// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	pb "github.com/openconfig/gnmi/proto/gnmi"
)

// NewTarget creates a new target
func NewTarget(model *gnmi.Model, callback gnmi.ConfigCallback) (*target, error) { //nolint
	s, err := gnmi.NewServer(model, callback)
	if err != nil {
		return nil, err
	}

	channelUpdate := make(chan *pb.Update)
	target := target{Server: s,
		Model:       model,
		UpdateChann: channelUpdate,
	}

	return &target, nil
}
