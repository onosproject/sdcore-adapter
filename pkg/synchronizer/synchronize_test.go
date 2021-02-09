// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package gnmi implements a gnmi server to mock a device with YANG models.
package synchronizer

import (
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
