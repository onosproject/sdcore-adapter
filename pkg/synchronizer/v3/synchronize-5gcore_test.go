// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package synchronizerv3

import (
	models_v3 "github.com/onosproject/config-models/modelplugin/aether-3.0.0/aether_3_0_0"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

// to facilitate easy declaring of pointers to strings
func aStr(s string) *string {
	return &s
}

// to facilitate easy declaring of pointers to bools
func aBool(b bool) *bool {
	return &b
}

// to facilitate easy declaring of pointers to uint32
func aUint32(u uint32) *uint32 {
	return &u
}

// to facilitate easy declaring of pointers to uint64
func aUint64(u uint64) *uint64 {
	return &u
}

// populate an Enterprise structure
func MakeEnterprise(desc string, displayName string, id string, cs []string) *models_v3.Enterprise_Enterprise_Enterprise {
	csList := map[string]*models_v3.Enterprise_Enterprise_Enterprise_ConnectivityService{}

	for _, csID := range cs {
		csList[csID] = &models_v3.Enterprise_Enterprise_Enterprise_ConnectivityService{
			ConnectivityService: aStr(csID),
			Enabled:             aBool(true),
		}
	}

	ent := models_v3.Enterprise_Enterprise_Enterprise{
		Description:         aStr(desc),
		DisplayName:         aStr(displayName),
		Id:                  aStr(id),
		ConnectivityService: csList,
	}

	return &ent
}

func MakeCs(desc string, displayName string, id string) *models_v3.ConnectivityService_ConnectivityService_ConnectivityService {
	cs := models_v3.ConnectivityService_ConnectivityService_ConnectivityService{
		Description:     aStr(desc),
		DisplayName:     aStr(displayName),
		Id:              aStr(id),
		Core_5GEndpoint: aStr("http://5gcore"),
	}

	return &cs
}

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
	device := models_v3.Device{}
	err = s.SynchronizeDevice(&device)
	assert.Nil(t, err)

	content, err := ioutil.ReadFile(tempFileName)
	assert.Nil(t, err)
	assert.Equal(t, "", string(content))
}

func TestSynchronizeDeviceCSEnt(t *testing.T) {
	ent := MakeEnterprise("sample-ent-desc", "sample-ent-dn", "sample-ent", []string{"sample-cs"})
	cs := MakeCs("sample-cs-desc", "sample-cs-dn", "sample-cs")

	m := &MemPusher{}
	s := Synchronizer{}
	s.SetPusher(m)

	device := models_v3.Device{
		Enterprise:          &models_v3.Enterprise_Enterprise{Enterprise: map[string]*models_v3.Enterprise_Enterprise_Enterprise{"sample-ent": ent}},
		ConnectivityService: &models_v3.ConnectivityService_ConnectivityService{ConnectivityService: map[string]*models_v3.ConnectivityService_ConnectivityService_ConnectivityService{"sample-cs": cs}},
	}
	err := s.SynchronizeDevice(&device)
	assert.Nil(t, err)
}

func BuildSampleDeviceGroup() (
	*models_v3.Enterprise_Enterprise_Enterprise,
	*models_v3.ConnectivityService_ConnectivityService_ConnectivityService,
	*models_v3.IpDomain_IpDomain_IpDomain,
	*models_v3.Site_Site_Site,
	*models_v3.DeviceGroup_DeviceGroup_DeviceGroup) {
	ent := MakeEnterprise("sample-ent-desc", "sample-ent-dn", "sample-ent", []string{"sample-cs"})
	cs := MakeCs("sample-cs-desc", "sample-cs-dn", "sample-cs")

	ipd := &models_v3.IpDomain_IpDomain_IpDomain{
		Description: aStr("sample-ipd-desc"),
		DisplayName: aStr("sample-ipd-dn"),
		Id:          aStr("sample-ipd"),
		Subnet:      aStr("1.2.3.4/24"),
		DnsPrimary:  aStr("8.8.8.8"),
		Mtu:         aUint32(1492),
		Dnn:         aStr("5ginternet"),
	}
	imsiDef := &models_v3.Site_Site_Site_ImsiDefinition{
		Mcc:        aUint32(123),
		Mnc:        aUint32(456),
		Enterprise: aUint32(789),
		Format:     aStr("CCCNNNEEESSSSSS"),
	}
	site := &models_v3.Site_Site_Site{
		Description:    aStr("sample-site-desc"),
		DisplayName:    aStr("sample-site-dn"),
		Id:             aStr("sample-site"),
		Enterprise:     aStr("sample-ent"),
		ImsiDefinition: imsiDef,
	}
	imsi := models_v3.DeviceGroup_DeviceGroup_DeviceGroup_Imsis{
		ImsiRangeFrom: aUint64(1),
	}
	dg := &models_v3.DeviceGroup_DeviceGroup_DeviceGroup{
		//Description: aStr("sample-dg-desc"),
		DisplayName: aStr("sample-dg-dn"),
		Id:          aStr("sample-dg"),
		Site:        aStr("sample-site"),
		IpDomain:    aStr("sample-ipd"),
		Imsis:       map[string]*models_v3.DeviceGroup_DeviceGroup_DeviceGroup_Imsis{"sample-imsi": &imsi},
	}

	return ent, cs, ipd, site, dg
}

func BuildSampleVcs() (
	*models_v3.ApList_ApList_ApList,
	*models_v3.Application_Application_Application,
	*models_v3.Template_Template_Template,
	*models_v3.TrafficClass_TrafficClass_TrafficClass,
	*models_v3.Upf_Upf_Upf,
	*models_v3.Vcs_Vcs_Vcs) {

	ep := &models_v3.Application_Application_Application_Endpoint{
		Address:   aStr("1.2.3.4"),
		Name:      aStr("sample-app-ep"),
		PortStart: aUint32(123),
		PortEnd:   aUint32(124),
		Protocol:  aStr("UDP"),
	}

	ap := &models_v3.ApList_ApList_ApList_AccessPoints{
		Address: aStr("6.7.8.9"),
		Enable:  aBool(true),
		Tac:     aUint32(77),
	}

	apl := &models_v3.ApList_ApList_ApList{
		Id:           aStr("sample-aplist"),
		AccessPoints: map[string]*models_v3.ApList_ApList_ApList_AccessPoints{"sample-ap": ap},
		Description:  aStr("sample-aplist-desc"),
		DisplayName:  aStr("sample-aplist-dn"),
		Enterprise:   aStr("sample-ent"),
	}

	app := &models_v3.Application_Application_Application{
		Id:          aStr("sample-app"),
		Description: aStr("sample-app-desc"),
		DisplayName: aStr("sample-app-dn"),
		Endpoint:    map[string]*models_v3.Application_Application_Application_Endpoint{"sample-app-ep": ep},
		Enterprise:  aStr("sample-ent"),
	}

	appLink := &models_v3.Vcs_Vcs_Vcs_Application{
		Allow:       aBool(true),
		Application: aStr("sample-app"),
	}

	dgLink := &models_v3.Vcs_Vcs_Vcs_DeviceGroup{
		DeviceGroup: aStr("sample-dg"),
		Enable:      aBool(true),
	}

	tp := &models_v3.Template_Template_Template{
		Id:           aStr("sample-template"),
		Description:  aStr("sample-template-desc"),
		DisplayName:  aStr("sample-template-dn"),
		Downlink:     aUint32(4321),
		Uplink:       aUint32(8765),
		Sd:           aUint32(111),
		Sst:          aUint32(222),
		TrafficClass: aStr("sample-traffic-class"),
	}

	tc := &models_v3.TrafficClass_TrafficClass_TrafficClass{
		Id:          aStr("sample-traffic-class"),
		Description: aStr("sample-traffic-class-desc"),
		DisplayName: aStr("sample-traffic-class-dn"),
		Pdb:         aUint32(333),
		Pelr:        aUint32(444),
		Qci:         aUint32(55),
	}

	upf := &models_v3.Upf_Upf_Upf{
		Id:          aStr("sample-upf"),
		Address:     aStr("2.3.4.5"),
		Description: aStr("sample-upf-desc"),
		DisplayName: aStr("sample-upf-dn"),
		Port:        aUint32(66),
	}

	vcs := &models_v3.Vcs_Vcs_Vcs{
		Ap:           aStr("sample-aplist"),
		Application:  map[string]*models_v3.Vcs_Vcs_Vcs_Application{"sample-app": appLink},
		Description:  aStr("sample-vcs-desc"),
		DeviceGroup:  map[string]*models_v3.Vcs_Vcs_Vcs_DeviceGroup{"sample-dg": dgLink},
		DisplayName:  aStr("sample-app-dn"),
		Downlink:     aUint32(4321),
		Uplink:       aUint32(8765),
		Id:           aStr("sample-vcs"),
		Sd:           aUint32(111),
		Sst:          aUint32(222),
		Template:     aStr("sample-template"),
		TrafficClass: aStr("sample-traffic-class"),
		Upf:          aStr("sample-upf"),
	}

	return apl, app, tp, tc, upf, vcs
}

func TestSynchronizeDeviceDeviceGroup(t *testing.T) {

	m := NewMemPusher()
	s := Synchronizer{}
	s.SetPusher(m)

	ent, cs, ipd, site, dg := BuildSampleDeviceGroup()

	device := models_v3.Device{
		Enterprise:          &models_v3.Enterprise_Enterprise{Enterprise: map[string]*models_v3.Enterprise_Enterprise_Enterprise{"sample-ent": ent}},
		ConnectivityService: &models_v3.ConnectivityService_ConnectivityService{ConnectivityService: map[string]*models_v3.ConnectivityService_ConnectivityService_ConnectivityService{"sample-cs": cs}},
		Site:                &models_v3.Site_Site{Site: map[string]*models_v3.Site_Site_Site{"sample-site": site}},
		IpDomain:            &models_v3.IpDomain_IpDomain{IpDomain: map[string]*models_v3.IpDomain_IpDomain_IpDomain{"sample-ipd": ipd}},
		DeviceGroup:         &models_v3.DeviceGroup_DeviceGroup{DeviceGroup: map[string]*models_v3.DeviceGroup_DeviceGroup_DeviceGroup{"sample-dg": dg}},
	}
	err := s.SynchronizeDevice(&device)
	assert.Nil(t, err)

	json, okay := m.Pushes["http://5gcore/v1/device-group/sample-dg"]
	assert.True(t, okay)
	if okay {
		expectedResult := `{
			"imsis": [
			  "123456789000001"
			],
			"ip-domain-name": "sample-ipd",
			"site-info": "sample-site",
			"ip-domain-expanded": {
			  "dnn": "5ginternet",
			  "ue-ip-pool": "1.2.3.4/24",
			  "dns-primary": "8.8.8.8",
			  "mtu": 1492
			}
		  }`
		require.JSONEq(t, expectedResult, json)
	}
}

func TestSynchronizeVCS(t *testing.T) {
	m := NewMemPusher()
	s := Synchronizer{}
	s.SetPusher(m)

	ent, cs, ipd, site, dg := BuildSampleDeviceGroup()
	apl, app, tp, tc, upf, vcs := BuildSampleVcs()

	device := models_v3.Device{
		Enterprise:          &models_v3.Enterprise_Enterprise{Enterprise: map[string]*models_v3.Enterprise_Enterprise_Enterprise{"sample-ent": ent}},
		ConnectivityService: &models_v3.ConnectivityService_ConnectivityService{ConnectivityService: map[string]*models_v3.ConnectivityService_ConnectivityService_ConnectivityService{"sample-cs": cs}},
		Site:                &models_v3.Site_Site{Site: map[string]*models_v3.Site_Site_Site{"sample-site": site}},
		IpDomain:            &models_v3.IpDomain_IpDomain{IpDomain: map[string]*models_v3.IpDomain_IpDomain_IpDomain{"sample-ipd": ipd}},
		DeviceGroup:         &models_v3.DeviceGroup_DeviceGroup{DeviceGroup: map[string]*models_v3.DeviceGroup_DeviceGroup_DeviceGroup{*dg.Id: dg}},
		ApList:              &models_v3.ApList_ApList{ApList: map[string]*models_v3.ApList_ApList_ApList{*apl.Id: apl}},
		Application:         &models_v3.Application_Application{Application: map[string]*models_v3.Application_Application_Application{*app.Id: app}},
		Template:            &models_v3.Template_Template{Template: map[string]*models_v3.Template_Template_Template{*tp.Id: tp}},
		TrafficClass:        &models_v3.TrafficClass_TrafficClass{TrafficClass: map[string]*models_v3.TrafficClass_TrafficClass_TrafficClass{*tc.Id: tc}},
		Upf:                 &models_v3.Upf_Upf{Upf: map[string]*models_v3.Upf_Upf_Upf{*upf.Id: upf}},
		Vcs:                 &models_v3.Vcs_Vcs{Vcs: map[string]*models_v3.Vcs_Vcs_Vcs{*vcs.Id: vcs}},
	}

	err := s.SynchronizeDevice(&device)
	assert.Nil(t, err)
	json, okay := m.Pushes["http://5gcore/v1/network-slice/sample-vcs"]
	assert.True(t, okay)
	if okay {
		expectedResult := `{
			"slice-id": {
			  "sst": "222",
			  "sd": "00006F"
			},
			"qos": {
			  "uplink": 8765,
			  "downlink": 4321,
			  "traffic-class": "sample-traffic-class"
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
				  "tac": 77
				}
			  ],
			  "upf": {
				"upf-name": "2.3.4.5",
				"upf-port": 66
			  }
			},
			"deny-applications": [],
			"permit-applications": [
			  "sample-app"
			],
			"applications-information": [
			  {
				"app-name": "sample-app",
				"endpoint": "1.2.3.4/32",
				"start-port": 123,
				"end-port": 124,
				"protocol": 17
			  }
			]
		  }`

		require.JSONEq(t, expectedResult, json)
	}
}

func TestSynchronizeVCSEmptySD(t *testing.T) {
	m := NewMemPusher()
	s := Synchronizer{}
	s.SetPusher(m)

	ent, cs, ipd, site, dg := BuildSampleDeviceGroup()
	apl, app, tp, tc, upf, vcs := BuildSampleVcs()

	// Set the SD to nil.
	vcs.Sd = nil

	device := models_v3.Device{
		Enterprise:          &models_v3.Enterprise_Enterprise{Enterprise: map[string]*models_v3.Enterprise_Enterprise_Enterprise{"sample-ent": ent}},
		ConnectivityService: &models_v3.ConnectivityService_ConnectivityService{ConnectivityService: map[string]*models_v3.ConnectivityService_ConnectivityService_ConnectivityService{"sample-cs": cs}},
		Site:                &models_v3.Site_Site{Site: map[string]*models_v3.Site_Site_Site{"sample-site": site}},
		IpDomain:            &models_v3.IpDomain_IpDomain{IpDomain: map[string]*models_v3.IpDomain_IpDomain_IpDomain{"sample-ipd": ipd}},
		DeviceGroup:         &models_v3.DeviceGroup_DeviceGroup{DeviceGroup: map[string]*models_v3.DeviceGroup_DeviceGroup_DeviceGroup{*dg.Id: dg}},
		ApList:              &models_v3.ApList_ApList{ApList: map[string]*models_v3.ApList_ApList_ApList{*apl.Id: apl}},
		Application:         &models_v3.Application_Application{Application: map[string]*models_v3.Application_Application_Application{*app.Id: app}},
		Template:            &models_v3.Template_Template{Template: map[string]*models_v3.Template_Template_Template{*tp.Id: tp}},
		TrafficClass:        &models_v3.TrafficClass_TrafficClass{TrafficClass: map[string]*models_v3.TrafficClass_TrafficClass_TrafficClass{*tc.Id: tc}},
		Upf:                 &models_v3.Upf_Upf{Upf: map[string]*models_v3.Upf_Upf_Upf{*upf.Id: upf}},
		Vcs:                 &models_v3.Vcs_Vcs{Vcs: map[string]*models_v3.Vcs_Vcs_Vcs{*vcs.Id: vcs}},
	}

	err := s.SynchronizeDevice(&device)
	assert.Nil(t, err)
	json, okay := m.Pushes["http://5gcore/v1/network-slice/sample-vcs"]
	assert.True(t, okay)
	if okay {
		expectedResult := `{
			"slice-id": {
			  "sst": "222",
			  "sd": ""
			},
			"qos": {
			  "uplink": 8765,
			  "downlink": 4321,
			  "traffic-class": "sample-traffic-class"
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
				  "tac": 77
				}
			  ],
			  "upf": {
				"upf-name": "2.3.4.5",
				"upf-port": 66
			  }
			},
			"deny-applications": [],
			"permit-applications": [
			  "sample-app"
			],
			"applications-information": [
			  {
				"app-name": "sample-app",
				"endpoint": "1.2.3.4/32",
				"start-port": 123,
				"end-port": 124,
				"protocol": 17
			  }
			]
		  }`

		require.JSONEq(t, expectedResult, json)
	}
}
