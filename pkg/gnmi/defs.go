// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package gnmi implements a gnmi server to mock a device with YANG models.
package gnmi

import (
	"sync"

	"github.com/eapache/channels"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	pb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
)

// ConfigCallbackType is a type of configuration operation with enumerated values
type ConfigCallbackType int

const (
	// Initial is a callback for initial configuration
	Initial ConfigCallbackType = iota

	// Apply is a callback for an Apply
	Apply

	// Rollback is a callback for a Rollback
	Rollback

	// Forced is a callback when forced by the user
	Forced
)

func (c ConfigCallbackType) String() string {
	return [...]string{"Initial", "Apply", "Rollback"}[c]
}

// ConfigCallback is the signature of the function to apply a validated config to the physical device.
type ConfigCallback func(ygot.ValidatedGoStruct, ConfigCallbackType) error

var (
	pbRootPath         = &pb.Path{}
	supportedEncodings = []pb.Encoding{pb.Encoding_JSON, pb.Encoding_JSON_IETF}
	dataTypes          = []string{"config", "state", "operational", "all"}
)

// Server struct maintains the data structure for device config and implements the interface of gnmi server. It supports Capabilities, Get, and Set APIs.
// Typical usage:
//	g := grpc.NewServer()
//	s, err := Server.NewServer(model, config, callback)
//	pb.NewServer(g, s)
//	reflection.Register(g)
//	listen, err := net.Listen("tcp", ":8080")
//	g.Serve(listen)
//
// For a real device, apply the config changes to the hardware in the callback function.
// Arguments:
//		newConfig: new root config to be applied on the device.
// func callback(newConfig ygot.ValidatedGoStruct) error {
//		// Apply the config to your device and return nil if success. return error if fails.
//		//
//		// Do something ...
// }
type Server struct {
	model        *Model
	callback     ConfigCallback
	config       ygot.ValidatedGoStruct
	ConfigUpdate *channels.RingChannel
	mu           sync.RWMutex // mu is the RW lock to protect the access to config
	subscribed   map[string][]*streamClient
}

var (
	lowestSampleInterval uint64 = 5000000000 // 5000000000 nanoseconds
)

type streamClient struct {
	sr             *pb.SubscribeRequest
	stream         pb.GNMI_SubscribeServer
	UpdateChan     chan *pb.Update
	sampleInterval uint64
}

var log = logging.GetLogger("gnmi")
