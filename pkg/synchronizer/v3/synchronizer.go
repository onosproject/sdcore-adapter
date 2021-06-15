// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package gnmi implements a gnmi server to mock a device with YANG models.
package synchronizerv3

import (
	models "github.com/onosproject/config-models/modelplugin/aether-3.0.0/aether_3_0_0"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	gnmiproto "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
	"reflect"
	"time"
)

var log = logging.GetLogger("synchronizer")

/* ModelData: List of Models returned by gNMI Capabilities for this adapter.
 *
 * This data was formerly in the config-models repository, but is no longer authoritatively
 * stored there due to those models being moved to helm charts used with the onos operator.
 *
 * NOTE: It's unclear how useful this is -- the actual yang contents of the models are
 *   not returned by capabilities, only the names of the models.
 */
var ModelData = []*gnmiproto.ModelData{
	{Name: "connectivity-service", Organization: "Open Networking Foundation", Version: "2021-06-02"},
	{Name: "enterprise", Organization: "Open Networking Foundation", Version: "2021-06-02"},

	{Name: "aether-types", Organization: "Open Networking Foundation", Version: "2021-06-02"},
	{Name: "ap-list", Organization: "Open Networking Foundation", Version: "2021-06-02"},
	{Name: "application", Organization: "Open Networking Foundation", Version: "2021-06-02"},
	{Name: "device-group", Organization: "Open Networking Foundation", Version: "2021-06-02"},
	{Name: "device-model-list", Organization: "Open Networking Foundation", Version: "2021-06-02"},
	{Name: "ip-domain", Organization: "Open Networking Foundation", Version: "2021-06-02"},
	{Name: "network", Organization: "Open Networking Foundation", Version: "2021-06-02"},
	{Name: "site", Organization: "Open Networking Foundation", Version: "2021-06-02"},
	{Name: "upf", Organization: "Open Networking Foundation", Version: "2021-06-02"},
	{Name: "vcs", Organization: "Open Networking Foundation", Version: "2021-06-02"},
	{Name: "template", Organization: "Open Networking Foundation", Version: "2021-06-02"},
	{Name: "traffic-class", Organization: "Open Networking Foundation", Version: "2021-06-02"},
}

func (s *Synchronizer) Synchronize(config ygot.ValidatedGoStruct, callbackType gnmi.ConfigCallbackType) error {
	log.Infof("Synchronize, type=%s", callbackType)
	err := s.SynchronizeDevice(config)
	return err
}

func (s *Synchronizer) GetModels() *gnmi.Model {
	model := gnmi.NewModel(ModelData,
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
