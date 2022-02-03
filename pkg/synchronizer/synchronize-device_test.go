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
	"os"
	"testing"
)

// an empty device should yield empty json
func TestSynchronizeDeviceEmpty(t *testing.T) {
	// Get a temporary file name and defer deletion of the file
	f, err := ioutil.TempFile("", "synchronizer-json")
	assert.Nil(t, err)
	tempFileName := f.Name()
	defer func() {
		assert.Nil(t, os.Remove(tempFileName))
	}()

	s := NewSynchronizer(WithOutputFileName(tempFileName))
	device := models.Device{}
	pushErrors, err := s.SynchronizeDevice(&device)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)

	content, err := ioutil.ReadFile(tempFileName)
	assert.Nil(t, err)
	assert.Equal(t, "", string(content))
}

func TestSynchronizeDeviceCSEnt(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	ent := MakeEnterprise("sample-ent-desc", "sample-ent-dn", "sample-ent", []string{"sample-cs"})
	cs := MakeCs("sample-cs-desc", "sample-cs-dn", "sample-cs")

	s := NewSynchronizer(WithPusher(mockPusher))

	// TODO: This is supposed to have all other objects empty, but in 2.0 they got embedded into Enterprises
	device := &RootDevice{
		Enterprises:          &models.OnfEnterprise_Enterprises{Enterprise: map[string]*Enterprise{"sample-ent": ent}},
		ConnectivityServices: &models.OnfConnectivityService_ConnectivityServices{ConnectivityService: map[string]*ConnectivityService{"sample-cs": cs}},
	}

	pushErrors, err := s.SynchronizeDevice(device)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)
}

func TestSynchronizeDeviceDeviceGroupWithQos(t *testing.T) {

	jsonData, err := ioutil.ReadFile("./testdata/sample-dg.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	ent, cs, _, _, _, _ := BuildSampleDeviceGroup() // nolint dogsled

	device := &RootDevice{
		Enterprises:          &models.OnfEnterprise_Enterprises{Enterprise: map[string]*Enterprise{"sample-ent": ent}},
		ConnectivityServices: &models.OnfConnectivityService_ConnectivityServices{ConnectivityService: map[string]*ConnectivityService{"sample-cs": cs}},
	}

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	pushErrors, err := s.SynchronizeDevice(device)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)

	json, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonData), json)
	}
}

func TestSynchronizeDeviceDeviceGroupWithQosSpecifiedPelrPDB(t *testing.T) {

	jsonDataDg, err := ioutil.ReadFile("./testdata/sample-dg1.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	ent, cs, tcList, _, _, _ := BuildSampleDeviceGroup() // nolint dogsled

	tcList["sample-traffic-class"].Pelr = aInt8(3)
	tcList["sample-traffic-class"].Pdb = aUint16(400)

	device := &RootDevice{
		Enterprises:          &models.OnfEnterprise_Enterprises{Enterprise: map[string]*Enterprise{"sample-ent": ent}},
		ConnectivityServices: &models.OnfConnectivityService_ConnectivityServices{ConnectivityService: map[string]*ConnectivityService{"sample-cs": cs}},
	}

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()

	pushErrors, err := s.SynchronizeDevice(device)
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

	ent, cs, _, _, _, dg := BuildSampleDeviceGroup() // nolint dogsled

	dg.TrafficClass = nil

	device := &RootDevice{
		Enterprises:          &models.OnfEnterprise_Enterprises{Enterprise: map[string]*Enterprise{"sample-ent": ent}},
		ConnectivityServices: &models.OnfConnectivityService_ConnectivityServices{ConnectivityService: map[string]*ConnectivityService{"sample-cs": cs}},
	}
	pushErrors, err := s.SynchronizeDevice(device)
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

	jsonDataDg, err := ioutil.ReadFile("./testdata/sample-dg.json")
	assert.NoError(t, err)
	jsonDataUpfSlice, err := ioutil.ReadFile("./testdata/sample-upfslice.json")
	assert.NoError(t, err)
	jsonDataCoreSlice, err := ioutil.ReadFile("./testdata/sample-coreslice.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	ent, cs, _, _, site, _ := BuildSampleDeviceGroup() // nolint dogsled
	_, _, _, _ = BuildSampleSlice(ent, site)           // nolint dogsled

	device := &RootDevice{
		Enterprises:          &models.OnfEnterprise_Enterprises{Enterprise: map[string]*Enterprise{"sample-ent": ent}},
		ConnectivityServices: &models.OnfConnectivityService_ConnectivityServices{ConnectivityService: map[string]*ConnectivityService{"sample-cs": cs}},
	}

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

	pushErrors, err := s.SynchronizeDevice(device)
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
	jsonDataDg, err := ioutil.ReadFile("./testdata/sample-dg.json")
	assert.NoError(t, err)
	jsonDataUpfSlice, err := ioutil.ReadFile("./testdata/sample-upfslice.json")
	assert.NoError(t, err)
	jsonDataCoreSlice, err := ioutil.ReadFile("./testdata/sample-coreslice.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	ent, cs, _, _, site, _ := BuildSampleDeviceGroup() // nolint dogsled
	_, _, _, _ = BuildSampleSlice(ent, site)           // nolint dogsled

	device := &RootDevice{
		Enterprises:          &models.OnfEnterprise_Enterprises{Enterprise: map[string]*Enterprise{"sample-ent": ent}},
		ConnectivityServices: &models.OnfConnectivityService_ConnectivityServices{ConnectivityService: map[string]*ConnectivityService{"sample-cs": cs}},
	}

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
	pushErrors, err := s.SynchronizeDevice(device)
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
	jsonDataDg, err := ioutil.ReadFile("./testdata/sample-dg.json")
	assert.NoError(t, err)
	jsonDataCoreSlice, err := ioutil.ReadFile("./testdata/sample-coreslice-allow-all.json")
	assert.NoError(t, err)
	jsonDataUpfSlice, err := ioutil.ReadFile("./testdata/sample-upfslice.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	ent, cs, _, _, site, _ := BuildSampleDeviceGroup() // nolint dogsled
	_, _, _, slice := BuildSampleSlice(ent, site)      // nolint dogsled

	slice.DefaultBehavior = aStr("ALLOW-ALL")

	device := &RootDevice{
		Enterprises:          &models.OnfEnterprise_Enterprises{Enterprise: map[string]*Enterprise{"sample-ent": ent}},
		ConnectivityServices: &models.OnfConnectivityService_ConnectivityServices{ConnectivityService: map[string]*ConnectivityService{"sample-cs": cs}},
	}

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
	pushErrors, err := s.SynchronizeDevice(device)
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
	jsonDataDg, err := ioutil.ReadFile("./testdata/sample-dg.json")
	assert.NoError(t, err)
	jsonDataCoreSlice, err := ioutil.ReadFile("./testdata/sample-coreslice-allow-public.json")
	assert.NoError(t, err)
	jsonDataUpfSlice, err := ioutil.ReadFile("./testdata/sample-upfslice.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	ent, cs, _, _, site, _ := BuildSampleDeviceGroup() // nolint dogsled
	_, _, _, slice := BuildSampleSlice(ent, site)      // nolint dogsled

	slice.DefaultBehavior = aStr("ALLOW-PUBLIC")

	device := &RootDevice{
		Enterprises:          &models.OnfEnterprise_Enterprises{Enterprise: map[string]*Enterprise{"sample-ent": ent}},
		ConnectivityServices: &models.OnfConnectivityService_ConnectivityServices{ConnectivityService: map[string]*ConnectivityService{"sample-cs": cs}},
	}

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
	pushErrors, err := s.SynchronizeDevice(device)
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
	jsonDataDg, err := ioutil.ReadFile("./testdata/sample-dg.json")
	assert.NoError(t, err)
	jsonDataUpfSlice, err := ioutil.ReadFile("./testdata/sample-upfslice.json")
	assert.NoError(t, err)
	jsonDataCoreSlice, err := ioutil.ReadFile("./testdata/sample-coreslice1.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	ent, cs, _, _, site, _ := BuildSampleDeviceGroup() // nolint dogsled
	apps, _, _, _ := BuildSampleSlice(ent, site)       // nolint dogsled

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

	apps["sample-app2"].Endpoint["zep3"] = ep3

	device := &RootDevice{
		Enterprises:          &models.OnfEnterprise_Enterprises{Enterprise: map[string]*Enterprise{"sample-ent": ent}},
		ConnectivityServices: &models.OnfConnectivityService_ConnectivityServices{ConnectivityService: map[string]*ConnectivityService{"sample-cs": cs}},
	}

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
	pushErrors, err := s.SynchronizeDevice(device)
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
	jsonDataDg, err := ioutil.ReadFile("./testdata/sample-dg.json")
	assert.NoError(t, err)
	jsonDataUpfSlice, err := ioutil.ReadFile("./testdata/sample-upfslice.json")
	assert.NoError(t, err)
	jsonDataCoreSlice, err := ioutil.ReadFile("./testdata/sample-coreslice2.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	ent, cs, _, _, site, _ := BuildSampleDeviceGroup() // nolint dogsled
	_, _, _, slice := BuildSampleSlice(ent, site)      // nolint dogsled

	// Set the SD to nil.
	slice.Sd = nil

	device := &RootDevice{
		Enterprises:          &models.OnfEnterprise_Enterprises{Enterprise: map[string]*Enterprise{"sample-ent": ent}},
		ConnectivityServices: &models.OnfConnectivityService_ConnectivityServices{ConnectivityService: map[string]*ConnectivityService{"sample-cs": cs}},
	}

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
	pushErrors, err := s.SynchronizeDevice(device)
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
	jsonDataDg, err := ioutil.ReadFile("./testdata/sample-dg.json")
	assert.NoError(t, err)
	jsonDataUpfSlice, err := ioutil.ReadFile("./testdata/sample-upfslice1.json")
	assert.NoError(t, err)
	jsonDataCoreSlice, err := ioutil.ReadFile("./testdata/sample-coreslice3.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	ent, cs, _, _, site, dg := BuildSampleDeviceGroup() // nolint dogsled
	_, _, _, slice := BuildSampleSlice(ent, site)       // nolint dogsled

	// Disable the one and only DeviceGroup
	slice.DeviceGroup[*dg.DeviceGroupId].Enable = aBool(false)

	device := &RootDevice{
		Enterprises:          &models.OnfEnterprise_Enterprises{Enterprise: map[string]*Enterprise{"sample-ent": ent}},
		ConnectivityServices: &models.OnfConnectivityService_ConnectivityServices{ConnectivityService: map[string]*ConnectivityService{"sample-cs": cs}},
	}

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
	pushErrors, err := s.SynchronizeDevice(device)
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
	jsonDataUpfSlice, err := ioutil.ReadFile("./testdata/sample-upfslice1.json")
	assert.NoError(t, err)
	jsonDataCoreSlice, err := ioutil.ReadFile("./testdata/sample-coreslice3.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	ent, cs, _, _, site, _ := BuildSampleDeviceGroup() // nolint dogsled
	_, _, _, slice := BuildSampleSlice(ent, site)      // nolint dogsled

	// Delete the one and only DeviceGroup
	slice.DeviceGroup = nil

	device := &RootDevice{
		Enterprises:          &models.OnfEnterprise_Enterprises{Enterprise: map[string]*Enterprise{"sample-ent": ent}},
		ConnectivityServices: &models.OnfConnectivityService_ConnectivityServices{ConnectivityService: map[string]*ConnectivityService{"sample-cs": cs}},
	}

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
	pushErrors, err := s.SynchronizeDevice(device)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)

	// TODO(smbaker): Behavior has changed between Aether-1.6 and Aether-2.0. We now push a
	// DG even when it is not related to a VCS. Is this a problem?
	_, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)

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
