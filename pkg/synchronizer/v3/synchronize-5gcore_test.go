// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package synchronizerv3

import (
	models_v3 "github.com/onosproject/config-models/modelplugin/aether-3.0.0/aether_3_0_0"
	"github.com/stretchr/testify/assert"
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

// populate an Enterprise structure
func MakeEnterprise(desc string, displayName string, id string, cs []string) *models_v3.Enterprise_Enterprise_Enterprise {
	csList := map[string]*models_v3.Enterprise_Enterprise_Enterprise_ConnectivityService{}

	for _, csId := range cs {
		csList[csId] = &models_v3.Enterprise_Enterprise_Enterprise_ConnectivityService{
			ConnectivityService: aStr(csId),
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
		os.Remove(tempFileName)
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

func TestSynchronizeDeviceDeviceGroup(t *testing.T) {
	ent := MakeEnterprise("sample-ent-desc", "sample-ent-dn", "sample-ent", []string{"sample-cs"})
	cs := MakeCs("sample-cs-desc", "sample-cs-dn", "sample-cs")

	ipd := &models_v3.IpDomain_IpDomain_IpDomain{
		Description: aStr("sample-ipd-desc"),
		DisplayName: aStr("sample-ipd-dn"),
		Id:          aStr("sample-ipd"),
		Subnet:      aStr("1.2.3.4/24"),
		DnsPrimary:  aStr("8.8.8.8"),
		Mtu:         aUint32(1492),
	}
	site := &models_v3.Site_Site_Site{
		Description: aStr("sample-site-desc"),
		DisplayName: aStr("sample-site-dn"),
		Id:          aStr("sample-site"),
		Enterprise:  aStr("sample-ent"),
	}
	dg := &models_v3.DeviceGroup_DeviceGroup_DeviceGroup{
		//Description: aStr("sample-dg-desc"),
		DisplayName: aStr("sample-dg-dn"),
		Id:          aStr("sample-dg"),
		Site:        aStr("sample-site"),
		IpDomain:    aStr("sample-ipd"),
	}

	m := NewMemPusher()
	s := Synchronizer{}
	s.SetPusher(m)

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
		expected_result := `{
  "imsis": null,
  "ip-domain-name": "sample-ipd",
  "site-info": "sample-site",
  "ip-domain-expanded": {
    "dnn": "Internet",
    "ue-ip-pool": "1.2.3.4/24",
    "dns-primary": "8.8.8.8",
    "mtu": 1492
  }
}`
		assert.Equal(t, expected_result, json)
	}
}
