// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package target

import (
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	pb "github.com/openconfig/gnmi/proto/gnmi"
)

// NewTarget creates a new target
func NewTarget(model *gnmi.Model, config []byte, callback gnmi.ConfigCallback) (*target, error) { //nolint
	s, err := gnmi.NewServer(model, config, callback)
	if err != nil {
		return nil, err
	}

	newconfig, _ := model.NewConfigStruct(config)
	channelUpdate := make(chan *pb.Update)
	target := target{Server: s,
		Model:        model,
		configStruct: newconfig,
		UpdateChann:  channelUpdate,
	}

	return &target, nil
}
