// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package synchronizerv4

import (
	models "github.com/onosproject/config-models/modelplugin/aether-4.0.0/aether_4_0_0"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSynchronizeVcsUPF(t *testing.T) {
	m := NewMemPusher()
	s := Synchronizer{}
	s.SetPusher(m)

	ent, cs, ipd, site, dg := BuildSampleDeviceGroup()
	apps, tp, tc, upf, vcs := BuildSampleVcs()

	device := models.Device{
		Enterprise:          &models.OnfEnterprise_Enterprise{Enterprise: map[string]*models.OnfEnterprise_Enterprise_Enterprise{"sample-ent": ent}},
		ConnectivityService: &models.OnfConnectivityService_ConnectivityService{ConnectivityService: map[string]*models.OnfConnectivityService_ConnectivityService_ConnectivityService{"sample-cs": cs}},
		Site:                &models.OnfSite_Site{Site: map[string]*models.OnfSite_Site_Site{"sample-site": site}},
		IpDomain:            &models.OnfIpDomain_IpDomain{IpDomain: map[string]*models.OnfIpDomain_IpDomain_IpDomain{"sample-ipd": ipd}},
		DeviceGroup:         &models.OnfDeviceGroup_DeviceGroup{DeviceGroup: map[string]*models.OnfDeviceGroup_DeviceGroup_DeviceGroup{*dg.Id: dg}},
		Application:         &models.OnfApplication_Application{Application: apps},
		Template:            &models.OnfTemplate_Template{Template: map[string]*models.OnfTemplate_Template_Template{*tp.Id: tp}},
		TrafficClass:        &models.OnfTrafficClass_TrafficClass{TrafficClass: map[string]*models.OnfTrafficClass_TrafficClass_TrafficClass{*tc.Id: tc}},
		Upf:                 &models.OnfUpf_Upf{Upf: map[string]*models.OnfUpf_Upf_Upf{*upf.Id: upf}},
		Vcs:                 &models.OnfVcs_Vcs{Vcs: map[string]*models.OnfVcs_Vcs_Vcs{*vcs.Id: vcs}},
	}

	pushFailures, err := s.SynchronizeVcsUPF(&device, vcs)
	assert.Nil(t, err)
	t.Logf("%+v", m.Pushes)
	json, okay := m.Pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	assert.Equal(t, 0, pushFailures)
	if okay {
		expectedResult := `{
			"sliceName": "sample-vcs",
			"sliceQos": {
				"uplinkMBR": 333,
				"downlinkMBR": 444
			},
			"ueResourceInfo": [
				{
					"uePoolId": "sample-dg",
					"dnn": "5ginternet"
				}
			]
		}`

		require.JSONEq(t, expectedResult, json)
	}
}
