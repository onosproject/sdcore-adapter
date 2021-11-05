// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizerv3 implements a synchronizer for Aether v3 models
package synchronizerv3

import (
	models "github.com/onosproject/config-models/modelplugin/aether-3.0.0/aether_3_0_0"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	gnmiproto "github.com/openconfig/gnmi/proto/gnmi"
	pb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
	"reflect"
	"time"

	"github.com/onosproject/sdcore-adapter/pkg/synchronizer"
)

var log = logging.GetLogger("synchronizer")

// ModelData is a list of Models returned by gNMI Capabilities for this adapter.
//
// This data was formerly in the config-models repository, but is no longer authoritatively
// stored there due to those models being moved to helm charts used with the onos operator.
//
// NOTE: It's unclear how useful this is -- the actual yang contents of the models are
//       not returned by capabilities, only the names of the models.
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

// Synchronize synchronizes the state to the underlying service.
func (s *Synchronizer) Synchronize(config ygot.ValidatedGoStruct, callbackType gnmi.ConfigCallbackType, path *pb.Path) error {
	var err error
	if callbackType == gnmi.Deleted {
		// do stuff
		err = nil
	} else {
		err = s.enqueue(config, callbackType)
	}
	return err
}

// SynchronizeAndRetry automatically retries if synchronization fails
func (s *Synchronizer) SynchronizeAndRetry(update *SynchronizerUpdate) {
	for {
		// If something new has come along, then don't bother with the one we're working on
		if s.newUpdatesPending() {
			log.Infof("Current synchronizer update has been obsoleted")
			return
		}

		err := s.synchronizeDeviceFunc(update.config)
		if err == nil {
			// Success!
			log.Infof("Synchronization success")
			return
		}

		log.Infof("Synchronization error: %v", err)

		// We erred. Sleep before trying again.
		// Implements a fixed interval for now; We can go exponential should it prove to
		// be a problem.
		time.Sleep(s.retryInterval)
	}
}

// Loop runs an infitite loop servicing synchronization requests.
func (s *Synchronizer) Loop() {
	log.Infof("Starting synchronizer loop")
	for {
		update := s.dequeue()

		log.Infof("Synchronize, type=%s", update.callbackType)

		s.SynchronizeAndRetry(update)

		s.complete()
	}
}

// GetModels gets the list of models.
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

// SetOutputFileName sets the output filename. Obsolete.
func (s *Synchronizer) SetOutputFileName(fileName string) {
	s.outputFileName = fileName
}

// SetPostEnable enables or disables Posting to service
func (s *Synchronizer) SetPostEnable(postEnable bool) {
	s.postEnable = postEnable
}

// SetPostTimeout sets the timeout for post requests.
func (s *Synchronizer) SetPostTimeout(postTimeout time.Duration) {
	s.postTimeout = postTimeout
}

// SetPusher sets the Pusher function for the Synchronizer
func (s *Synchronizer) SetPusher(pusher synchronizer.PusherInterface) {
	s.pusher = pusher
}

// Start the synchronizer by launching the synchronizer loop inside a thread.
func (s *Synchronizer) Start() {
	log.Infof("Synchronizer starting (outputFileName=%s, postEnable=%s, postTimeout=%d)",
		s.outputFileName,
		s.postEnable,
		s.postTimeout)

	// TODO: Eventually we'll create a thread here that waits for config changes
	go s.Loop()
}

// NewSynchronizer creates a new Synchronizer
func NewSynchronizer(outputFileName string, postEnable bool, postTimeout time.Duration) *Synchronizer {
	// By default, push via REST. Test infrastructure can override this.
	p := &RESTPusher{}

	s := &Synchronizer{
		outputFileName: outputFileName,
		postEnable:     postEnable,
		postTimeout:    postTimeout,
		pusher:         p,
		updateChannel:  make(chan *SynchronizerUpdate, 1),
		retryInterval:  5 * time.Second,
	}
	s.synchronizeDeviceFunc = s.SynchronizeDevice
	return s
}
