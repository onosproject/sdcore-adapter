// SPDX-FileCopyrightText: 2022-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package steps

import (
	modelsv2 "github.com/onosproject/config-models/modelplugin/aether-2.0.0/aether_2_0_0"
	modelpluginv2 "github.com/onosproject/config-models/modelplugin/aether-2.0.0/modelplugin"
	modelsv4 "github.com/onosproject/config-models/modelplugin/aether-4.0.0/aether_4_0_0"
	modelpluginv4 "github.com/onosproject/config-models/modelplugin/aether-4.0.0/modelplugin"
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	"github.com/onosproject/sdcore-adapter/pkg/migration"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"reflect"
	"testing"
)

func Test_MigrateV4V2(t *testing.T) {
	srcJSON, err := ioutil.ReadFile("./testdata/mega_patch_400_src.json")
	assert.NoError(t, err)
	srcValJSON := &gpb.TypedValue{
		Value: &gpb.TypedValue_JsonVal{
			JsonVal: srcJSON,
		},
	}

	dstValJSON := &gpb.TypedValue{
		Value: &gpb.TypedValue_JsonVal{},
	}

	v4Models := gnmi.NewModel(modelpluginv4.ModelData,
		reflect.TypeOf((*modelsv4.Device)(nil)),
		modelsv4.SchemaTree["Device"],
		modelsv4.Unmarshal,
		//models.ΛEnum  // NOTE: There is no Enum in the aether models? So use a blank map.
		map[string]map[int64]ygot.EnumDefinition{},
	)

	v2Models := gnmi.NewModel(modelpluginv2.ModelData,
		reflect.TypeOf((*modelsv2.Device)(nil)),
		modelsv2.SchemaTree["Device"],
		modelsv2.Unmarshal,
		//models.ΛEnum  // NOTE: There is no Enum in the aether models? So use a blank map.
		map[string]map[int64]ygot.EnumDefinition{},
	)

	migrateV3V4Step := &migration.MigrationStep{
		FromVersion:   "4.0.0",
		FromModels:    v4Models,
		ToVersion:     "2.0.0",
		ToModels:      v2Models,
		MigrationFunc: MigrateV4V2,
		Migrator:      nil,
	}

	actions, err := MigrateV4V2(migrateV3V4Step, "cs4", "cs2", srcValJSON, dstValJSON)
	assert.NoError(t, err)
	assert.Len(t, actions, 49, "unexpected: actions is %d items", len(actions))

	// Expecting action 0-1 to be connectivity-service order is changeable
	for idx := 0; idx <= 1; idx++ {
		csAction := actions[idx]
		assert.Empty(t, csAction.DeletePrefix)
		assert.Len(t, csAction.Deletes, 1)
		assert.Equal(t, "connectivity-services", csAction.UpdatePrefix.GetElem()[0].GetName())
		assert.Equal(t, "cs2", csAction.Updates[0].Path.Target)
	}

	// Expecting action 2-4 to be enterprise - order is changeable
	for idx := 2; idx <= 4; idx++ {
		enterpriseAction := actions[idx]
		assert.Empty(t, enterpriseAction.DeletePrefix)
		assert.Equal(t, "enterprises", enterpriseAction.UpdatePrefix.GetElem()[0].GetName(),
			"unexpected type for %d", idx)

		entname, ok := enterpriseAction.UpdatePrefix.Elem[1].Key["enterprise-id"]
		assert.True(t, ok)
		switch entname {
		case "acme", "defaultent", "starbucks":
			assert.Equal(t, "description", enterpriseAction.Updates[0].Path.GetElem()[0].Name)
			if entname == "acme" {
				assert.Len(t, enterpriseAction.Deletes, 1)
				assert.Len(t, enterpriseAction.Updates, 3)
				assert.Equal(t, "ACME Corporation", enterpriseAction.Updates[0].Val.GetStringVal())
				assert.Equal(t, "ACME Corp", enterpriseAction.Updates[1].Val.GetStringVal())
				assert.Equal(t, "connectivity-service", enterpriseAction.Updates[2].Path.GetElem()[0].Name)
				cs, ok := enterpriseAction.Updates[2].Path.GetElem()[0].Key["connectivity-service"]
				assert.True(t, ok)
				assert.Equal(t, "cs5gtest", cs)
				assert.Equal(t, len(entname)%2 == 0, enterpriseAction.Updates[2].Val.GetBoolVal(),
					"expected connectivity-service for %s to be true", entname)
			} else if entname == "defaultent" {
				assert.Len(t, enterpriseAction.Deletes, 1)
				assert.Len(t, enterpriseAction.Updates, 2)
				assert.Equal(t, "This Enterprise holds discovered IMSIs that cannot be associated elsewhere.",
					enterpriseAction.Updates[0].Val.GetStringVal())
				assert.Equal(t, "Default Enterprise", enterpriseAction.Updates[1].Val.GetStringVal())
			} else if entname == "starbucks" {
				assert.Len(t, enterpriseAction.Deletes, 1)
				assert.Len(t, enterpriseAction.Updates, 4)
				assert.Equal(t, "Starbucks Corporation", enterpriseAction.Updates[0].Val.GetStringVal())
				assert.Equal(t, "Starbucks Inc.", enterpriseAction.Updates[1].Val.GetStringVal())
				assert.Equal(t, "connectivity-service", enterpriseAction.Updates[2].Path.GetElem()[0].Name)
			}
			assert.Equal(t, "display-name", enterpriseAction.Updates[1].Path.GetElem()[0].Name)

		default:
			t.Errorf("unexpected enterprise: %s", entname)
		}
	}

	//Expecting actions 5-7 to be Applications - order is changeable
	for idx := 5; idx <= 7; idx++ {
		applicationAction := actions[idx]
		assert.Empty(t, applicationAction.DeletePrefix)
		assert.Equal(t, "application", applicationAction.UpdatePrefix.GetElem()[2].GetName(),
			"unexpected type for %d", idx)
		assert.Len(t, applicationAction.Deletes, 1)
		assert.Len(t, applicationAction.Updates, 10)
		appID, ok := applicationAction.UpdatePrefix.GetElem()[2].GetKey()["application-id"]
		assert.True(t, ok)
		switch appID {
		case "acme-dataacquisition":
			assert.Equal(t, "Data Acquisition", applicationAction.Updates[0].Val.GetStringVal())
			assert.Equal(t, "data acquisition endpoint", applicationAction.Updates[3].Val.GetStringVal())
			assert.Equal(t, "da", applicationAction.Updates[3].GetPath().GetElem()[0].Key["endpoint-id"])
		case "starbucks-fidelio":
			assert.Equal(t, "Fidelio", applicationAction.Updates[1].Val.GetStringVal())
			assert.Equal(t, uint64(2000000), applicationAction.Updates[7].Val.GetUintVal())
			assert.Equal(t, "fidelio", applicationAction.Updates[3].GetPath().GetElem()[0].Key["endpoint-id"])
		case "starbucks-nvr":
			assert.Equal(t, "NVR", applicationAction.Updates[1].Val.GetStringVal())
			assert.Equal(t, "endpoint", applicationAction.Updates[4].GetPath().GetElem()[0].GetName())
			assert.Equal(t, "cs2", applicationAction.Updates[2].GetPath().Target)
		default:
			t.Errorf("Unexpected Application Profile ID %s for %d", appID, idx)

		}
	}

	// Expecting actions 8-16 to be Traffic-Class - order is changeable
	for idx := 14; idx <= 16; idx++ {
		trafficClassAction := actions[idx]
		assert.Empty(t, trafficClassAction.DeletePrefix)
		assert.Equal(t, "traffic-class", trafficClassAction.UpdatePrefix.GetElem()[2].GetName(),
			"unexpected type for %d", idx)
		assert.Len(t, trafficClassAction.Deletes, 0)
		assert.Len(t, trafficClassAction.Updates, 6)

		tcID, ok := trafficClassAction.UpdatePrefix.GetElem()[2].GetKey()["traffic-class-id"]
		assert.True(t, ok)
		switch tcID {
		case "class-1":
			assert.Equal(t, "High Priority TC", trafficClassAction.Updates[0].Val.GetStringVal())
			assert.Equal(t, uint64(10), trafficClassAction.Updates[2].Val.GetUintVal())
			assert.Equal(t, "pdb", trafficClassAction.Updates[4].GetPath().GetElem()[0].GetName())
		case "class-2":
			assert.Equal(t, "Class 2", trafficClassAction.Updates[1].Val.GetStringVal())
			assert.Equal(t, uint64(20), trafficClassAction.Updates[2].Val.GetUintVal())
			assert.Equal(t, "pelr", trafficClassAction.Updates[3].GetPath().GetElem()[0].GetName())
		case "class-3":
			assert.Equal(t, "Low Priority TC", trafficClassAction.Updates[0].Val.GetStringVal())
			assert.Equal(t, "cs2", trafficClassAction.Updates[2].GetPath().GetTarget())
			assert.Equal(t, uint64(100), trafficClassAction.Updates[4].Val.GetUintVal())
		default:
			t.Errorf("Unexpected Traffic Class Profile ID %s for %d", tcID, idx)
		}
	}

	// Expecting actions 17-22 Template - order is changeable
	for idx := 17; idx <= 22; idx++ {
		templateAction := actions[idx]
		assert.Empty(t, templateAction.DeletePrefix)
		assert.Equal(t, "template", templateAction.UpdatePrefix.GetElem()[2].GetName(), "unexpected type for %d", idx)
		assert.Len(t, templateAction.Deletes, 1)
		templateID, ok := templateAction.UpdatePrefix.GetElem()[2].GetKey()["template-id"]
		assert.True(t, ok)
		switch templateID {
		case "template-1":
			assert.Len(t, templateAction.Updates, 9)
			assert.Equal(t, "VCS Template 1", templateAction.Updates[0].Val.GetStringVal())
			assert.Equal(t, "default-behavior", templateAction.Updates[4].GetPath().GetElem()[0].GetName())
			assert.Equal(t, "mbr", templateAction.Updates[5].GetPath().GetElem()[1].GetName())
		case "template-2":
			assert.Len(t, templateAction.Updates, 8)
			assert.Equal(t, "slice", templateAction.Updates[6].GetPath().GetElem()[0].GetName())
			assert.Equal(t, "DENY-ALL", templateAction.Updates[4].Val.GetStringVal())
			assert.Equal(t, "sst", templateAction.Updates[2].GetPath().GetElem()[0].GetName())
		default:
			t.Errorf("Unexpected Site Profile ID %s for %d", templateID, idx)
		}
	}

	// Expecting actions 23-26 Site - order is changeable
	for idx := 23; idx <= 26; idx++ {
		siteAction := actions[idx]
		assert.Empty(t, siteAction.DeletePrefix)
		assert.Equal(t, "site", siteAction.UpdatePrefix.GetElem()[2].GetName(), "unexpected type for %d", idx)
		assert.Len(t, siteAction.Deletes, 1)
		siteID, ok := siteAction.UpdatePrefix.GetElem()[2].GetKey()["site-id"]
		assert.True(t, ok)
		switch siteID {
		case "acme-chicago":
			assert.Len(t, siteAction.Updates, 16)
			assert.Equal(t, "cell number one", siteAction.Updates[2].Val.GetStringVal())
			assert.Equal(t, "123", siteAction.Updates[12].Val.GetStringVal())
			assert.Equal(t, "edge-device", siteAction.Updates[11].GetPath().GetElem()[1].GetName())
		case "defaultent-defaultsite":
			assert.Len(t, siteAction.Updates, 6)
			assert.Equal(t, "SSSSSSSSSSSSSSS", siteAction.Updates[4].Val.GetStringVal())
			assert.Equal(t, "Global Default Site", siteAction.Updates[0].Val.GetStringVal())
			assert.Equal(t, "imsi-definition", siteAction.Updates[2].GetPath().GetElem()[0].GetName())
		case "starbucks-newyork":
			assert.Len(t, siteAction.Updates, 14)
			assert.Equal(t, "ap2.newyork.starbucks.com", siteAction.Updates[3].Val.GetStringVal())
			assert.Equal(t, "mcc", siteAction.Updates[10].GetPath().GetElem()[1].GetName())
			assert.Equal(t, "starbucks-newyork-monitoring-pi-1", siteAction.Updates[9].GetPath().GetElem()[1].Key["edge-device-id"])
		case "starbucks-seattle":
			assert.Len(t, siteAction.Updates, 18)
			assert.Equal(t, "tac", siteAction.Updates[4].GetPath().GetElem()[1].GetName())
			assert.Equal(t, "prometheus-amp", siteAction.Updates[11].Val.GetStringVal())
			assert.Equal(t, uint64(2), siteAction.Updates[17].Val.GetUintVal())
		default:
			t.Errorf("Unexpected Site Profile ID %s for %d", siteID, idx)
		}
	}

	// Expecting actions 27-33 to be Upf - order is changeable
	for idx := 27; idx <= 33; idx++ {
		upfAction := actions[idx]
		assert.Empty(t, upfAction.DeletePrefix)
		assert.Equal(t, "upf", upfAction.UpdatePrefix.GetElem()[3].GetName(), "expected upf for %d", idx)
		upfID, ok := upfAction.UpdatePrefix.GetElem()[3].GetKey()["upf-id"]
		assert.Len(t, upfAction.Deletes, 1)
		assert.True(t, ok)

		switch upfID {
		case "acme-chicago-pool-entry1":
			assert.Len(t, upfAction.Updates, 4)
			assert.Equal(t, "Chicago Pool 1", upfAction.Updates[1].Val.GetStringVal())
			assert.Equal(t, "address", upfAction.Updates[2].GetPath().GetElem()[0].GetName())
			assert.Equal(t, uint64(6161), upfAction.Updates[3].Val.GetUintVal())
		case "acme-chicago-pool-entry2":
			assert.Len(t, upfAction.Updates, 4)
			assert.Equal(t, "Chicago UPF Pool - Entry 2", upfAction.Updates[0].Val.GetStringVal())
			assert.Equal(t, "entry2.upfpool.chicago.acme.com", upfAction.Updates[2].Val.GetStringVal())
			assert.Equal(t, "port", upfAction.Updates[3].GetPath().GetElem()[0].GetName())
		case "starbucks-newyork-pool-entry1":
			assert.Len(t, upfAction.Updates, 4)
			assert.Equal(t, "New York Pool 1", upfAction.Updates[1].Val.GetStringVal())
			assert.Equal(t, "address", upfAction.Updates[2].GetPath().GetElem()[0].GetName())
			assert.Equal(t, uint64(6161), upfAction.Updates[3].Val.GetUintVal())
		case "starbucks-newyork-pool-entry2":
			assert.Len(t, upfAction.Updates, 4)
			assert.Equal(t, "New York UPF Pool - Entry 2", upfAction.Updates[0].Val.GetStringVal())
			assert.Equal(t, "cs2", upfAction.Updates[1].GetPath().GetTarget())
			assert.Equal(t, "address", upfAction.Updates[2].GetPath().GetElem()[0].GetName())
		case "starbucks-newyork-pool-entry3":
			assert.Len(t, upfAction.Updates, 4)
			assert.Equal(t, "New York Pool 3", upfAction.Updates[1].Val.GetStringVal())
			assert.Equal(t, "display-name", upfAction.Updates[1].GetPath().GetElem()[0].GetName())
			assert.Equal(t, "entry3.upfpool.newyork.starbucks.com", upfAction.Updates[2].Val.GetStringVal())
		case "starbucks-seattle-pool-entry1":
			assert.Len(t, upfAction.Updates, 5)
			assert.Equal(t, "http://entry1-seattle", upfAction.Updates[4].Val.GetStringVal())
			assert.Equal(t, "display-name", upfAction.Updates[1].GetPath().GetElem()[0].GetName())
			assert.Equal(t, "cs2", upfAction.Updates[2].Path.Target)
		case "starbucks-seattle-pool-entry2":
			assert.Len(t, upfAction.Updates, 5)
			assert.Equal(t, "entry2.upfpool.seattle.starbucks.com", upfAction.Updates[2].Val.GetStringVal())
			assert.Equal(t, uint64(9229), upfAction.Updates[3].Val.GetUintVal())
			assert.Equal(t, "address", upfAction.Updates[2].GetPath().GetElem()[0].GetName())
		default:
			t.Errorf("Unexpected UPF Profile ID %s", upfID)
		}
	}

	// Expecting actions 34-36 to be Slice - order is changeable
	for idx := 34; idx <= 36; idx++ {
		sliceAction := actions[idx]
		assert.Empty(t, sliceAction.DeletePrefix)
		assert.Equal(t, "slice", sliceAction.UpdatePrefix.GetElem()[3].GetName(), "expected upf for %d", idx)
		sliceID, ok := sliceAction.UpdatePrefix.GetElem()[3].GetKey()["slice-id"]
		assert.Len(t, sliceAction.Deletes, 1)
		assert.True(t, ok)

		switch sliceID {
		case "starbucks-newyork-cameras":
			assert.Len(t, sliceAction.Updates, 12)
			assert.Equal(t, "NY Cams", sliceAction.Updates[1].Val.GetStringVal())
			assert.Equal(t, "filter", sliceAction.Updates[7].GetPath().GetElem()[0].GetName())
			assert.Equal(t, uint64(600000), sliceAction.Updates[10].Val.GetUintVal())
		case "starbucks-seattle-cameras":
			assert.Len(t, sliceAction.Updates, 10)
			assert.Equal(t, "Seattle Cameras", sliceAction.Updates[0].Val.GetStringVal())
			assert.Equal(t, "starbucks-seattle-cameras", sliceAction.Updates[6].GetPath().GetElem()[0].Key["device-group"])
			assert.Equal(t, false, sliceAction.Updates[7].Val.GetBoolVal())
			assert.Equal(t, "allow", sliceAction.Updates[7].GetPath().GetElem()[1].GetName())
		case "acme-chicago-robots":
			assert.Len(t, sliceAction.Updates, 10)
			assert.Equal(t, "Chicago Robots VCS", sliceAction.Updates[1].Val.GetStringVal())
			assert.Equal(t, "acme-dataacquisition", sliceAction.Updates[7].GetPath().GetElem()[0].Key["application"])
			assert.Equal(t, uint64(5000000), sliceAction.Updates[8].Val.GetUintVal())
		default:
			t.Errorf("Unexpected UPF Profile ID %s", sliceID)
		}
	}

	//Expecting actions 37-39 to be Device Group - order is changeable
	for idx := 37; idx <= 39; idx++ {
		deviceGrouproupAction := actions[idx]
		assert.Empty(t, deviceGrouproupAction.DeletePrefix)
		assert.Equal(t, "device-group", deviceGrouproupAction.UpdatePrefix.GetElem()[3].GetName(),
			"unexpected type for %d", idx)
		dgID, ok := deviceGrouproupAction.UpdatePrefix.Elem[3].Key["device-group-id"]
		assert.Len(t, deviceGrouproupAction.Deletes, 1)
		assert.True(t, ok)
		switch dgID {
		case "starbucks-newyork-cameras":
			assert.Len(t, deviceGrouproupAction.Updates, 12)
			assert.Equal(t, "New York Cameras", deviceGrouproupAction.Updates[0].Val.GetStringVal())
			assert.Equal(t, "class-1", deviceGrouproupAction.Updates[3].Val.GetStringVal())
		case "starbucks-seattle-cameras":
			assert.Len(t, deviceGrouproupAction.Updates, 14)
			assert.Equal(t, "starbucks-seattle", deviceGrouproupAction.Updates[1].Val.GetStringVal())
			assert.Equal(t, "enable", deviceGrouproupAction.Updates[12].GetPath().GetElem()[1].GetName())
			assert.Equal(t, true, deviceGrouproupAction.Updates[11].Val.GetBoolVal())
		case "acme-chicago-robots":
			assert.Len(t, deviceGrouproupAction.Updates, 12)
			assert.Equal(t, "ACME Robots", deviceGrouproupAction.Updates[0].Val.GetStringVal())
			assert.Equal(t, "downlink", deviceGrouproupAction.Updates[3].GetPath().GetElem()[1].GetName())
			assert.Equal(t, true, deviceGrouproupAction.Updates[5].Val.GetBoolVal())

		default:
			t.Errorf("unexpected device group: %s", dgID)
		}
	}

	//Expecting actions 20-23 to be IP Domains - order is changeable
	for idx := 40; idx <= 42; idx++ {
		ipDomainAction := actions[idx]
		assert.Empty(t, ipDomainAction.DeletePrefix)
		assert.Equal(t, "ip-domain", ipDomainAction.UpdatePrefix.GetElem()[3].GetName(), "unexpected type for %d", idx)
		assert.Len(t, ipDomainAction.Deletes, 1)
		assert.Len(t, ipDomainAction.Updates, 8)
		ipdID, ok := ipDomainAction.UpdatePrefix.GetElem()[3].GetKey()["ip-domain-id"]
		assert.True(t, ok)
		switch ipdID {
		case "acme-chicago":
			assert.Equal(t, "Chicago", ipDomainAction.Updates[1].Val.GetStringVal())
			assert.Equal(t, "8.8.8.4", ipDomainAction.Updates[4].Val.GetStringVal())
			assert.Equal(t, "subnet", ipDomainAction.Updates[6].GetPath().GetElem()[0].GetName())
		case "starbucks-newyork":
			assert.Equal(t, "New York", ipDomainAction.Updates[1].Val.GetStringVal())
			assert.Equal(t, "254.186.117.251/31", ipDomainAction.Updates[6].Val.GetStringVal())
			assert.Equal(t, "dns-secondary", ipDomainAction.Updates[4].GetPath().GetElem()[0].GetName())
		case "starbucks-seattle":
			assert.Equal(t, "ENABLE", ipDomainAction.Updates[7].Val.GetStringVal())
			assert.Equal(t, "dns-primary", ipDomainAction.Updates[3].GetPath().GetElem()[0].GetName())
			assert.Equal(t, uint64(12690), ipDomainAction.Updates[5].Val.GetUintVal())
		default:
			t.Errorf("Unexpected ip-domain Profile ID %s", ipdID)
		}
	}

	//Expecting actions 43-48 to be Device and Sim from Imsi-Range - order is changeable
	for idx := 43; idx <= 48; idx++ {
		deviceSimAction := actions[idx]
		assert.Empty(t, deviceSimAction.DeletePrefix)
		siteID, ok := deviceSimAction.UpdatePrefix.GetElem()[2].Key["site-id"]
		assert.True(t, ok)
		switch siteID {
		case "acme-chicago":
			if len(deviceSimAction.Updates) == 16 {
				assert.Equal(t, "production-0", deviceSimAction.Updates[0].GetPath().GetElem()[0].Key["device-id"])
				assert.Equal(t, uint64(123456001000001), deviceSimAction.Updates[7].Val.GetUintVal())
				assert.Equal(t, "Sim production 3", deviceSimAction.Updates[14].Val.GetStringVal())
				assert.Equal(t, "sim-production-2", deviceSimAction.Updates[10].GetPath().GetElem()[0].Key["sim-id"])
			} else if len(deviceSimAction.Updates) == 12 {
				assert.Equal(t, "Sim warehouse 10", deviceSimAction.Updates[2].Val.GetStringVal())
				assert.Equal(t, "imsi", deviceSimAction.Updates[7].GetPath().GetElem()[1].GetName())
				assert.Equal(t, uint64(123456001000011), deviceSimAction.Updates[7].Val.GetUintVal())
				assert.Equal(t, "sim-card", deviceSimAction.Updates[11].GetPath().GetElem()[0].GetName())
			}
		case "starbucks-newyork":
			if len(deviceSimAction.Updates) == 24 {
				assert.Equal(t, uint64(21032002000052), deviceSimAction.Updates[11].Val.GetUintVal())
				assert.Equal(t, "store 50", deviceSimAction.Updates[0].Val.GetStringVal())
				assert.Equal(t, "sim-store-53", deviceSimAction.Updates[14].GetPath().GetElem()[0].Key["sim-id"])
				assert.Equal(t, "sim-card", deviceSimAction.Updates[3].GetPath().GetElem()[0].GetName())
			} else if len(deviceSimAction.Updates) == 8 {
				assert.Equal(t, "sim-front-40", deviceSimAction.Updates[1].Val.GetStringVal())
				assert.Equal(t, uint64(21032002000041), deviceSimAction.Updates[7].Val.GetUintVal())
				assert.Equal(t, "sim-front-41", deviceSimAction.Updates[6].GetPath().GetElem()[0].Key["sim-id"])
				assert.Equal(t, "display-name", deviceSimAction.Updates[6].GetPath().GetElem()[1].GetName())
			}
		case "starbucks-seattle":
			if len(deviceSimAction.Updates) == 16 {
				assert.Equal(t, uint64(265122002000001), deviceSimAction.Updates[7].Val.GetUintVal())
				assert.Equal(t, "counters-2", deviceSimAction.Updates[8].GetPath().GetElem()[0].Key["device-id"])
				assert.Equal(t, uint64(265122002000003), deviceSimAction.Updates[15].Val.GetUintVal())
				assert.Equal(t, "imsi", deviceSimAction.Updates[7].GetPath().GetElem()[1].GetName())
			} else if len(deviceSimAction.Updates) == 20 {
				assert.Equal(t, "Sim store 13", deviceSimAction.Updates[14].Val.GetStringVal())
				assert.Equal(t, uint64(265122002000010), deviceSimAction.Updates[3].Val.GetUintVal())
				assert.Equal(t, "store-12", deviceSimAction.Updates[8].GetPath().GetElem()[0].Key["device-id"])
				assert.Equal(t, "sim-card", deviceSimAction.Updates[19].GetPath().GetElem()[0].GetName())
			}

		default:
			t.Errorf("Unexpected site Profile ID %s", siteID)

		}
	}
}
