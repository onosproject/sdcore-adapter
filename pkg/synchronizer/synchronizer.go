// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package gnmi implements a gnmi server to mock a device with YANG models.
package synchronizer

import (
	models "github.com/onosproject/config-models/modelplugin/aether-2.0.0/aether_2_0_0"
	modelplugin_v2 "github.com/onosproject/config-models/modelplugin/aether-2.0.0/modelplugin"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	"github.com/openconfig/ygot/ygot"
	"reflect"
	"time"
)

var log = logging.GetLogger("synchronizer")

func (s *Synchronizer) Synchronize(config ygot.ValidatedGoStruct) error {
	err := s.SynchronizeDevice(config)
	return err
}

func (s *Synchronizer) GetModels() *gnmi.Model {
	model := gnmi.NewModel(modelplugin_v2.ModelData,
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

func (s *Synchronizer) SetPostEnable(postEnable bool) {
	s.postEnable = postEnable
}

func (s *Synchronizer) SetPostTimeout(postTimeout time.Duration) {
	s.postTimeout = postTimeout
}

func (s *Synchronizer) Start() {
	log.Infof("Synchronizer starting (outputFileName=%s, postEnable=%s, postTimeout=%d)",
		s.outputFileName,
		s.postEnable,
		s.postTimeout)

	// TODO: Eventually we'll create a thread here that waits for config changes
}

func NewSynchronizer(outputFileName string, postEnable bool, postTimeout time.Duration) *Synchronizer {
	s := &Synchronizer{
		outputFileName: outputFileName,
		postEnable:     postEnable,
		postTimeout:    postTimeout,
	}
	return s
}
