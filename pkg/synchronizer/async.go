// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

// Package synchronizer implements a synchronizer for Aether models
package synchronizer

import (
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	"github.com/openconfig/ygot/ygot"
	"sync/atomic"
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
			atomic.AddInt32(&s.busy, -1)
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
	update := ConfigUpdate{
		config:       configCopy.(ygot.ValidatedGoStruct),
		callbackType: callbackType,
	}

	// Increment our busy count
	atomic.AddInt32(&s.busy, 1)

	// We don't care about any pending synchronizations; throw away any old ones
	// and queue the latest one.
	s.drain()
	s.updateChannel <- &update

	return nil
}

// Dequeue an update request. This call will block until a request is ready.
func (s *Synchronizer) dequeue() *ConfigUpdate {
	update := <-s.updateChannel
	return update
}

// Call complete when the synchronizer has finished servicing a request
func (s *Synchronizer) complete() {
	atomic.AddInt32(&s.busy, -1)
}

// Returns true if new updates have arrived, not including the one currently
// being serviced.
func (s *Synchronizer) newUpdatesPending() bool {
	return len(s.updateChannel) > 0
}

// Returns true if the synchronizer is idle; if there are no requests being
// worked on and no pending requests.
func (s *Synchronizer) isIdle() bool {
	return atomic.LoadInt32(&s.busy) == 0
}
