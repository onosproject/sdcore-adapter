// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package gnmi implements a gnmi server to mock a device with YANG models.
package synchronizer

import (
	models "github.com/onosproject/config-models/modelplugin/aether-1.0.0/aether_1_0_0"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	"github.com/onosproject/sdcore-adapter/pkg/synchronizer/modeldata"
	"github.com/openconfig/ygot/ygot"
	"reflect"
)

var log = logging.GetLogger("synchronizer")

func (s *Synchronizer) Synchronize(config ygot.ValidatedGoStruct) error {
	err := s.SynchronizeSpgw(config)
	return err
}

func (s *Synchronizer) GetModels() *gnmi.Model {
	model := gnmi.NewModel(modeldata.ModelData,
		reflect.TypeOf((*models.Device)(nil)),
		models.SchemaTree["Device"],
		models.Unmarshal,
		//models.ΛEnum  // NOTE: There is no Enum in the aether models? So use a blank map.
		map[string]map[int64]ygot.EnumDefinition{},
	)

	return model
}

func (s *Synchronizer) SetOutputFileName(fileName string) {
	s.outputFileName = fileName
}

func NewSynchronizer(outputFileName string) *Synchronizer {
	s := &Synchronizer{outputFileName: outputFileName}
	return s
}
