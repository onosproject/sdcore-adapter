// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package gnmi implements a gnmi server to mock a device with YANG models.
package gnmi

import (
	pb "github.com/openconfig/gnmi/proto/gnmi"
)

// NewServer creates an instance of Server with given json config.
func NewServer(model *Model, config []byte, callback ConfigCallback, synchronizer SynchronizerInterface) (*Server, error) {
	rootStruct, err := model.NewConfigStruct(config)
	if err != nil {
		return nil, err
	}
	s := &Server{
		model:        model,
		config:       rootStruct,
		callback:     callback,
		synchronizer: synchronizer,
	}
	if config != nil && s.callback != nil {
		if err := s.callback(rootStruct); err != nil {
			return nil, err
		}
	}
	// Initialize readOnlyUpdateValue variable

	val := &pb.TypedValue{
		Value: &pb.TypedValue_StringVal{
			StringVal: "INIT_STATE",
		},
	}
	s.readOnlyUpdateValue = &pb.Update{Path: nil, Val: val}
	s.subscribers = make(map[string]*streamClient)
	s.ConfigUpdate = make(chan *pb.Update, 100)

	return s, nil
}
