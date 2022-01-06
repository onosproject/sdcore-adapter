// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

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

func TestSynchronizeVcsUPF(t *testing.T) {
	jsonData, err := ioutil.ReadFile("./testdata/sample-slice.json")
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

	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(jsonData)
		return nil
	}).AnyTimes()
	pushFailures, err := s.SynchronizeVcsUPF(&device, vcs)
	assert.Nil(t, err)
	json, okay := pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	assert.Equal(t, 0, pushFailures)
	if okay {
		require.JSONEq(t, string(jsonData), json)
	}
}

func TestSynchronizeVcsUPFWithBurst(t *testing.T) {
	jsonData, err := ioutil.ReadFile("./testdata/sample-slice.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	ent, cs, tcList, ipd, site, dg := BuildSampleDeviceGroup()
	apps, tp, upf, vcs := BuildSampleVcs()

	vcs.Slice.Mbr.UplinkBurstSize = aUint32(111222)
	vcs.Slice.Mbr.DownlinkBurstSize = aUint32(333444)

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

	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(jsonData)
		return nil
	}).AnyTimes()
	pushFailures, err := s.SynchronizeVcsUPF(&device, vcs)
	assert.Nil(t, err)
	json, okay := pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	assert.Equal(t, 0, pushFailures)
	if okay {
		require.JSONEq(t, string(jsonData), json)
	}
}

func TestSynchronizeVcsUPFNoSliceQos(t *testing.T) {
	jsonData, err := ioutil.ReadFile("./testdata/sample-slice1.json")
	assert.NoError(t, err)
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := NewSynchronizer(WithPusher(mockPusher))

	ent, cs, tcList, ipd, site, dg := BuildSampleDeviceGroup()
	apps, tp, upf, vcs := BuildSampleVcs()

	vcs.Slice = nil // remove the slice QoS

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

	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = string(jsonData)
		return nil
	}).AnyTimes()
	pushFailures, err := s.SynchronizeVcsUPF(&device, vcs)
	assert.Nil(t, err)
	json, okay := pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	assert.Equal(t, 0, pushFailures)
	if okay {
		require.JSONEq(t, string(jsonData), json)
	}
}
