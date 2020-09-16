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

package gnmi_target

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
