// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizerv4 implements the version 4 synchronizer.
package synchronizerv4

import (
	"errors"
	"github.com/openconfig/ygot/ygot"
	"reflect"
	"testing"
	"time"
)

// MockConfig is an empty ValidatedGoStruct
type mockConfig struct{}

func (m *mockConfig) Validate(...ygot.ValidationOption) error { return nil }
func (m *mockConfig) ΛEnumTypeMap() map[string][]reflect.Type { return nil }
func (m *mockConfig) IsYANGGoStruct()                         {}

var (
	mockSynchronizeDeviceCalls     []ygot.ValidatedGoStruct // list of calls to MockSynchronizeDevice that succeeded
	mockSynchronizeDeviceFails     []ygot.ValidatedGoStruct // list of calls to MockSynchronizeDevice that failed
	mockSynchronizeDeviceFailCount int                      // Cause MockSynchronizeDevice to fail the specified number of times
	mockSynchronizeDeviceDelay     time.Duration            // Cause MockSynchronizeDevice to take some time
)

func mockSynchronizeDevice(config ygot.ValidatedGoStruct) error {
	time.Sleep(mockSynchronizeDeviceDelay)
	if mockSynchronizeDeviceFailCount > 0 {
		mockSynchronizeDeviceFailCount--
		mockSynchronizeDeviceFails = append(mockSynchronizeDeviceFails, config)
		return errors.New("Mock error")
	}
	mockSynchronizeDeviceCalls = append(mockSynchronizeDeviceCalls, config)
	return nil
}

// Reset mockSynchronizeDevice for a new set of tests
//    failCount = number of times to fail befor returning success
//    delay = amount of time to delay before returning
func mockSynchronizeDeviceReset(failCount int, delay time.Duration) {
	mockSynchronizeDeviceCalls = nil
	mockSynchronizeDeviceFails = nil
	mockSynchronizeDeviceFailCount = failCount
	mockSynchronizeDeviceDelay = delay
}

// Wait for the synchronizer to be idle. Used in unit tests to perform asserts
// when a predictable state is reached.
func waitForSyncIdle(t *testing.T, s *Synchronizer, timeout time.Duration) {
	elapsed := 0 * time.Second
	for {
		if s.isIdle() {
			return
		}
		time.Sleep(100 * time.Millisecond)
		elapsed += 100 * time.Millisecond
		if elapsed > timeout {
			t.Fatal("waitForSyncIdle failed to complete")
		}
	}
}
