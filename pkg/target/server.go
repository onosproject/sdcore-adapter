// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package target

import (
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	pb "github.com/openconfig/gnmi/proto/gnmi"
)

func NewServer(model *gnmi.Model, config []byte, synchronizer gnmi.SynchronizerInterface) (*server, error) {
	s, err := gnmi.NewServer(model, config, nil, synchronizer)

	if err != nil {
		return nil, err
	}

	newconfig, _ := model.NewConfigStruct(config)
	channelUpdate := make(chan *pb.Update)
	server := server{Server: s, Model: model,
		configStruct: newconfig,
		UpdateChann:  channelUpdate,
	}

	return &server, nil
}
