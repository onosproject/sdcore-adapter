// Copyright 2019-present Open Networking Foundation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
