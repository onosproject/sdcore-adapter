// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package synchronizer

import (
	"fmt"
	"github.com/golang/mock/gomock"
	models "github.com/onosproject/config-models/modelplugin/aether-4.0.0/aether_4_0_0"
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

	device := models.Device{
		Enterprise:          &models.OnfEnterprise_Enterprise{Enterprise: map[string]*models.OnfEnterprise_Enterprise_Enterprise{"sample-ent": ent}},
		ConnectivityService: &models.OnfConnectivityService_ConnectivityService{ConnectivityService: map[string]*models.OnfConnectivityService_ConnectivityService_ConnectivityService{"sample-cs": cs}},
	}
	pushErrors, err := s.SynchronizeDevice(&device)
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

	ent, cs, tcList, ipd, site, dg := BuildSampleDeviceGroup()

	device := models.Device{
		Enterprise:          &models.OnfEnterprise_Enterprise{Enterprise: map[string]*models.OnfEnterprise_Enterprise_Enterprise{"sample-ent": ent}},
		ConnectivityService: &models.OnfConnectivityService_ConnectivityService{ConnectivityService: map[string]*models.OnfConnectivityService_ConnectivityService_ConnectivityService{"sample-cs": cs}},
		Site:                &models.OnfSite_Site{Site: map[string]*models.OnfSite_Site_Site{"sample-site": site}},
		IpDomain:            &models.OnfIpDomain_IpDomain{IpDomain: map[string]*models.OnfIpDomain_IpDomain_IpDomain{"sample-ipd": ipd}},
		DeviceGroup:         &models.OnfDeviceGroup_DeviceGroup{DeviceGroup: map[string]*models.OnfDeviceGroup_DeviceGroup_DeviceGroup{"sample-dg": dg}},
		TrafficClass:        &models.OnfTrafficClass_TrafficClass{TrafficClass: tcList},
	}

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	pushErrors, err := s.SynchronizeDevice(&device)
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

	ent, cs, tcList, ipd, site, dg := BuildSampleDeviceGroup()

	tcList["sample-traffic-class"].Pelr = aInt8(3)
	tcList["sample-traffic-class"].Pdb = aUint16(400)

	device := models.Device{
		Enterprise:          &models.OnfEnterprise_Enterprise{Enterprise: map[string]*models.OnfEnterprise_Enterprise_Enterprise{"sample-ent": ent}},
		ConnectivityService: &models.OnfConnectivityService_ConnectivityService{ConnectivityService: map[string]*models.OnfConnectivityService_ConnectivityService_ConnectivityService{"sample-cs": cs}},
		Site:                &models.OnfSite_Site{Site: map[string]*models.OnfSite_Site_Site{"sample-site": site}},
		IpDomain:            &models.OnfIpDomain_IpDomain{IpDomain: map[string]*models.OnfIpDomain_IpDomain_IpDomain{"sample-ipd": ipd}},
		DeviceGroup:         &models.OnfDeviceGroup_DeviceGroup{DeviceGroup: map[string]*models.OnfDeviceGroup_DeviceGroup_DeviceGroup{"sample-dg": dg}},
		TrafficClass:        &models.OnfTrafficClass_TrafficClass{TrafficClass: tcList},
	}

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()

	pushErrors, err := s.SynchronizeDevice(&device)
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

	ent, cs, tcList, ipd, site, dg := BuildSampleDeviceGroup()

	dg.Device.TrafficClass = nil

	device := models.Device{
		Enterprise:          &models.OnfEnterprise_Enterprise{Enterprise: map[string]*models.OnfEnterprise_Enterprise_Enterprise{"sample-ent": ent}},
		ConnectivityService: &models.OnfConnectivityService_ConnectivityService{ConnectivityService: map[string]*models.OnfConnectivityService_ConnectivityService_ConnectivityService{"sample-cs": cs}},
		Site:                &models.OnfSite_Site{Site: map[string]*models.OnfSite_Site_Site{"sample-site": site}},
		IpDomain:            &models.OnfIpDomain_IpDomain{IpDomain: map[string]*models.OnfIpDomain_IpDomain_IpDomain{"sample-ipd": ipd}},
		DeviceGroup:         &models.OnfDeviceGroup_DeviceGroup{DeviceGroup: map[string]*models.OnfDeviceGroup_DeviceGroup_DeviceGroup{"sample-dg": dg}},
		TrafficClass:        &models.OnfTrafficClass_TrafficClass{TrafficClass: tcList},
	}
	pushErrors, err := s.SynchronizeDevice(&device)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)

	// The above will fail synchronization with a nonfatal error because TrafficClass is missing

	_, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.False(t, okay)
}

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

	device := models.Device{
		Enterprise:          &models.OnfEnterprise_Enterprise{Enterprise: map[string]*models.OnfEnterprise_Enterprise_Enterprise{"sample-ent": ent}},
		ConnectivityService: &models.OnfConnectivityService_ConnectivityService{ConnectivityService: map[string]*models.OnfConnectivityService_ConnectivityService_ConnectivityService{"sample-cs": cs}},
		Site:                &models.OnfSite_Site{Site: map[string]*models.OnfSite_Site_Site{"sample-site": site}},
		IpDomain:            &models.OnfIpDomain_IpDomain{IpDomain: map[string]*models.OnfIpDomain_IpDomain_IpDomain{"sample-ipd": ipd}},
		DeviceGroup:         &models.OnfDeviceGroup_DeviceGroup{DeviceGroup: map[string]*models.OnfDeviceGroup_DeviceGroup_DeviceGroup{"sample-dg": dg}},
		TrafficClass:        &models.OnfTrafficClass_TrafficClass{TrafficClass: tcList},
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
		pushErrors, err := s.SynchronizeDevice(&device)
		assert.Equal(t, 0, pushErrors)
		assert.Nil(t, err)
	}

	assert.Equal(t, 10, len(pushes))

	// All 10 pushes should have identical text
	for i := 1; i < 10; i++ {
		assert.Equal(t, pushes[0], pushes[i])
	}
}

func TestSynchronizeDeviceDeviceGroupLinkedToVCS(t *testing.T) {

	jsonDataDg, err := ioutil.ReadFile("./testdata/sample-dg.json")
	assert.NoError(t, err)
	jsonDataVcs, err := ioutil.ReadFile("./testdata/sample-vcs.json")
	assert.NoError(t, err)
	jsonDataSlice, err := ioutil.ReadFile("./testdata/sample-slice.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	ent, cs, tcList, ipd, site, dg := BuildSampleDeviceGroup()
	apps, tp, upf, vcs := BuildSampleVcs()

	device := models.Device{
		Enterprise:          &models.OnfEnterprise_Enterprise{Enterprise: map[string]*models.OnfEnterprise_Enterprise_Enterprise{"sample-ent": ent}},
		ConnectivityService: &models.OnfConnectivityService_ConnectivityService{ConnectivityService: map[string]*models.OnfConnectivityService_ConnectivityService_ConnectivityService{"sample-cs": cs}},
		Site:                &models.OnfSite_Site{Site: map[string]*models.OnfSite_Site_Site{"sample-site": site}},
		IpDomain:            &models.OnfIpDomain_IpDomain{IpDomain: map[string]*models.OnfIpDomain_IpDomain_IpDomain{"sample-ipd": ipd}},
		DeviceGroup:         &models.OnfDeviceGroup_DeviceGroup{DeviceGroup: map[string]*models.OnfDeviceGroup_DeviceGroup_DeviceGroup{"sample-dg": dg}},
		Application:         &models.OnfApplication_Application{Application: apps},
		Template:            &models.OnfTemplate_Template{Template: map[string]*models.OnfTemplate_Template_Template{*tp.Id: tp}},
		TrafficClass:        &models.OnfTrafficClass_TrafficClass{TrafficClass: tcList},
		Upf:                 &models.OnfUpf_Upf{Upf: map[string]*models.OnfUpf_Upf_Upf{*upf.Id: upf}},
		Vcs:                 &models.OnfVcs_Vcs{Vcs: map[string]*models.OnfVcs_Vcs_Vcs{*vcs.Id: vcs}},
	}

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/network-slice/sample-vcs", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()

	pushErrors, err := s.SynchronizeDevice(&device)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)

	// Note: With an associated VCS, we'll pick up the QoS settings
	json, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataDg), json)
	}
	json, okay = pushes["http://5gcore/v1/network-slice/sample-vcs"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataVcs), json)
	}
	json, okay = pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataSlice), json)
	}
}

func TestSynchronizeVCS(t *testing.T) {
	jsonDataDg, err := ioutil.ReadFile("./testdata/sample-dg.json")
	assert.NoError(t, err)
	jsonDataVcs, err := ioutil.ReadFile("./testdata/sample-vcs.json")
	assert.NoError(t, err)
	jsonDataSlice, err := ioutil.ReadFile("./testdata/sample-slice.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	ent, cs, tcList, ipd, site, dg := BuildSampleDeviceGroup()
	apps, tp, upf, vcs := BuildSampleVcs()

	device := models.Device{
		Enterprise:          &models.OnfEnterprise_Enterprise{Enterprise: map[string]*models.OnfEnterprise_Enterprise_Enterprise{"sample-ent": ent}},
		ConnectivityService: &models.OnfConnectivityService_ConnectivityService{ConnectivityService: map[string]*models.OnfConnectivityService_ConnectivityService_ConnectivityService{"sample-cs": cs}},
		Site:                &models.OnfSite_Site{Site: map[string]*models.OnfSite_Site_Site{"sample-site": site}},
		IpDomain:            &models.OnfIpDomain_IpDomain{IpDomain: map[string]*models.OnfIpDomain_IpDomain_IpDomain{"sample-ipd": ipd}},
		DeviceGroup:         &models.OnfDeviceGroup_DeviceGroup{DeviceGroup: map[string]*models.OnfDeviceGroup_DeviceGroup_DeviceGroup{*dg.Id: dg}},
		Application:         &models.OnfApplication_Application{Application: apps},
		Template:            &models.OnfTemplate_Template{Template: map[string]*models.OnfTemplate_Template_Template{*tp.Id: tp}},
		TrafficClass:        &models.OnfTrafficClass_TrafficClass{TrafficClass: tcList},
		Upf:                 &models.OnfUpf_Upf{Upf: map[string]*models.OnfUpf_Upf_Upf{*upf.Id: upf}},
		Vcs:                 &models.OnfVcs_Vcs{Vcs: map[string]*models.OnfVcs_Vcs_Vcs{*vcs.Id: vcs}},
	}

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/network-slice/sample-vcs", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	pushErrors, err := s.SynchronizeDevice(&device)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)

	json, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataDg), json)
	}
	json, okay = pushes["http://5gcore/v1/network-slice/sample-vcs"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataVcs), json)
	}
	json, okay = pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataSlice), json)
	}
}

func TestSynchronizeVCSAllowAll(t *testing.T) {
	jsonDataDg, err := ioutil.ReadFile("./testdata/sample-dg.json")
	assert.NoError(t, err)
	jsonDataVcs, err := ioutil.ReadFile("./testdata/sample-vcs-allow-all.json")
	assert.NoError(t, err)
	jsonDataSlice, err := ioutil.ReadFile("./testdata/sample-slice.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	ent, cs, tcList, ipd, site, dg := BuildSampleDeviceGroup()
	apps, tp, upf, vcs := BuildSampleVcs()

	vcs.DefaultBehavior = aStr("ALLOW-ALL")

	device := models.Device{
		Enterprise:          &models.OnfEnterprise_Enterprise{Enterprise: map[string]*models.OnfEnterprise_Enterprise_Enterprise{"sample-ent": ent}},
		ConnectivityService: &models.OnfConnectivityService_ConnectivityService{ConnectivityService: map[string]*models.OnfConnectivityService_ConnectivityService_ConnectivityService{"sample-cs": cs}},
		Site:                &models.OnfSite_Site{Site: map[string]*models.OnfSite_Site_Site{"sample-site": site}},
		IpDomain:            &models.OnfIpDomain_IpDomain{IpDomain: map[string]*models.OnfIpDomain_IpDomain_IpDomain{"sample-ipd": ipd}},
		DeviceGroup:         &models.OnfDeviceGroup_DeviceGroup{DeviceGroup: map[string]*models.OnfDeviceGroup_DeviceGroup_DeviceGroup{*dg.Id: dg}},
		Application:         &models.OnfApplication_Application{Application: apps},
		Template:            &models.OnfTemplate_Template{Template: map[string]*models.OnfTemplate_Template_Template{*tp.Id: tp}},
		TrafficClass:        &models.OnfTrafficClass_TrafficClass{TrafficClass: tcList},
		Upf:                 &models.OnfUpf_Upf{Upf: map[string]*models.OnfUpf_Upf_Upf{*upf.Id: upf}},
		Vcs:                 &models.OnfVcs_Vcs{Vcs: map[string]*models.OnfVcs_Vcs_Vcs{*vcs.Id: vcs}},
	}

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/network-slice/sample-vcs", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	pushErrors, err := s.SynchronizeDevice(&device)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)

	json, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataDg), json)
	}
	json, okay = pushes["http://5gcore/v1/network-slice/sample-vcs"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataVcs), json)
	}
	json, okay = pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataSlice), json)
	}
}

func TestSynchronizeVCSAllowPublic(t *testing.T) {
	jsonDataDg, err := ioutil.ReadFile("./testdata/sample-dg.json")
	assert.NoError(t, err)
	jsonDataVcs, err := ioutil.ReadFile("./testdata/sample-vcs-allow-public.json")
	assert.NoError(t, err)
	jsonDataSlice, err := ioutil.ReadFile("./testdata/sample-slice.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	ent, cs, tcList, ipd, site, dg := BuildSampleDeviceGroup()
	apps, tp, upf, vcs := BuildSampleVcs()

	vcs.DefaultBehavior = aStr("ALLOW-PUBLIC")

	device := models.Device{
		Enterprise:          &models.OnfEnterprise_Enterprise{Enterprise: map[string]*models.OnfEnterprise_Enterprise_Enterprise{"sample-ent": ent}},
		ConnectivityService: &models.OnfConnectivityService_ConnectivityService{ConnectivityService: map[string]*models.OnfConnectivityService_ConnectivityService_ConnectivityService{"sample-cs": cs}},
		Site:                &models.OnfSite_Site{Site: map[string]*models.OnfSite_Site_Site{"sample-site": site}},
		IpDomain:            &models.OnfIpDomain_IpDomain{IpDomain: map[string]*models.OnfIpDomain_IpDomain_IpDomain{"sample-ipd": ipd}},
		DeviceGroup:         &models.OnfDeviceGroup_DeviceGroup{DeviceGroup: map[string]*models.OnfDeviceGroup_DeviceGroup_DeviceGroup{*dg.Id: dg}},
		Application:         &models.OnfApplication_Application{Application: apps},
		Template:            &models.OnfTemplate_Template{Template: map[string]*models.OnfTemplate_Template_Template{*tp.Id: tp}},
		TrafficClass:        &models.OnfTrafficClass_TrafficClass{TrafficClass: tcList},
		Upf:                 &models.OnfUpf_Upf{Upf: map[string]*models.OnfUpf_Upf_Upf{*upf.Id: upf}},
		Vcs:                 &models.OnfVcs_Vcs{Vcs: map[string]*models.OnfVcs_Vcs_Vcs{*vcs.Id: vcs}},
	}

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/network-slice/sample-vcs", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	pushErrors, err := s.SynchronizeDevice(&device)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)

	json, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataDg), json)
	}
	json, okay = pushes["http://5gcore/v1/network-slice/sample-vcs"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataVcs), json)
	}
	json, okay = pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataSlice), json)
	}
}

func TestSynchronizeVCSTwoEnpoints(t *testing.T) {
	jsonDataDg, err := ioutil.ReadFile("./testdata/sample-dg.json")
	assert.NoError(t, err)
	jsonDataVcs, err := ioutil.ReadFile("./testdata/sample-vcs1.json")
	assert.NoError(t, err)
	jsonDataSlice, err := ioutil.ReadFile("./testdata/sample-slice.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	ent, cs, tcList, ipd, site, dg := BuildSampleDeviceGroup()
	apps, tp, upf, vcs := BuildSampleVcs()

	mbr3 := &models.OnfApplication_Application_Application_Endpoint_Mbr{
		Uplink:   aUint64(44332211),
		Downlink: aUint64(88776655),
	}

	ep3 := &models.OnfApplication_Application_Application_Endpoint{
		EndpointId:   aStr("zep3"),
		PortStart:    aUint16(5555),
		PortEnd:      aUint16(5556),
		Protocol:     aStr("UDP"),
		Mbr:          mbr3,
		TrafficClass: aStr("sample-traffic-class"),
	}

	apps["sample-app2"].Endpoint["zep3"] = ep3

	device := models.Device{
		Enterprise:          &models.OnfEnterprise_Enterprise{Enterprise: map[string]*models.OnfEnterprise_Enterprise_Enterprise{"sample-ent": ent}},
		ConnectivityService: &models.OnfConnectivityService_ConnectivityService{ConnectivityService: map[string]*models.OnfConnectivityService_ConnectivityService_ConnectivityService{"sample-cs": cs}},
		Site:                &models.OnfSite_Site{Site: map[string]*models.OnfSite_Site_Site{"sample-site": site}},
		IpDomain:            &models.OnfIpDomain_IpDomain{IpDomain: map[string]*models.OnfIpDomain_IpDomain_IpDomain{"sample-ipd": ipd}},
		DeviceGroup:         &models.OnfDeviceGroup_DeviceGroup{DeviceGroup: map[string]*models.OnfDeviceGroup_DeviceGroup_DeviceGroup{*dg.Id: dg}},
		Application:         &models.OnfApplication_Application{Application: apps},
		Template:            &models.OnfTemplate_Template{Template: map[string]*models.OnfTemplate_Template_Template{*tp.Id: tp}},
		TrafficClass:        &models.OnfTrafficClass_TrafficClass{TrafficClass: tcList},
		Upf:                 &models.OnfUpf_Upf{Upf: map[string]*models.OnfUpf_Upf_Upf{*upf.Id: upf}},
		Vcs:                 &models.OnfVcs_Vcs{Vcs: map[string]*models.OnfVcs_Vcs_Vcs{*vcs.Id: vcs}},
	}

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/network-slice/sample-vcs", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	pushErrors, err := s.SynchronizeDevice(&device)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)

	json, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataDg), json)
	}
	json, okay = pushes["http://5gcore/v1/network-slice/sample-vcs"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataVcs), json)
	}
	json, okay = pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataSlice), json)
	}
}

func TestSynchronizeVCSEmptySD(t *testing.T) {
	jsonDataDg, err := ioutil.ReadFile("./testdata/sample-dg.json")
	assert.NoError(t, err)
	jsonDataVcs, err := ioutil.ReadFile("./testdata/sample-vcs2.json")
	assert.NoError(t, err)
	jsonDataSlice, err := ioutil.ReadFile("./testdata/sample-slice.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	ent, cs, tcList, ipd, site, dg := BuildSampleDeviceGroup()
	apps, tp, upf, vcs := BuildSampleVcs()

	// Set the SD to nil.
	vcs.Sd = nil

	device := models.Device{
		Enterprise:          &models.OnfEnterprise_Enterprise{Enterprise: map[string]*models.OnfEnterprise_Enterprise_Enterprise{"sample-ent": ent}},
		ConnectivityService: &models.OnfConnectivityService_ConnectivityService{ConnectivityService: map[string]*models.OnfConnectivityService_ConnectivityService_ConnectivityService{"sample-cs": cs}},
		Site:                &models.OnfSite_Site{Site: map[string]*models.OnfSite_Site_Site{"sample-site": site}},
		IpDomain:            &models.OnfIpDomain_IpDomain{IpDomain: map[string]*models.OnfIpDomain_IpDomain_IpDomain{"sample-ipd": ipd}},
		DeviceGroup:         &models.OnfDeviceGroup_DeviceGroup{DeviceGroup: map[string]*models.OnfDeviceGroup_DeviceGroup_DeviceGroup{*dg.Id: dg}},
		Application:         &models.OnfApplication_Application{Application: apps},
		Template:            &models.OnfTemplate_Template{Template: map[string]*models.OnfTemplate_Template_Template{*tp.Id: tp}},
		TrafficClass:        &models.OnfTrafficClass_TrafficClass{TrafficClass: tcList},
		Upf:                 &models.OnfUpf_Upf{Upf: map[string]*models.OnfUpf_Upf_Upf{*upf.Id: upf}},
		Vcs:                 &models.OnfVcs_Vcs{Vcs: map[string]*models.OnfVcs_Vcs_Vcs{*vcs.Id: vcs}},
	}

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/network-slice/sample-vcs", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	pushErrors, err := s.SynchronizeDevice(&device)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)
	json, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataDg), json)
	}
	json, okay = pushes["http://5gcore/v1/network-slice/sample-vcs"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataVcs), json)
	}
	json, okay = pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataSlice), json)
	}
}

func TestSynchronizeVCSDisabledDG(t *testing.T) {
	jsonDataDg, err := ioutil.ReadFile("./testdata/sample-dg.json")
	assert.NoError(t, err)
	jsonDataVcs, err := ioutil.ReadFile("./testdata/sample-vcs3.json")
	assert.NoError(t, err)
	jsonDataSlice, err := ioutil.ReadFile("./testdata/sample-slice1.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	ent, cs, tcList, ipd, site, dg := BuildSampleDeviceGroup()
	apps, tp, upf, vcs := BuildSampleVcs()

	// Disable the one and only DeviceGroup
	vcs.DeviceGroup[*dg.Id].Enable = aBool(false)

	device := models.Device{
		Enterprise:          &models.OnfEnterprise_Enterprise{Enterprise: map[string]*models.OnfEnterprise_Enterprise_Enterprise{"sample-ent": ent}},
		ConnectivityService: &models.OnfConnectivityService_ConnectivityService{ConnectivityService: map[string]*models.OnfConnectivityService_ConnectivityService_ConnectivityService{"sample-cs": cs}},
		Site:                &models.OnfSite_Site{Site: map[string]*models.OnfSite_Site_Site{"sample-site": site}},
		IpDomain:            &models.OnfIpDomain_IpDomain{IpDomain: map[string]*models.OnfIpDomain_IpDomain_IpDomain{"sample-ipd": ipd}},
		DeviceGroup:         &models.OnfDeviceGroup_DeviceGroup{DeviceGroup: map[string]*models.OnfDeviceGroup_DeviceGroup_DeviceGroup{*dg.Id: dg}},
		Application:         &models.OnfApplication_Application{Application: apps},
		Template:            &models.OnfTemplate_Template{Template: map[string]*models.OnfTemplate_Template_Template{*tp.Id: tp}},
		TrafficClass:        &models.OnfTrafficClass_TrafficClass{TrafficClass: tcList},
		Upf:                 &models.OnfUpf_Upf{Upf: map[string]*models.OnfUpf_Upf_Upf{*upf.Id: upf}},
		Vcs:                 &models.OnfVcs_Vcs{Vcs: map[string]*models.OnfVcs_Vcs_Vcs{*vcs.Id: vcs}},
	}

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/network-slice/sample-vcs", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	pushErrors, err := s.SynchronizeDevice(&device)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)
	json, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataDg), json)
	}
	json, okay = pushes["http://5gcore/v1/network-slice/sample-vcs"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataVcs), json)
	}
	json, okay = pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataSlice), json)
	}
}

func TestSynchronizeVCSMissingDG(t *testing.T) {
	jsonDataVcs, err := ioutil.ReadFile("./testdata/sample-vcs3.json")
	assert.NoError(t, err)
	jsonDataSlice, err := ioutil.ReadFile("./testdata/sample-slice1.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	ent, cs, tcList, ipd, site, _ := BuildSampleDeviceGroup()
	apps, tp, upf, vcs := BuildSampleVcs()

	// Delete the one and only DeviceGroup
	vcs.DeviceGroup = nil

	device := models.Device{
		Enterprise:          &models.OnfEnterprise_Enterprise{Enterprise: map[string]*models.OnfEnterprise_Enterprise_Enterprise{"sample-ent": ent}},
		ConnectivityService: &models.OnfConnectivityService_ConnectivityService{ConnectivityService: map[string]*models.OnfConnectivityService_ConnectivityService_ConnectivityService{"sample-cs": cs}},
		Site:                &models.OnfSite_Site{Site: map[string]*models.OnfSite_Site_Site{"sample-site": site}},
		IpDomain:            &models.OnfIpDomain_IpDomain{IpDomain: map[string]*models.OnfIpDomain_IpDomain_IpDomain{"sample-ipd": ipd}},
		Application:         &models.OnfApplication_Application{Application: apps},
		Template:            &models.OnfTemplate_Template{Template: map[string]*models.OnfTemplate_Template_Template{*tp.Id: tp}},
		TrafficClass:        &models.OnfTrafficClass_TrafficClass{TrafficClass: tcList},
		Upf:                 &models.OnfUpf_Upf{Upf: map[string]*models.OnfUpf_Upf_Upf{*upf.Id: upf}},
		Vcs:                 &models.OnfVcs_Vcs{Vcs: map[string]*models.OnfVcs_Vcs_Vcs{*vcs.Id: vcs}},
	}

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/network-slice/sample-vcs", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(data)
		return nil
	}).AnyTimes()
	pushErrors, err := s.SynchronizeDevice(&device)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)
	_, okay := pushes["http://5gcore/v1/device-group/sample-dg"] // no DG in this test
	assert.False(t, okay)
	json, okay := pushes["http://5gcore/v1/network-slice/sample-vcs"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataVcs), json)
	}
	json, okay = pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, string(jsonDataSlice), json)
	}
}
