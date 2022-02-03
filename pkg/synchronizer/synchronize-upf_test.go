// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package synchronizer

import (
	"github.com/golang/mock/gomock"
	models "github.com/onosproject/config-models/modelplugin/aether-2.0.0/aether_2_0_0"
	"github.com/onosproject/sdcore-adapter/pkg/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

func TestSynchronizeSliceUPF(t *testing.T) {
	jsonData, err := ioutil.ReadFile("./testdata/sample-upfslice.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	ent, cs, _, _, site, _ := BuildSampleDeviceGroup() // nolint dogsled
	_, _, _, slice := BuildSampleSlice(ent, site)      // nolint dogsled

	device := &RootDevice{
		Enterprises:          &models.OnfEnterprise_Enterprises{Enterprise: map[string]*Enterprise{"sample-ent": ent}},
		ConnectivityServices: &models.OnfConnectivityService_ConnectivityServices{ConnectivityService: map[string]*ConnectivityService{"sample-cs": cs}},
	}

	scope, err := BuildScope(device, "sample-ent", "sample-site", "sample-cs")
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
	jsonData, err := ioutil.ReadFile("./testdata/sample-upfslice.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	ent, cs, _, _, site, _ := BuildSampleDeviceGroup() // nolint dogsled
	_, _, _, slice := BuildSampleSlice(ent, site)      // nolint dogsled

	slice.Mbr.UplinkBurstSize = aUint32(111222)
	slice.Mbr.DownlinkBurstSize = aUint32(333444)

	device := &RootDevice{
		Enterprises:          &models.OnfEnterprise_Enterprises{Enterprise: map[string]*Enterprise{"sample-ent": ent}},
		ConnectivityServices: &models.OnfConnectivityService_ConnectivityServices{ConnectivityService: map[string]*ConnectivityService{"sample-cs": cs}},
	}

	scope, err := BuildScope(device, "sample-ent", "sample-site", "sample-cs")
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
	jsonData, err := ioutil.ReadFile("./testdata/sample-upfslice1.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	ent, cs, _, _, site, _ := BuildSampleDeviceGroup() // nolint dogsled
	_, _, _, slice := BuildSampleSlice(ent, site)      // nolint dogsled

	slice.Mbr = nil // remove the slice QoS

	device := &RootDevice{
		Enterprises:          &models.OnfEnterprise_Enterprises{Enterprise: map[string]*Enterprise{"sample-ent": ent}},
		ConnectivityServices: &models.OnfConnectivityService_ConnectivityServices{ConnectivityService: map[string]*ConnectivityService{"sample-cs": cs}},
	}

	scope, err := BuildScope(device, "sample-ent", "sample-site", "sample-cs")
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
