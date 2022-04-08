// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package synchronizer

import (
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewSynchronizer(t *testing.T) {
	sync := NewSynchronizer()
	assert.NotNil(t, sync)

	assert.Equal(t, "", sync.outputFileName)
	assert.Equal(t, true, sync.postEnable)
	assert.Equal(t, 10*time.Second, sync.postTimeout)
	assert.Equal(t, true, sync.partialUpdateEnable)

	sync = NewSynchronizer(
		WithOutputFileName("/tmp/somefile.json"),
		WithPostEnable(false),
		WithPostTimeout(7*time.Second),
		WithPartialUpdateEnable(false),
	)

	assert.Equal(t, "/tmp/somefile.json", sync.outputFileName)
	assert.Equal(t, false, sync.postEnable)
	assert.Equal(t, 7*time.Second, sync.postTimeout)
	assert.Equal(t, false, sync.partialUpdateEnable)
}

func TestSynchronizerLoop(t *testing.T) {
	sync := NewSynchronizer()
	assert.NotNil(t, sync)

	config := gnmi.ConfigForest{} //&mockConfig{}

	sync.retryInterval = 100 * time.Millisecond
	sync.synchronizeDeviceFunc = mockSynchronizeDevice
	sync.Start()

	// Normal synchronization
	mockSynchronizeDeviceReset(0, 0, 0*time.Second)
	err := sync.Synchronize(config, gnmi.Apply, "sample-ent", nil)
	assert.Nil(t, err)
	waitForSyncIdle(t, sync, 5*time.Second)
	assert.Equal(t, 1, len(mockSynchronizeDeviceCalls))
	assert.Equal(t, config, mockSynchronizeDeviceCalls[0])

	// Fail and retry once

	mockSynchronizeDeviceReset(0, 1, 0*time.Second)
	err = sync.Synchronize(config, gnmi.Apply, "sample-ent", nil)
	assert.Nil(t, err)
	waitForSyncIdle(t, sync, 5*time.Second)
	assert.Equal(t, 1, len(mockSynchronizeDevicePushFails))
	assert.Equal(t, 1, len(mockSynchronizeDeviceCalls))
	assert.Equal(t, config, mockSynchronizeDeviceCalls[0])

	// several queued changes should only get the last one
	mockSynchronizeDeviceReset(0, 1, 100*time.Millisecond)
	err = sync.Synchronize(gnmi.ConfigForest{}, gnmi.Apply, "sample-ent", nil) // this one will fail...
	assert.Nil(t, err)
	err = sync.Synchronize(gnmi.ConfigForest{}, gnmi.Apply, "sample-ent", nil) // this one will be ignored...
	assert.Nil(t, err)
	err = sync.Synchronize(gnmi.ConfigForest{}, gnmi.Apply, "sample-ent", nil) // this one will also be ignored...
	assert.Nil(t, err)
	err = sync.Synchronize(config, gnmi.Apply, "sample-ent", nil) // this one will succeed!
	assert.Nil(t, err)
	waitForSyncIdle(t, sync, 5*time.Second)
	assert.Equal(t, 1, len(mockSynchronizeDevicePushFails))
	assert.Equal(t, 1, len(mockSynchronizeDeviceCalls))
	assert.Equal(t, config, mockSynchronizeDeviceCalls[0])
}
