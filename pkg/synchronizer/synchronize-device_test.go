// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package synchronizer

import (
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

	s := Synchronizer{}
	s.SetOutputFileName(tempFileName)
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

	s := Synchronizer{}
	s.SetPusher(mockPusher)

	device := models.Device{
		Enterprise:          &models.OnfEnterprise_Enterprise{Enterprise: map[string]*models.OnfEnterprise_Enterprise_Enterprise{"sample-ent": ent}},
		ConnectivityService: &models.OnfConnectivityService_ConnectivityService{ConnectivityService: map[string]*models.OnfConnectivityService_ConnectivityService_ConnectivityService{"sample-cs": cs}},
	}
	pushErrors, err := s.SynchronizeDevice(&device)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)
}

func TestSynchronizeDeviceDeviceGroupWithQos(t *testing.T) {

	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := Synchronizer{}
	s.SetPusher(mockPusher)

	ent, cs, tcList, ipd, site, dg := BuildSampleDeviceGroup()

	device := models.Device{
		Enterprise:          &models.OnfEnterprise_Enterprise{Enterprise: map[string]*models.OnfEnterprise_Enterprise_Enterprise{"sample-ent": ent}},
		ConnectivityService: &models.OnfConnectivityService_ConnectivityService{ConnectivityService: map[string]*models.OnfConnectivityService_ConnectivityService_ConnectivityService{"sample-cs": cs}},
		Site:                &models.OnfSite_Site{Site: map[string]*models.OnfSite_Site_Site{"sample-site": site}},
		IpDomain:            &models.OnfIpDomain_IpDomain{IpDomain: map[string]*models.OnfIpDomain_IpDomain_IpDomain{"sample-ipd": ipd}},
		DeviceGroup:         &models.OnfDeviceGroup_DeviceGroup{DeviceGroup: map[string]*models.OnfDeviceGroup_DeviceGroup_DeviceGroup{"sample-dg": dg}},
		TrafficClass:        &models.OnfTrafficClass_TrafficClass{TrafficClass: tcList},
	}

	jsonData := `{
			"imsis": [
			  "123456789000001"
			],
			"ip-domain-name": "sample-ipd",
			"site-info": "sample-site",
			"ip-domain-expanded": {
			  "dnn": "5ginternet",
			  "ue-ip-pool": "1.2.3.4/24",
			  "dns-primary": "8.8.8.8",
				"mtu": 1492,
				"ue-dnn-qos": {
					"dnn-mbr-downlink": 4321,
					"dnn-mbr-uplink": 8765,
					"bitrate-unit": "bps",
					"traffic-class": {
						"name": "sample-traffic-class",
						"arp": 3,
						"pdb": 300,
						"pelr": 6,
						"qci": 55
					}					
				}				
			}
		  }`
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = jsonData
		return nil
	}).AnyTimes()
	pushErrors, err := s.SynchronizeDevice(&device)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)

	json, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, jsonData, json)
	}
}

func TestSynchronizeDeviceDeviceGroupWithQosSpecifiedPelrPDB(t *testing.T) {

	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := Synchronizer{}
	s.SetPusher(mockPusher)

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

	jsonDataDg := `{
			"imsis": [
			  "123456789000001"
			],
			"ip-domain-name": "sample-ipd",
			"site-info": "sample-site",
			"ip-domain-expanded": {
			  "dnn": "5ginternet",
			  "ue-ip-pool": "1.2.3.4/24",
			  "dns-primary": "8.8.8.8",
				"mtu": 1492,
				"ue-dnn-qos": {
					"dnn-mbr-downlink": 4321,
					"dnn-mbr-uplink": 8765,
					"bitrate-unit": "bps",
					"traffic-class": {
						"name": "sample-traffic-class",
						"arp": 3,
						"pdb": 400,
						"pelr": 3,
						"qci": 55
					}					
				}				
			}
		  }`

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = jsonDataDg
		return nil
	}).AnyTimes()

	pushErrors, err := s.SynchronizeDevice(&device)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)

	json, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, jsonDataDg, json)
	}
}

func TestSynchronizeDeviceDeviceGroupWithQosButNoTC(t *testing.T) {

	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := Synchronizer{}
	s.SetPusher(mockPusher)

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

func TestSynchronizeDeviceDeviceGroupLinkedToVCS(t *testing.T) {

	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := Synchronizer{}
	s.SetPusher(mockPusher)

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

	jsonDataDg := `{
			"imsis": [
			  "123456789000001"
			],
			"ip-domain-name": "sample-ipd",
			"site-info": "sample-site",
			"ip-domain-expanded": {
			  "dnn": "5ginternet",
			  "ue-ip-pool": "1.2.3.4/24",
			  "dns-primary": "8.8.8.8",
				"mtu": 1492,
				"ue-dnn-qos": {
					"dnn-mbr-downlink": 4321,
					"dnn-mbr-uplink": 8765,
					"bitrate-unit": "bps",
					"traffic-class": {
						"name": "sample-traffic-class",
						"arp": 3,
						"pdb": 300,
						"pelr": 6,
						"qci": 55
					}
				}
			}
		  }`

	jsonDataVcs := `{
          "slice-id": {
            "sst": "222",
            "sd": "00006F"
          },
          "site-device-group": [
            "sample-dg"
          ],
          "site-info": {
            "site-name": "sample-site",
            "plmn": {
              "mcc": "123",
              "mnc": "456"
            },
            "gNodeBs": [
              {
                "name": "6.7.8.9",
                "tac": 30635
              }
            ],
            "upf": {
              "upf-name": "2.3.4.5",
              "upf-port": 66
            }
          },
          "application-filtering-rules": [
            {
              "rule-name": "sample-app",
              "priority": 7,
              "action": "permit",
              "endpoint": "1.2.3.4/32",
              "dest-port-start": 123,
              "dest-port-end": 124,
              "protocol": 17
            },
            {
              "rule-name": "sample-app2",
              "priority": 8,
              "action": "deny",
              "endpoint": "1.2.3.5/32",
              "dest-port-start": 123,
              "dest-port-end": 124,
              "protocol": 17,
              "app-mbr-uplink": 11223344,
              "app-mbr-downlink": 55667788,
              "traffic-class": {
                "name": "sample-traffic-class",
                "qci": 55,
                "arp": 3,
                "pdb": 300,
                "pelr": 6
              }
            },
            {
              "rule-name": "DENY-ALL",
              "priority": 250,
              "action": "deny",
              "endpoint": "0.0.0.0/0"
            }
          ]
        }`

	jsonDataSlice := `{
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
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = jsonDataDg
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/network-slice/sample-vcs", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = jsonDataVcs
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = jsonDataSlice
		return nil
	}).AnyTimes()

	pushErrors, err := s.SynchronizeDevice(&device)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)

	// Note: With an associated VCS, we'll pick up the QoS settings
	json, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, jsonDataDg, json)
	}
	json, okay = pushes["http://5gcore/v1/network-slice/sample-vcs"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, jsonDataVcs, json)
	}
	json, okay = pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, jsonDataSlice, json)
	}
}

func TestSynchronizeVCS(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
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

	jsonDataDg := `{
			"imsis": [
			  "123456789000001"
			],
			"ip-domain-name": "sample-ipd",
			"site-info": "sample-site",
			"ip-domain-expanded": {
			  "dnn": "5ginternet",
			  "ue-ip-pool": "1.2.3.4/24",
			  "dns-primary": "8.8.8.8",
				"mtu": 1492,
				"ue-dnn-qos": {
					"dnn-mbr-downlink": 4321,
					"dnn-mbr-uplink": 8765,
					"bitrate-unit": "bps",
					"traffic-class": {
						"name": "sample-traffic-class",
						"arp": 3,
						"pdb": 300,
						"pelr": 6,
						"qci": 55
					}
				}
			}
		  }`

	jsonDataVcs := `{
			"slice-id": {
			  "sst": "222",
			  "sd": "00006F"
			},
			"site-device-group": [
			  "sample-dg"
			],
			"site-info": {
			  "site-name": "sample-site",
			  "plmn": {
				"mcc": "123",
				"mnc": "456"
			  },
			  "gNodeBs": [
				{
				  "name": "6.7.8.9",
				  "tac": 30635
				}
			  ],
			  "upf": {
				"upf-name": "2.3.4.5",
				"upf-port": 66
			  }
			},
			"application-filtering-rules": [{
				"rule-name": "sample-app-sample-app-ep",
				"dest-port-start": 123,
				"dest-port-end": 124,
				"endpoint": "1.2.3.4/32",
				"action": "permit",
				"protocol": 17,
				"priority": 7
			},
			{
				"rule-name": "sample-app2-sample-app2-ep",
				"dest-port-start": 123,
				"dest-port-end": 124,
				"endpoint": "1.2.3.5/32",
				"action": "deny",
				"protocol": 17,
				"priority": 8,
				"app-mbr-downlink": 55667788,
				"app-mbr-uplink": 11223344,
				"bitrate-unit": "bps",
				"traffic-class": {
					"name": "sample-traffic-class",
					"arp": 3,
					"pdb": 300,
					"pelr": 6,
					"qci": 55
				}
			},
			{
				"rule-name": "DENY-ALL",
				"endpoint": "0.0.0.0/0",
				"priority": 250,
				"action": "deny"
			}]
		}`

	jsonDataSlice := `{
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
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = jsonDataDg
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/network-slice/sample-vcs", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = jsonDataVcs
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = jsonDataSlice
		return nil
	}).AnyTimes()
	pushErrors, err := s.SynchronizeDevice(&device)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)

	json, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, jsonDataDg, json)
	}
	json, okay = pushes["http://5gcore/v1/network-slice/sample-vcs"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, jsonDataVcs, json)
	}
	json, okay = pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, jsonDataSlice, json)
	}
}

func TestSynchronizeVCSTwoEnpoints(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := Synchronizer{}
	s.SetPusher(mockPusher)

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
	jsonDataDg := `{
			"imsis": [
			  "123456789000001"
			],
			"ip-domain-name": "sample-ipd",
			"site-info": "sample-site",
			"ip-domain-expanded": {
			  "dnn": "5ginternet",
			  "ue-ip-pool": "1.2.3.4/24",
			  "dns-primary": "8.8.8.8",
				"mtu": 1492,
				"ue-dnn-qos": {
					"dnn-mbr-downlink": 4321,
					"dnn-mbr-uplink": 8765,
					"bitrate-unit": "bps",
					"traffic-class": {
						"name": "sample-traffic-class",
						"arp": 3,
						"pdb": 300,
						"pelr": 6,
						"qci": 55
					}
				}
			}
		  }`

	jsonDataVcs := `{
			"slice-id": {
			  "sst": "222",
			  "sd": "00006F"
			},
			"site-device-group": [
			  "sample-dg"
			],
			"site-info": {
			  "site-name": "sample-site",
			  "plmn": {
				"mcc": "123",
				"mnc": "456"
			  },
			  "gNodeBs": [
				{
				  "name": "6.7.8.9",
				  "tac": 30635
				}
			  ],
			  "upf": {
				"upf-name": "2.3.4.5",
				"upf-port": 66
			  }
			},
			"application-filtering-rules": [{
				"rule-name": "sample-app-sample-app-ep",
				"dest-port-start": 123,
				"dest-port-end": 124,
				"endpoint": "1.2.3.4/32",
				"action": "permit",
				"protocol": 17,
				"priority": 7
			},
			{
				"rule-name": "sample-app2-sample-app2-ep",
				"dest-port-start": 123,
				"dest-port-end": 124,
				"endpoint": "1.2.3.5/32",
				"action": "deny",
				"protocol": 17,
				"priority": 8,
				"app-mbr-downlink": 55667788,
				"app-mbr-uplink": 11223344,
				"bitrate-unit": "bps",
				"traffic-class": {
					"name": "sample-traffic-class",
					"arp": 3,
					"pdb": 300,
					"pelr": 6,
					"qci": 55
				}
			},
			{
				"rule-name": "sample-app2-zep3",
				"dest-port-start": 5555,
				"dest-port-end": 5556,
				"endpoint": "1.2.3.5/32",
				"action": "deny",
				"protocol": 17,
				"priority": 8,
				"app-mbr-downlink": 88776655,
				"app-mbr-uplink": 44332211,
				"bitrate-unit": "bps",
				"traffic-class": {
					"name": "sample-traffic-class",
					"arp": 3,
					"pdb": 300,
					"pelr": 6,
					"qci": 55
				}
			},			
			{
				"rule-name": "DENY-ALL",
				"endpoint": "0.0.0.0/0",
				"priority": 250,
				"action": "deny"
			}]
		}`

	jsonDataSlice := `{
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
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = jsonDataDg
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/network-slice/sample-vcs", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = jsonDataVcs
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = jsonDataSlice
		return nil
	}).AnyTimes()
	pushErrors, err := s.SynchronizeDevice(&device)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)

	json, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, jsonDataDg, json)
	}
	json, okay = pushes["http://5gcore/v1/network-slice/sample-vcs"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, jsonDataVcs, json)
	}
	json, okay = pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, jsonDataSlice, json)
	}
}

func TestSynchronizeVCSEmptySD(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := Synchronizer{}
	s.SetPusher(mockPusher)

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

	jsonDataDg := `{
			"slice-id": {
			  "sst": "222",
			  "sd": ""
			},
			"site-device-group": [
			  "sample-dg"
			],
			"site-info": {
			  "site-name": "sample-site",
			  "plmn": {
				"mcc": "123",
				"mnc": "456"
			  },
			  "gNodeBs": [
				{
				  "name": "6.7.8.9",
				  "tac": 30635
				}
			  ],
			  "upf": {
				"upf-name": "2.3.4.5",
				"upf-port": 66
			  }
			},
			"application-filtering-rules": [{
				"rule-name": "sample-app-sample-app-ep",
				"dest-port-start": 123,
				"dest-port-end": 124,
				"endpoint": "1.2.3.4/32",
				"action": "permit",
				"protocol": 17,
				"priority": 7
			},
			{
				"rule-name": "sample-app2-sample-app2-ep",
				"dest-port-start": 123,
				"dest-port-end": 124,
				"endpoint": "1.2.3.5/32",
				"action": "deny",
				"protocol": 17,
				"priority": 8,
				"app-mbr-downlink": 55667788,
				"app-mbr-uplink": 11223344,
				"bitrate-unit": "bps",
				"traffic-class": {
					"name": "sample-traffic-class",
					"arp": 3,
					"pdb": 300,
					"pelr": 6,
					"qci": 55
				}
			},
			{
				"rule-name": "DENY-ALL",
				"endpoint": "0.0.0.0/0",
				"priority": 250,
				"action": "deny"
			}]
		}`

	jsonDataVcs := `{
			"slice-id": {
			  "sst": "222",
			  "sd": ""
			},
			"site-device-group": [
			  "sample-dg"
			],
			"site-info": {
			  "site-name": "sample-site",
			  "plmn": {
				"mcc": "123",
				"mnc": "456"
			  },
			  "gNodeBs": [
				{
				  "name": "6.7.8.9",
				  "tac": 30635
				}
			  ],
			  "upf": {
				"upf-name": "2.3.4.5",
				"upf-port": 66
			  }
			},
			"application-filtering-rules": [{
				"rule-name": "sample-app-sample-app-ep",
				"dest-port-start": 123,
				"dest-port-end": 124,
				"endpoint": "1.2.3.4/32",
				"action": "permit",
				"protocol": 17,
				"priority": 7
			},
			{
				"rule-name": "sample-app2-sample-app2-ep",
				"dest-port-start": 123,
				"dest-port-end": 124,
				"endpoint": "1.2.3.5/32",
				"action": "deny",
				"protocol": 17,
				"priority": 8,
				"app-mbr-downlink": 55667788,
				"app-mbr-uplink": 11223344,
				"bitrate-unit": "bps",
				"traffic-class": {
					"name": "sample-traffic-class",
					"arp": 3,
					"pdb": 300,
					"pelr": 6,
					"qci": 55
				}
			},
			{
				"rule-name": "DENY-ALL",
				"endpoint": "0.0.0.0/0",
				"priority": 250,
				"action": "deny"
			}]
		}`

	jsonDataSlice := `{
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
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = jsonDataDg
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/network-slice/sample-vcs", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = jsonDataVcs
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = jsonDataSlice
		return nil
	}).AnyTimes()
	pushErrors, err := s.SynchronizeDevice(&device)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)
	json, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, jsonDataDg, json)
	}
	json, okay = pushes["http://5gcore/v1/network-slice/sample-vcs"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, jsonDataVcs, json)
	}
	json, okay = pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, jsonDataSlice, json)
	}
}

func TestSynchronizeVCSDisabledDG(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := Synchronizer{}
	s.SetPusher(mockPusher)

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
	jsonDataDg := `{
			"slice-id": {
			  "sst": "222",
			  "sd": "00006F"
			},
			"site-info": {
			  "site-name": "sample-site",
			  "plmn": {
				"mcc": "123",
				"mnc": "456"
			  },
			  "gNodeBs": [
				{
				  "name": "6.7.8.9",
				  "tac": 30635
				}
			  ],
			  "upf": {
				"upf-name": "2.3.4.5",
				"upf-port": 66
			  }
			},
			"application-filtering-rules": [{
				"rule-name": "sample-app-sample-app-ep",
				"dest-port-start": 123,
				"dest-port-end": 124,
				"endpoint": "1.2.3.4/32",
				"action": "permit",
				"protocol": 17,
				"priority": 7
			},
			{
				"rule-name": "sample-app2-sample-app2-ep",
				"dest-port-start": 123,
				"dest-port-end": 124,
				"endpoint": "1.2.3.5/32",
				"action": "deny",
				"protocol": 17,
				"priority": 8,
				"app-mbr-downlink": 55667788,
				"app-mbr-uplink": 11223344,
				"bitrate-unit": "bps",
				"traffic-class": {
					"name": "sample-traffic-class",
					"arp": 3,
					"pdb": 300,
					"pelr": 6,
					"qci": 55
				}
			},
			{
				"rule-name": "DENY-ALL",
				"endpoint": "0.0.0.0/0",
				"priority": 250,
				"action": "deny"
			}]
		}`

	jsonDataVcs := `{
			"slice-id": {
			  "sst": "222",
			  "sd": "00006F"
			},
			"site-info": {
			  "site-name": "sample-site",
			  "plmn": {
				"mcc": "123",
				"mnc": "456"
			  },
			  "gNodeBs": [
				{
				  "name": "6.7.8.9",
				  "tac": 30635
				}
			  ],
			  "upf": {
				"upf-name": "2.3.4.5",
				"upf-port": 66
			  }
			},
			"application-filtering-rules": [{
				"rule-name": "sample-app-sample-app-ep",
				"dest-port-start": 123,
				"dest-port-end": 124,
				"endpoint": "1.2.3.4/32",
				"action": "permit",
				"protocol": 17,
				"priority": 7
			},
			{
				"rule-name": "sample-app2-sample-app2-ep",
				"dest-port-start": 123,
				"dest-port-end": 124,
				"endpoint": "1.2.3.5/32",
				"action": "deny",
				"protocol": 17,
				"priority": 8,
				"app-mbr-downlink": 55667788,
				"app-mbr-uplink": 11223344,
				"bitrate-unit": "bps",
				"traffic-class": {
					"name": "sample-traffic-class",
					"arp": 3,
					"pdb": 300,
					"pelr": 6,
					"qci": 55
				}
			},
			{
				"rule-name": "DENY-ALL",
				"endpoint": "0.0.0.0/0",
				"priority": 250,
				"action": "deny"
			}]
		}`

	jsonDataSlice := `{
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
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = jsonDataDg
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/network-slice/sample-vcs", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = jsonDataVcs
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = jsonDataSlice
		return nil
	}).AnyTimes()
	pushErrors, err := s.SynchronizeDevice(&device)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)
	json, okay := pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, jsonDataDg, json)
	}
	json, okay = pushes["http://5gcore/v1/network-slice/sample-vcs"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, jsonDataVcs, json)
	}
	json, okay = pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, jsonDataSlice, json)
	}
}

func TestSynchronizeVCSMissingDG(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := make(map[string]string)
	s := Synchronizer{}
	s.SetPusher(mockPusher)

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

	jsonDataVcs := `{
			"slice-id": {
			  "sst": "222",
			  "sd": "00006F"
			},
			"site-info": {
			  "site-name": "sample-site",
			  "plmn": {
				"mcc": "123",
				"mnc": "456"
			  },
			  "gNodeBs": [
				{
				  "name": "6.7.8.9",
				  "tac": 30635
				}
			  ],
			  "upf": {
				"upf-name": "2.3.4.5",
				"upf-port": 66
			  }
			},
			"application-filtering-rules": [{
				"rule-name": "sample-app-sample-app-ep",
				"dest-port-start": 123,
				"dest-port-end": 124,
				"endpoint": "1.2.3.4/32",
				"action": "permit",
				"protocol": 17,
				"priority": 7
			},
			{
				"rule-name": "sample-app2-sample-app2-ep",
				"dest-port-start": 123,
				"dest-port-end": 124,
				"endpoint": "1.2.3.5/32",
				"action": "deny",
				"protocol": 17,
				"priority": 8,
				"app-mbr-downlink": 55667788,
				"app-mbr-uplink": 11223344,
				"bitrate-unit": "bps",
				"traffic-class": {
					"name": "sample-traffic-class",
					"arp": 3,
					"pdb": 300,
					"pelr": 6,
					"qci": 55
				}
			},
			{
				"rule-name": "DENY-ALL",
				"endpoint": "0.0.0.0/0",
				"priority": 250,
				"action": "deny"
			}]
		}`

	jsonDataSlice := `{
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
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/network-slice/sample-vcs", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = jsonDataVcs
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes[endpoint] = jsonDataSlice
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
		require.JSONEq(t, jsonDataVcs, json)
	}
	json, okay = pushes["http://upf/v1/config/network-slices"]
	assert.True(t, okay)
	if okay {
		require.JSONEq(t, jsonDataSlice, json)
	}
}
