// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package modeldata contains the following model data in gnmi proto struct:
//	openconfig-interfaces 2.0.0,
//	openconfig-openflow 0.1.0,
//	openconfig-platform 0.5.0,
//	openconfig-system 0.2.0.
package modeldata

import (
	pb "github.com/openconfig/gnmi/proto/gnmi"
)

const (
	AetherSubscriberModel    = "aether-subscriber"
	AetherApnProfileModel    = "apn-profile"
	AetherQosProfileModel    = "qos-profile"
	AetherUpProfileModel     = "up-profile"
	AetherAccessProfileModel = "access-profile"
)

var (
	// ModelData is a list of supported models.
	ModelData = []*pb.ModelData{
		{
			Name:         AetherSubscriberModel,
			Organization: "Open Networking Foundation.",
			Version:      "2018-08-18",
		},
		{
			Name:         AetherQosProfileModel,
			Organization: "Open Networking Foundation.",
			Version:      "2018-08-18",
		},
		{
			Name:         AetherUpProfileModel,
			Organization: "Open Networking Foundation.",
			Version:      "2018-08-18",
		},
		{
			Name:         AetherAccessProfileModel,
			Organization: "Open Networking Foundation.",
			Version:      "2018-08-18",
		},
		{
			Name:         AetherApnProfileModel,
			Organization: "Open Networking Foundation.",
			Version:      "2018-08-18",
		}}
)
