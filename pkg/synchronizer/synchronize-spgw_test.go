// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package synchronizer

import (
	models_v2 "github.com/onosproject/config-models/modelplugin/aether-2.0.0/aether_2_0_0"
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

// to facilitate easy declaring of pointers to uint64
func aUint64(u uint64) *uint64 {
	return &u
}

// populate an Enterprise structure
func MakeEnterprise(desc string, displayName string, id string, cs []string) *models_v2.Enterprise_Enterprise_Enterprise {
	csList := map[string]*models_v2.Enterprise_Enterprise_Enterprise_ConnectivityService{}

	for _, csId := range cs {
		csList[csId] = &models_v2.Enterprise_Enterprise_Enterprise_ConnectivityService{
			ConnectivityService: aStr(csId),
			Enabled:             aBool(true),
		}
	}

	ent := models_v2.Enterprise_Enterprise_Enterprise{
		Description:         aStr(desc),
		DisplayName:         aStr(displayName),
		Id:                  aStr(id),
		ConnectivityService: csList,
	}

	return &ent
}

func MakeCs(desc string, displayName string, id string) *models_v2.ConnectivityService_ConnectivityService_ConnectivityService {
	cs := models_v2.ConnectivityService_ConnectivityService_ConnectivityService{
		Description: aStr(desc),
		DisplayName: aStr(displayName),
		Id:          aStr(id),
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
	device := models_v2.Device{}
	err = s.SynchronizeDevice(&device)
	assert.Nil(t, err)

	content, err := ioutil.ReadFile(tempFileName)
	assert.Nil(t, err)
	assert.Equal(t, "", string(content))
}

func TestSynchronizeDeviceEnterpriseAndConnectivityOnly(t *testing.T) {
	// Get a temporary file name and defer deletion of the file
	f, err := ioutil.TempFile("", "synchronizer-json")
	assert.Nil(t, err)
	tempFileName := f.Name()
	defer func() {
		os.Remove(tempFileName)
	}()

	ent := MakeEnterprise("sample-ent-desc", "sample-ent-dn", "sample-ent", []string{"sample-cs"})
	cs := MakeCs("sample-cs-desc", "sample-cs-dn", "sample-cs")

	s := Synchronizer{}
	s.SetOutputFileName(tempFileName)
	device := models_v2.Device{
		Enterprise:          &models_v2.Enterprise_Enterprise{Enterprise: map[string]*models_v2.Enterprise_Enterprise_Enterprise{"sample-ent": ent}},
		ConnectivityService: &models_v2.ConnectivityService_ConnectivityService{ConnectivityService: map[string]*models_v2.ConnectivityService_ConnectivityService_ConnectivityService{"sample-cs": cs}},
	}
	err = s.SynchronizeDevice(&device)
	assert.Nil(t, err)

	content, err := ioutil.ReadFile(tempFileName)
	assert.Nil(t, err)
	assert.Equal(t, "{}", string(content))
}

func TestSynchronizeDeviceConnectivityServiceNotFound(t *testing.T) {
	// Get a temporary file name and defer deletion of the file
	f, err := ioutil.TempFile("", "synchronizer-json")
	assert.Nil(t, err)
	tempFileName := f.Name()
	defer func() {
		os.Remove(tempFileName)
	}()

	ent := MakeEnterprise("sample-ent-desc", "sample-ent-dn", "sample-ent", []string{"cs-missing"})
	cs := MakeCs("sample-cs-desc", "sample-cs-dn", "sample-cs")

	s := Synchronizer{}
	s.SetOutputFileName(tempFileName)
	device := models_v2.Device{
		Enterprise:          &models_v2.Enterprise_Enterprise{Enterprise: map[string]*models_v2.Enterprise_Enterprise_Enterprise{"sample-ent": ent}},
		ConnectivityService: &models_v2.ConnectivityService_ConnectivityService{ConnectivityService: map[string]*models_v2.ConnectivityService_ConnectivityService_ConnectivityService{"sample-cs": cs}},
	}
	err = s.SynchronizeDevice(&device)
	assert.EqualError(t, err, "Failed to find connectivity service cs-missing")
}

// a fully populated device should yield a decent chunk of json
func TestSynchronizeDevicePopulated(t *testing.T) {
	// Get a temporary file name and defer deletion of the file
	f, err := ioutil.TempFile("", "synchronizer-json")
	assert.Nil(t, err)
	tempFileName := f.Name()
	defer func() {
		os.Remove(tempFileName)
	}()

	s := Synchronizer{}
	s.SetOutputFileName(tempFileName)

	ent := MakeEnterprise("sample-ent-desc", "sample-ent-dn", "sample-ent", []string{"sample-cs"})
	cs := MakeCs("sample-cs-desc", "sample-cs-dn", "sample-cs")

	acp := models_v2.AccessProfile_AccessProfile_AccessProfile{
		Description: aStr("sample-acp-desc"),
		DisplayName: aStr("sample-acp-dn"),
		Filter:      aStr("sample-acp-filter"),
		Id:          aStr("sample-acp"),
		Type:        aStr("sample-acp-type"),
	}

	app := models_v2.ApnProfile_ApnProfile_ApnProfile{
		ApnName:      aStr("sample-app-name"),
		Description:  aStr("sample-app-description"),
		DisplayName:  aStr("sample-app-displayname"),
		DnsPrimary:   aStr("sample-app-dnsprimary"),
		DnsSecondary: aStr("sample-app-dnssecondary"),
		Id:           aStr("sample-app"),
		GxEnabled:    aBool(true),
		Mtu:          aUint32(123),
	}

	qp := models_v2.QosProfile_QosProfile_QosProfile{
		ApnAmbr:     &models_v2.QosProfile_QosProfile_QosProfile_ApnAmbr{Downlink: aUint32(123), Uplink: aUint32(456)},
		Description: aStr("sample-qp-desc"),
		DisplayName: aStr("sample-qp-displayname"),
		Id:          aStr("sample-qp"),
	}

	up := models_v2.UpProfile_UpProfile_UpProfile{
		AccessControl: aStr("sample-up-ac"),
		Description:   aStr("sample-up-desc"),
		DisplayName:   aStr("sample-up-displayname"),
		Id:            aStr("sample-up"),
		UserPlane:     aStr("sample-up-up"),
	}

	ue := models_v2.AetherSubscriber_Subscriber_Ue{
		DisplayName:   aStr("sample-ue-displayname"),
		Enabled:       aBool(true),
		Id:            aStr("68a3b12e-253e-11eb-adc1-0242ac120002"),
		ImsiRangeFrom: aUint64(123456),
		ImsiRangeTo:   aUint64(1234567),
		Priority:      aUint32(1),
		Enterprise:    aStr("sample-ent"),
		Profiles: &models_v2.AetherSubscriber_Subscriber_Ue_Profiles{
			ApnProfile: aStr("sample-app"),
			QosProfile: aStr("sample-qp"),
			UpProfile:  aStr("sample-up"),
			AccessProfile: map[string]*models_v2.AetherSubscriber_Subscriber_Ue_Profiles_AccessProfile{
				"sample-acp": &models_v2.AetherSubscriber_Subscriber_Ue_Profiles_AccessProfile{
					AccessProfile: aStr("sample-acp"),
					Allowed:       aBool(true),
				},
			},
		},

		RequestedApn: aStr("sample-ue-req-apn"),
		ServingPlmn: &models_v2.AetherSubscriber_Subscriber_Ue_ServingPlmn{
			Mcc: aUint32(123),
			Mnc: aUint32(456),
			Tac: aUint32(789),
		},
	}

	device := models_v2.Device{
		Enterprise:          &models_v2.Enterprise_Enterprise{Enterprise: map[string]*models_v2.Enterprise_Enterprise_Enterprise{"sample-ent": ent}},
		ConnectivityService: &models_v2.ConnectivityService_ConnectivityService{ConnectivityService: map[string]*models_v2.ConnectivityService_ConnectivityService_ConnectivityService{"sample-cs": cs}},
		AccessProfile:       &models_v2.AccessProfile_AccessProfile{AccessProfile: map[string]*models_v2.AccessProfile_AccessProfile_AccessProfile{"sample-acp": &acp}},
		ApnProfile:          &models_v2.ApnProfile_ApnProfile{ApnProfile: map[string]*models_v2.ApnProfile_ApnProfile_ApnProfile{"sample-app": &app}},
		QosProfile:          &models_v2.QosProfile_QosProfile{QosProfile: map[string]*models_v2.QosProfile_QosProfile_QosProfile{"sample-qp": &qp}},
		Subscriber:          &models_v2.AetherSubscriber_Subscriber{Ue: map[string]*models_v2.AetherSubscriber_Subscriber_Ue{"68a3b12e-253e-11eb-adc1-0242ac120002": &ue}},
		UpProfile:           &models_v2.UpProfile_UpProfile{UpProfile: map[string]*models_v2.UpProfile_UpProfile_UpProfile{"sample-up": &up}},
	}

	err = s.SynchronizeDevice(&device)
	assert.Nil(t, err)

	// define the expected json here
	expected_result := `{
  "subscriber-selection-rules": [
    {
      "priority": 1,
      "keys": {
        "imsi-range": {
          "from": 123456,
          "to": 1234567
        },
        "serving-plmn": {
          "mcc": 123,
          "mnc": 456,
          "tac": 789
        },
        "requested-apn": "sample-ue-req-apn"
      },
      "selected-apn-profile": "sample-app",
      "selected-access-profile": [
        "sample-acp"
      ],
      "selected-qos-profile": "sample-qp",
      "selected-user-plane-profile": "sample-up"
    }
  ],
  "access-profiles": {
    "sample-acp": {
      "type": "sample-acp-type",
      "filter": "sample-acp-filter"
    }
  },
  "apn-profiles": {
    "sample-app": {
      "apn-name": "sample-app-name",
      "dns-primary": "sample-app-dnsprimary",
      "dns-secondary": "sample-app-dnssecondary",
      "mtu": 123,
      "gx-enabled": true,
      "network": "lbo",
      "usage": 1
    }
  },
  "qos-profiles": {
    "sample-qp": {
      "apn-ambr": [
        123,
        456
      ]
    }
  },
  "user-plane-profiles": {
    "sample-up": {
      "user-plane": "sample-up-up",
      "access-control": "sample-up-ac",
      "access-tags": {
        "tag1": "ACC"
      },
      "qos-tags": {
        "tag1": "BW"
      }
    }
  }
}`

	content, err := ioutil.ReadFile(tempFileName)
	assert.Nil(t, err)
	assert.Equal(t, expected_result, string(content))
}

// If the subscriber is not attached to the enterprise, then we should not output it
func TestSynchronizeDeviceUeNotInEnterprise(t *testing.T) {
	// Get a temporary file name and defer deletion of the file
	f, err := ioutil.TempFile("", "synchronizer-json")
	assert.Nil(t, err)
	tempFileName := f.Name()
	defer func() {
		os.Remove(tempFileName)
	}()

	s := Synchronizer{}
	s.SetOutputFileName(tempFileName)

	ent := MakeEnterprise("sample-ent-desc", "sample-ent-dn", "sample-ent", []string{"sample-cs"})
	cs := MakeCs("sample-cs-desc", "sample-cs-dn", "sample-cs")

	ue := models_v2.AetherSubscriber_Subscriber_Ue{
		DisplayName:   aStr("sample-ue-displayname"),
		Enabled:       aBool(true),
		Id:            aStr("68a3b12e-253e-11eb-adc1-0242ac120002"),
		ImsiRangeFrom: aUint64(123456),
		ImsiRangeTo:   aUint64(1234567),
		Priority:      aUint32(1),
		Enterprise:    aStr("wrong-ent"),
	}

	device := models_v2.Device{
		Enterprise:          &models_v2.Enterprise_Enterprise{Enterprise: map[string]*models_v2.Enterprise_Enterprise_Enterprise{"sample-ent": ent}},
		ConnectivityService: &models_v2.ConnectivityService_ConnectivityService{ConnectivityService: map[string]*models_v2.ConnectivityService_ConnectivityService_ConnectivityService{"sample-cs": cs}},
		Subscriber:          &models_v2.AetherSubscriber_Subscriber{Ue: map[string]*models_v2.AetherSubscriber_Subscriber_Ue{"68a3b12e-253e-11eb-adc1-0242ac120002": &ue}},
	}

	err = s.SynchronizeDevice(&device)
	assert.Nil(t, err)

	// define the expected json here
	expected_result := `{}`

	content, err := ioutil.ReadFile(tempFileName)
	assert.Nil(t, err)
	assert.Equal(t, expected_result, string(content))
}
