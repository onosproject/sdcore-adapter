// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizer implements a synchronizer for Aether models
package synchronizer

import (
	models "github.com/onosproject/config-models/modelplugin/aether-2.0.0/aether_2_0_0"
	modelplugin "github.com/onosproject/config-models/modelplugin/aether-2.0.0/modelplugin"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	pb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
	"reflect"
	"time"
)

var log = logging.GetLogger("synchronizer")

// Synchronize synchronizes the state to the underlying service.
func (s *Synchronizer) Synchronize(config ygot.ValidatedGoStruct, callbackType gnmi.ConfigCallbackType, path *pb.Path) error {
	var err error
	if callbackType == gnmi.Deleted {
		return s.HandleDelete(config, path)
	}

	if callbackType == gnmi.Forced {
		s.CacheInvalidate() // invalidate the post cache if this resync was forced by Diagnostic API
	}

	err = s.enqueue(config, callbackType)
	return err
}

// SynchronizeAndRetry automatically retries if synchronization fails
func (s *Synchronizer) SynchronizeAndRetry(update *ConfigUpdate) {
	for {
		// If something new has come along, then don't bother with the one we're working on
		if s.newUpdatesPending() {
			log.Infof("Current synchronizer update has been obsoleted")
			return
		}

		pushErrors, err := s.synchronizeDeviceFunc(update.config)
		if err != nil {
			log.Errorf("Synchronization error: %v", err)
			return
		}

		if pushErrors == 0 {
			log.Infof("Synchronization success")
			return
		}

		log.Infof("Synchronization encountered %d push errors, scheduling retry", pushErrors)

		// We failed to push something to the core. Sleep before trying again.
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
	model := gnmi.NewModel(modelplugin.ModelData,
		reflect.TypeOf((*models.Device)(nil)),
		models.SchemaTree["Device"],
		models.Unmarshal,
		//models.ΛEnum  // NOTE: There is no Enum in the aether models? So use a blank map.
		map[string]map[int64]ygot.EnumDefinition{},
	)

	return model
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

// WithPostEnable sets the postEnable option
func WithPostEnable(postEnable bool) SynchronizerOption {
	return func(s *Synchronizer) {
		s.postEnable = postEnable
	}
}

// WithPostTimeout sets the postTimeout option
func WithPostTimeout(postTimeout time.Duration) SynchronizerOption {
	return func(s *Synchronizer) {
		s.postTimeout = postTimeout
	}
}

// WithPartialUpdateEnable sets the partialUpdateEnable option
func WithPartialUpdateEnable(partialUpdateEnable bool) SynchronizerOption {
	return func(s *Synchronizer) {
		s.partialUpdateEnable = partialUpdateEnable
	}
}

// WithOutputFileName sets the outputFileName option
func WithOutputFileName(outputFileName string) SynchronizerOption {
	return func(s *Synchronizer) {
		s.outputFileName = outputFileName
	}
}

// WithPusher sets the pusher for pushing REST to the core or UPF
func WithPusher(pusher PusherInterface) SynchronizerOption {
	return func(s *Synchronizer) {
		s.pusher = pusher
	}
}

// NewSynchronizer creates a new Synchronizer
func NewSynchronizer(opts ...SynchronizerOption) *Synchronizer {
	// By default, push via REST. Test infrastructure can override this.
	p := &RESTPusher{}

	s := &Synchronizer{
		pusher:              p,
		postEnable:          true,
		partialUpdateEnable: DefaultPartialUpdateEnable,
		postTimeout:         DefaultPostTimeout,
		updateChannel:       make(chan *ConfigUpdate, 1),
		retryInterval:       5 * time.Second,
		cache:               map[string]interface{}{},
	}

	for _, opt := range opts {
		opt(s)
	}

	s.synchronizeDeviceFunc = s.SynchronizeDevice
	return s
}
