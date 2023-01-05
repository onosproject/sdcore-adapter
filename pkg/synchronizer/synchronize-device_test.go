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

func TestSynchronizeDeviceDeviceGroupWithQos(t *testing.T) {

	jsonData, err := os.ReadFile("./testdata/sample-dg.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	config, _ := BuildSampleConfig()

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/network-slice/sample-slice", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		return nil
	}).AnyTimes()

	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		return nil
	}).AnyTimes()

	pushErrors, err := s.SynchronizeDevice(config)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)

	json, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonData), json)
	}
}

func TestSynchronizeDeviceDeviceGroupWithQosSpecifiedPelrPDB(t *testing.T) {

	jsonDataDg, err := os.ReadFile("./testdata/sample-dg-pelr-pdb.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	config, device := BuildSampleConfig()
	device.TrafficClass["sample-traffic-class"].Pelr = aInt8(3)
	device.TrafficClass["sample-traffic-class"].Pdb = aUint16(400)

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/network-slice/sample-slice", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		return nil
	}).AnyTimes()

	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		return nil
	}).AnyTimes()

	pushErrors, err := s.SynchronizeDevice(config)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)

	json, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataDg), json)
	}
}

func TestSynchronizeDeviceDeviceGroupWithQosButNoTC(t *testing.T) {

	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	config, device := BuildSampleConfig()
	device.TrafficClass = nil

	pushErrors, err := s.SynchronizeDevice(config)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)

	// The above will fail synchronization with a nonfatal error because TrafficClass is missing

	_, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.False(t, okay)
}

/*

  TODO: needs update for Device + SimCard

func TestSynchronizeDeviceDeviceGroupManyIMSIDeterministic(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := []string{}
	s := NewSynchronizer(WithPusher(mockPusher))

	ent, cs, tcList, ipd, site, dg := BuildSampleDeviceGroup()

	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("imsi%d", i)
		imsi := &models.OnfDeviceGroup_DeviceGroup_DeviceGroup_Imsis{
			ImsiRangeFrom: aUint64(uint64(100 + i)),
			ImsiId:        aStr(name),
		}

		dg.Imsis[*imsi.ImsiId] = imsi
	}

	device := &RootDevice{
		Enterprises:          &models.OnfEnterprise_Enterprises{Enterprise: map[string]*Enterprise{"sample-ent": ent}},
		ConnectivityServices: &models.OnfConnectivityService_ConnectivityServices{ConnectivityService: map[string]*ConnectivityService{"sample-cs": cs}},
	}

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes = append(pushes, string(data))
		return nil
	}).AnyTimes()

	// Push 10 times
	for i := 0; i < 10; i++ {
		// Make sure the cache does not prevent multiple pushes
		s.CacheInvalidate()

		// Do a push
		pushErrors, err := s.SynchronizeDevice(device)
		assert.Equal(t, 0, pushErrors)
		assert.Nil(t, err)
	}

	assert.Equal(t, 10, len(pushes))

	// All 10 pushes should have identical text
	for i := 1; i < 10; i++ {
		assert.Equal(t, pushes[0], pushes[i])
	}
}
*/

func TestSynchronizeDeviceDeviceGroupLinkedToVCS(t *testing.T) {

	jsonDataDg, err := os.ReadFile("./testdata/sample-dg.json")
	assert.NoError(t, err)
	jsonDataUpfSlice, err := os.ReadFile("./testdata/sample-upfslice.json")
	assert.NoError(t, err)
	jsonDataCoreSlice, err := os.ReadFile("./testdata/sample-coreslice.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	config, _ := BuildSampleConfig()

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/network-slice/sample-slice", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()

	pushErrors, err := s.SynchronizeDevice(config)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)

	// Note: With an associated VCS, we'll pick up the QoS settings
	json, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataDg), json)
	}
	json, okay = pushes["http://5gcore/v1/network-slice/sample-slice"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataCoreSlice), json)
	}
	json, okay = pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataUpfSlice), json)
	}
}

func TestSynchronizeVCS(t *testing.T) {
	jsonDataDg, err := os.ReadFile("./testdata/sample-dg.json")
	assert.NoError(t, err)
	jsonDataUpfSlice, err := os.ReadFile("./testdata/sample-upfslice.json")
	assert.NoError(t, err)
	jsonDataCoreSlice, err := os.ReadFile("./testdata/sample-coreslice.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	config, _ := BuildSampleConfig()

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/network-slice/sample-slice", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	pushErrors, err := s.SynchronizeDevice(config)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)

	json, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataDg), json)
	}
	json, okay = pushes["http://5gcore/v1/network-slice/sample-slice"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataCoreSlice), json)
	}
	json, okay = pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataUpfSlice), json)
	}
}

func TestSynchronizeVCSSSTLeadingZero(t *testing.T) {
	jsonDataDg, err := os.ReadFile("./testdata/sample-dg.json")
	assert.NoError(t, err)
	jsonDataUpfSlice, err := os.ReadFile("./testdata/sample-upfslice.json")
	assert.NoError(t, err)
	jsonDataCoreSlice, err := os.ReadFile("./testdata/sample-coreslice-leading-zero.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	config, device := BuildSampleConfig()
	device.Site["sample-site"].Slice["sample-slice"].Sd = aStr("000111")
	device.Site["sample-site"].Slice["sample-slice"].Sst = aStr("002")

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/network-slice/sample-slice", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	pushErrors, err := s.SynchronizeDevice(config)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)

	json, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataDg), json)
	}
	json, okay = pushes["http://5gcore/v1/network-slice/sample-slice"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataCoreSlice), json)
	}
	json, okay = pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataUpfSlice), json)
	}
}

func TestSynchronizeVCSIMSILeadingZero(t *testing.T) {
	jsonDataDg, err := os.ReadFile("./testdata/sample-dg-imsi-leading-zero.json")
	assert.NoError(t, err)
	jsonDataUpfSlice, err := os.ReadFile("./testdata/sample-upfslice.json")
	assert.NoError(t, err)
	jsonDataCoreSlice, err := os.ReadFile("./testdata/sample-coreslice.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	config, device := BuildSampleConfig()
	device.Site["sample-site"].SimCard["sample-sim"].Imsi = aStr("012345678901234")

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/network-slice/sample-slice", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	pushErrors, err := s.SynchronizeDevice(config)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)

	json, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataDg), json)
	}
	json, okay = pushes["http://5gcore/v1/network-slice/sample-slice"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataCoreSlice), json)
	}
	json, okay = pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataUpfSlice), json)
	}
}

func TestSynchronizeVCSDisabledSimCard(t *testing.T) {
	jsonDataDg, err := os.ReadFile("./testdata/sample-dg-sim-disabled.json")
	assert.NoError(t, err)
	jsonDataUpfSlice, err := os.ReadFile("./testdata/sample-upfslice.json")
	assert.NoError(t, err)
	jsonDataCoreSlice, err := os.ReadFile("./testdata/sample-coreslice.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	config, device := BuildSampleConfig()
	device.Site["sample-site"].SimCard["sample-sim"].Enable = aBool(false)

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/network-slice/sample-slice", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	pushErrors, err := s.SynchronizeDevice(config)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)

	json, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataDg), json)
	}
	json, okay = pushes["http://5gcore/v1/network-slice/sample-slice"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataCoreSlice), json)
	}
	json, okay = pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataUpfSlice), json)
	}
}

func TestSynchronizeVCSAllowAll(t *testing.T) {
	jsonDataDg, err := os.ReadFile("./testdata/sample-dg.json")
	assert.NoError(t, err)
	jsonDataCoreSlice, err := os.ReadFile("./testdata/sample-coreslice-allow-all.json")
	assert.NoError(t, err)
	jsonDataUpfSlice, err := os.ReadFile("./testdata/sample-upfslice.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	config, device := BuildSampleConfig()
	device.Site["sample-site"].Slice["sample-slice"].DefaultBehavior = aStr("ALLOW-ALL")

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/network-slice/sample-slice", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	pushErrors, err := s.SynchronizeDevice(config)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)

	json, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataDg), json)
	}
	json, okay = pushes["http://5gcore/v1/network-slice/sample-slice"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataCoreSlice), json)
	}
	json, okay = pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataUpfSlice), json)
	}
}

func TestSynchronizeVCSAllowPublic(t *testing.T) {
	jsonDataDg, err := os.ReadFile("./testdata/sample-dg.json")
	assert.NoError(t, err)
	jsonDataCoreSlice, err := os.ReadFile("./testdata/sample-coreslice-allow-public.json")
	assert.NoError(t, err)
	jsonDataUpfSlice, err := os.ReadFile("./testdata/sample-upfslice.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	config, device := BuildSampleConfig()

	device.Site["sample-site"].Slice["sample-slice"].DefaultBehavior = aStr("ALLOW-PUBLIC")

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/network-slice/sample-slice", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	pushErrors, err := s.SynchronizeDevice(config)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)

	json, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataDg), json)
	}
	json, okay = pushes["http://5gcore/v1/network-slice/sample-slice"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataCoreSlice), json)
	}
	json, okay = pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataUpfSlice), json)
	}
}

func TestSynchronizeVCSTwoEnpoints(t *testing.T) {
	jsonDataDg, err := os.ReadFile("./testdata/sample-dg.json")
	assert.NoError(t, err)
	jsonDataUpfSlice, err := os.ReadFile("./testdata/sample-upfslice.json")
	assert.NoError(t, err)
	jsonDataCoreSlice, err := os.ReadFile("./testdata/sample-coreslice1.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	config, device := BuildSampleConfig()

	mbr3 := &ApplicationEndpointMbr{
		Uplink:   aUint64(44332211),
		Downlink: aUint64(88776655),
	}

	ep3 := &ApplicationEndpoint{
		EndpointId:   aStr("zep3"),
		PortStart:    aUint16(5555),
		PortEnd:      aUint16(5556),
		Protocol:     aStr("UDP"),
		Mbr:          mbr3,
		TrafficClass: aStr("sample-traffic-class"),
	}

	device.Application["sample-app2"].Endpoint["zep3"] = ep3

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/network-slice/sample-slice", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	pushErrors, err := s.SynchronizeDevice(config)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)

	json, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataDg), json)
	}
	json, okay = pushes["http://5gcore/v1/network-slice/sample-slice"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataCoreSlice), json)
	}
	json, okay = pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataUpfSlice), json)
	}
}

func TestSynchronizeVCSEmptySD(t *testing.T) {
	jsonDataDg, err := os.ReadFile("./testdata/sample-dg.json")
	assert.NoError(t, err)
	jsonDataUpfSlice, err := os.ReadFile("./testdata/sample-upfslice.json")
	assert.NoError(t, err)
	jsonDataCoreSlice, err := os.ReadFile("./testdata/sample-coreslice2.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	config, device := BuildSampleConfig()

	// Set the SD to nil.
	device.Site["sample-site"].Slice["sample-slice"].Sd = nil

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/network-slice/sample-slice", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	pushErrors, err := s.SynchronizeDevice(config)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)
	json, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataDg), json)
	}
	json, okay = pushes["http://5gcore/v1/network-slice/sample-slice"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataCoreSlice), json)
	}
	json, okay = pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataUpfSlice), json)
	}
}

func TestSynchronizeVCSDisabledDG(t *testing.T) {
	jsonDataDg, err := os.ReadFile("./testdata/sample-dg.json")
	assert.NoError(t, err)
	jsonDataUpfSlice, err := os.ReadFile("./testdata/sample-upfslice1.json")
	assert.NoError(t, err)
	jsonDataCoreSlice, err := os.ReadFile("./testdata/sample-coreslice3.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	config, device := BuildSampleConfig()

	// Disable the one and only DeviceGroup
	device.Site["sample-site"].Slice["sample-slice"].DeviceGroup["sample-dg"].Enable = aBool(false)

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/network-slice/sample-slice", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	pushErrors, err := s.SynchronizeDevice(config)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)
	json, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataDg), json)
	}
	json, okay = pushes["http://5gcore/v1/network-slice/sample-slice"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataCoreSlice), json)
	}
	json, okay = pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataUpfSlice), json)
	}
}

func TestSynchronizeVCSMissingDG(t *testing.T) {
	jsonDataUpfSlice, err := os.ReadFile("./testdata/sample-upfslice1.json")
	assert.NoError(t, err)
	jsonDataCoreSlice, err := os.ReadFile("./testdata/sample-coreslice3.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	config, device := BuildSampleConfig()

	// Delete the one and only DeviceGroup
	device.Site["sample-site"].Slice["sample-slice"].DeviceGroup = nil

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/network-slice/sample-slice", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	pushErrors, err := s.SynchronizeDevice(config)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)

	// Because the DG is not in a slice, we have no way to determine the Core. Therefore,
	// we cannot push the DG
	_, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.False(t, okay)

	json, okay := pushes["http://5gcore/v1/network-slice/sample-slice"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataCoreSlice), json)
	}
	json, okay = pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataUpfSlice), json)
	}
}
