// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package synchronizerv4

import (
	"github.com/golang/mock/gomock"
	models "github.com/onosproject/config-models/modelplugin/aether-4.0.0/aether_4_0_0"
	"github.com/onosproject/sdcore-adapter/pkg/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSynchronizeVcsUPF(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	m := NewMemPusher()
	s := Synchronizer{}
	s.SetPusher(mockPusher)

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

	jsonData := `{
			"sliceName": "sample-vcs",
			"sliceQos": {
				"uplinkMBR": 333,
				"downlinkMBR": 444,
				"bitrateUnit": "bps"
			},
			"ueResourceInfo": [
				{
					"uePoolId": "sample-dg",
					"dnn": "5ginternet"
				}
			]
		}`
	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		m.Pushes[endpoint] = jsonData
		return nil
	}).AnyTimes()

	pushFailures, err := s.SynchronizeVcsUPF(&device, vcs)
	assert.Nil(t, err)
	json, okay := m.Pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	assert.Equal(t, 0, pushFailures)
	if okay {
		require.JSONEq(t, jsonData, json)
	}
}

func TestSynchronizeVcsUPFNoSliceQos(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	m := NewMemPusher()
	s := Synchronizer{}
	s.SetPusher(mockPusher)

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

	jsonData := `{
			"sliceName": "sample-vcs",
			"sliceQos": {
			},
			"ueResourceInfo": [
				{
					"uePoolId": "sample-dg",
					"dnn": "5ginternet"
				}
			]
		}`
	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		m.Pushes[endpoint] = jsonData
		return nil
	}).AnyTimes()
	pushFailures, err := s.SynchronizeVcsUPF(&device, vcs)
	assert.Nil(t, err)
	t.Logf("%+v", m.Pushes)
	json, okay := m.Pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	assert.Equal(t, 0, pushFailures)
	if okay {
		require.JSONEq(t, jsonData, json)
	}
}
