// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package gnmi implements a gnmi server to mock a device with YANG models.
package synchronizerv4

import (
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewSynchronizer(t *testing.T) {
	sync := NewSynchronizer("", false, 5*time.Second)
	assert.NotNil(t, sync)

	assert.Equal(t, "", sync.outputFileName)
	assert.Equal(t, false, sync.postEnable)
	assert.Equal(t, 5*time.Second, sync.postTimeout)

	sync.SetOutputFileName("/tmp/somefile.json")
	assert.Equal(t, "/tmp/somefile.json", sync.outputFileName)

	sync.SetPostEnable(true)
	assert.Equal(t, true, sync.postEnable)

	sync.SetPostTimeout(7 * time.Second)
	assert.Equal(t, 7*time.Second, sync.postTimeout)
}

func TestSynchronizerLoop(t *testing.T) {
	sync := NewSynchronizer("", false, 5*time.Second)
	assert.NotNil(t, sync)

	config := &mockConfig{}

	sync.retryInterval = 100 * time.Millisecond
	sync.synchronizeDeviceFunc = mockSynchronizeDevice
	sync.Start()

	// Normal synchronization
	mockSynchronizeDeviceReset(0, 0*time.Second)
	err := sync.Synchronize(config, gnmi.Apply)
	assert.Nil(t, err)
	waitForSyncIdle(t, sync, 5*time.Second)
	assert.Equal(t, 1, len(mockSynchronizeDeviceCalls))
	assert.Equal(t, config, mockSynchronizeDeviceCalls[0])

	// Fail and retry once

	mockSynchronizeDeviceReset(1, 0*time.Second)
	err = sync.Synchronize(config, gnmi.Apply)
	assert.Nil(t, err)
	waitForSyncIdle(t, sync, 5*time.Second)
	assert.Equal(t, 1, len(mockSynchronizeDeviceFails))
	assert.Equal(t, 1, len(mockSynchronizeDeviceCalls))
	assert.Equal(t, config, mockSynchronizeDeviceCalls[0])

	// several queued changes should only get the last one
	mockSynchronizeDeviceReset(1, 100*time.Millisecond)
	err = sync.Synchronize(&mockConfig{}, gnmi.Apply) // this one will fail...
	assert.Nil(t, err)
	err = sync.Synchronize(&mockConfig{}, gnmi.Apply) // this one will be ignored...
	assert.Nil(t, err)
	err = sync.Synchronize(&mockConfig{}, gnmi.Apply) // this one will also be ignored...
	assert.Nil(t, err)
	err = sync.Synchronize(config, gnmi.Apply) // this one will succeed!
	assert.Nil(t, err)
	waitForSyncIdle(t, sync, 5*time.Second)
	assert.Equal(t, 1, len(mockSynchronizeDeviceFails))
	assert.Equal(t, 1, len(mockSynchronizeDeviceCalls))
	assert.Equal(t, config, mockSynchronizeDeviceCalls[0])
}
