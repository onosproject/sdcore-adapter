// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package synchronizer

import (
	"github.com/golang/mock/gomock"
	"github.com/onosproject/sdcore-adapter/pkg/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestSynchronizeSliceUPF(t *testing.T) {
	jsonData, err := os.ReadFile("./testdata/sample-upfslice.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	_, device := BuildSampleConfig()
	slice := device.Site["sample-site"].Slice["sample-slice"]

	scope, err := BuildScope(device, "sample-ent", "sample-site")
	assert.Nil(t, err)

	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(jsonData)
		return nil
	}).AnyTimes()
	pushFailures, err := s.SynchronizeSliceUPF(scope, slice)
	assert.Nil(t, err)
	json, okay := pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	assert.Equal(t, 0, pushFailures)
	if okay {
		require.JSONEq(t, string(jsonData), json)
	}
}

func TestSynchronizeSliceUPFWithBurst(t *testing.T) {
	jsonData, err := os.ReadFile("./testdata/sample-upfslice.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	_, device := BuildSampleConfig()
	slice := device.Site["sample-site"].Slice["sample-slice"]

	slice.Mbr.UplinkBurstSize = aUint32(111222)
	slice.Mbr.DownlinkBurstSize = aUint32(333444)

	scope, err := BuildScope(device, "sample-ent", "sample-site")
	assert.Nil(t, err)

	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(jsonData)
		return nil
	}).AnyTimes()
	pushFailures, err := s.SynchronizeSliceUPF(scope, slice)
	assert.Nil(t, err)
	json, okay := pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	assert.Equal(t, 0, pushFailures)
	if okay {
		require.JSONEq(t, string(jsonData), json)
	}
}

func TestSynchronizeSliceUPFNoSliceQos(t *testing.T) {
	jsonData, err := os.ReadFile("./testdata/sample-upfslice1.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	_, device := BuildSampleConfig()
	slice := device.Site["sample-site"].Slice["sample-slice"]

	slice.Mbr = nil // remove the slice QoS

	scope, err := BuildScope(device, "sample-ent", "sample-site")
	assert.Nil(t, err)

	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(jsonData)
		return nil
	}).AnyTimes()
	pushFailures, err := s.SynchronizeSliceUPF(scope, slice)
	assert.Nil(t, err)
	json, okay := pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	assert.Equal(t, 0, pushFailures)
	if okay {
		require.JSONEq(t, string(jsonData), json)
	}
}
