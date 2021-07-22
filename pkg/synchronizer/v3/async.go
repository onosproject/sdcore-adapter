// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizerv3 implements a synchronizer for Aether v3 models
package synchronizerv3

import (
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	"github.com/openconfig/ygot/ygot"
)

/*
 * Synchronizer Async Support
 *
 * Implements a buffer between the gNMI server and the synchronizer. New updates are queued
 * for processing. If an update is pending and a new update is received, then the older update
 * will be discarded in favor of the newer one (there's no reason to keep old updates, as they're
 * fully obsoleted by newer updates)
 */

// Drain the synchronizer of any queued updates
func (s *Synchronizer) drain() {
L:
	for {
		select {
		case <-s.updateChannel:
			log.Infof("Drained a pending synchronization request")
		default:
			break L
		}
	}
}

// Queue an update request for future processing
func (s *Synchronizer) enqueue(config ygot.ValidatedGoStruct, callbackType gnmi.ConfigCallbackType) error {
	// Make a copy of the gostruct; we don't want it to change out from under us
	// if the gnmi server is updating it.
	configCopy, err := ygot.DeepCopy(config)
	if err != nil {
		return err
	}

	// This conversion is safe as DeepCopy will use the same underlying type as
	// `config`, which is a ValidatedGoStruct.
	update := SynchronizerUpdate{
		config:       configCopy.(ygot.ValidatedGoStruct),
		callbackType: callbackType,
	}

	// We don't care about any pending synchronizations; throw away any old ones
	// and queue the latest one.
	s.drain()
	s.updateChannel <- &update

	return nil
}

// Dequeue an update request. This call will block until a request is ready.
func (s *Synchronizer) dequeue() *SynchronizerUpdate {
	update := <-s.updateChannel
	return update
}

func (s *Synchronizer) newUpdatesPending() bool {
	return len(s.updateChannel) > 0
}
