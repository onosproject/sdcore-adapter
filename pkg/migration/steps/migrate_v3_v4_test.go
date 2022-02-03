// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package steps

import (
	modelsv3 "github.com/onosproject/config-models/modelplugin/aether-3.0.0/aether_3_0_0"
	modelpluginv3 "github.com/onosproject/config-models/modelplugin/aether-3.0.0/modelplugin"
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

func Test_MigrateV3V4(t *testing.T) {
	srcJSON, err := ioutil.ReadFile("./testdata/mega_patch_300_src.json")
	assert.NoError(t, err)
	srcValJSON := &gpb.TypedValue{
		Value: &gpb.TypedValue_JsonVal{
			JsonVal: srcJSON,
		},
	}

	dstValJSON := &gpb.TypedValue{
		Value: &gpb.TypedValue_JsonVal{},
	}

	v3Models := gnmi.NewModel(modelpluginv3.ModelData,
		reflect.TypeOf((*modelsv3.Device)(nil)),
		modelsv3.SchemaTree["Device"],
		modelsv3.Unmarshal,
		//models.ΛEnum  // NOTE: There is no Enum in the aether models? So use a blank map.
		map[string]map[int64]ygot.EnumDefinition{},
	)

	v4Models := gnmi.NewModel(modelpluginv4.ModelData,
		reflect.TypeOf((*modelsv4.Device)(nil)),
		modelsv4.SchemaTree["Device"],
		modelsv4.Unmarshal,
		//models.ΛEnum  // NOTE: There is no Enum in the aether models? So use a blank map.
		map[string]map[int64]ygot.EnumDefinition{},
	)

	migrateV3V4Step := &migration.MigrationStep{
		FromVersion:   "3.0.0",
		FromModels:    v3Models,
		ToVersion:     "4.0.0",
		ToModels:      v4Models,
		MigrationFunc: MigrateV3V4,
		Migrator:      nil,
	}

	actions, err := MigrateV3V4(migrateV3V4Step, "cs3", "cs4", srcValJSON, dstValJSON)
	assert.NoError(t, err)
	assert.Len(t, actions, 49, "unexpected: actions is %d items", len(actions))

	// Expecting action 0-1 to be connectivity-service order is changeable
	for idx := 0; idx <= 1; idx++ {
		csAction := actions[idx]
		assert.Empty(t, csAction.DeletePrefix)
		assert.Len(t, csAction.Deletes, 1)
		assert.Equal(t, "connectivity-service", csAction.UpdatePrefix.GetElem()[0].GetName())
		assert.Equal(t, "cs4", csAction.Updates[0].Path.Target)
	}

	// Expecting action 2-4 to be enterprise - order is changeable
	for idx := 2; idx <= 4; idx++ {
		enterpriseAction := actions[idx]
		assert.Empty(t, enterpriseAction.DeletePrefix)
		assert.Equal(t, "enterprise", enterpriseAction.UpdatePrefix.GetElem()[0].GetName(),
			"unexpected type for %d", idx)

		entname, ok := enterpriseAction.UpdatePrefix.Elem[1].Key["id"]
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

	//Expecting actions 5-13 to be Device Group - order is changeable
	for idx := 5; idx <= 13; idx++ {
		deviceGrouproupAction := actions[idx]
		assert.Empty(t, deviceGrouproupAction.DeletePrefix)
		assert.Equal(t, "device-group", deviceGrouproupAction.UpdatePrefix.GetElem()[0].GetName(),
			"unexpected type for %d", idx)
		dgname, ok := deviceGrouproupAction.UpdatePrefix.Elem[1].Key["id"]
		assert.True(t, ok)
		assert.Equal(t, "display-name", deviceGrouproupAction.Updates[0].Path.GetElem()[0].Name)
		switch dgname {
		case "acme-chicago-default", "starbucks-newyork-default", "starbucks-newyork-pos", "starbucks-seattle-cameras",
			"acme-chicago-robots", "defaultent-defaultsite-default", "starbucks-newyork-cameras", "starbucks-seattle-default",
			"starbucks-seattle-pos":
			if dgname == "acme-chicago-default" {
				assert.Len(t, deviceGrouproupAction.Deletes, 1)
				assert.Len(t, deviceGrouproupAction.Updates, 3)
				assert.Equal(t, "ACME Default", deviceGrouproupAction.Updates[0].Val.GetStringVal())
				assert.Equal(t, "acme-chicago", deviceGrouproupAction.Updates[1].Val.GetStringVal())
				assert.Equal(t, "acme-chicago", deviceGrouproupAction.Updates[2].Val.GetStringVal())
			} else if dgname == "starbucks-newyork-default" {
				assert.Len(t, deviceGrouproupAction.Deletes, 1)
				assert.Len(t, deviceGrouproupAction.Updates, 3)
				assert.Equal(t, "New York Default", deviceGrouproupAction.Updates[0].Val.GetStringVal())
				assert.Equal(t, "starbucks-newyork", deviceGrouproupAction.Updates[1].Val.GetStringVal())
				assert.Equal(t, "ip-domain", deviceGrouproupAction.Updates[2].Path.GetElem()[0].Name)
			} else if dgname == "starbucks-newyork-pos" {
				assert.Len(t, deviceGrouproupAction.Deletes, 1)
				assert.Len(t, deviceGrouproupAction.Updates, 7)
				assert.Equal(t, "New York POS", deviceGrouproupAction.Updates[0].Val.GetStringVal())
				//assert.Equal(t, uint64(73), deviceGrouproupAction.Updates[2].GetVal().GetUintVal())
			} else if dgname == "starbucks-seattle-cameras" {
				assert.Len(t, deviceGrouproupAction.Deletes, 1)
				assert.Len(t, deviceGrouproupAction.Updates, 6)
				assert.Equal(t, "Seattle Cameras", deviceGrouproupAction.Updates[0].Val.GetStringVal())
				//assert.Equal(t, "counters", deviceGrouproupAction.Updates[1].Path.GetElem()[0].Key["name"])

			} else if dgname == "acme-chicago-robots" {
				assert.Len(t, deviceGrouproupAction.Deletes, 1)
				assert.Len(t, deviceGrouproupAction.Updates, 6)
				assert.Equal(t, "ACME Robots", deviceGrouproupAction.Updates[0].Val.GetStringVal())
				assert.Equal(t, "warehouse", deviceGrouproupAction.Updates[2].Path.GetElem()[0].Key["imsi-id"])
			} else if dgname == "defaultent-defaultsite-default" {
				assert.Len(t, deviceGrouproupAction.Deletes, 1)
				assert.Len(t, deviceGrouproupAction.Updates, 3)
				assert.Equal(t, "defaultent-defaultsite", deviceGrouproupAction.Updates[1].Val.GetStringVal())
				assert.Equal(t, "ip-domain", deviceGrouproupAction.Updates[2].Path.GetElem()[0].GetName())
				assert.Equal(t, "defaultent-defaultip", deviceGrouproupAction.Updates[2].Val.GetStringVal())

			} else if dgname == "starbucks-newyork-cameras" {
				assert.Len(t, deviceGrouproupAction.Deletes, 1)
				assert.Len(t, deviceGrouproupAction.Updates, 7)
				assert.Equal(t, "imsi-range-from", deviceGrouproupAction.Updates[1].Path.GetElem()[1].GetName())
				//assert.Equal(t, uint64(50), deviceGrouproupAction.Updates[3].Val.GetUintVal())
				assert.Equal(t, "site", deviceGrouproupAction.Updates[5].Path.GetElem()[0].Name)
				//assert.Equal(t, "store", deviceGrouproupAction.Updates[4].Path.GetElem()[0].Key["name"])

			} else if dgname == "starbucks-seattle-default" {
				assert.Len(t, deviceGrouproupAction.Deletes, 1)
				assert.Len(t, deviceGrouproupAction.Updates, 3)
				assert.Equal(t, "Seattle Default", deviceGrouproupAction.Updates[0].Val.GetStringVal())
				assert.Equal(t, "starbucks-seattle", deviceGrouproupAction.Updates[1].Val.GetStringVal())
				assert.Equal(t, "ip-domain", deviceGrouproupAction.Updates[2].Path.GetElem()[0].Name)

			} else if dgname == "starbucks-seattle-pos" {
				assert.Len(t, deviceGrouproupAction.Deletes, 1)
				assert.Len(t, deviceGrouproupAction.Updates, 7)
				assert.Equal(t, "Seattle POS", deviceGrouproupAction.Updates[0].Val.GetStringVal())
				assert.Equal(t, "imsi-range-to", deviceGrouproupAction.Updates[4].GetPath().GetElem()[1].Name)
				assert.Equal(t, "imsis", deviceGrouproupAction.Updates[2].Path.GetElem()[0].Name)
				assert.Equal(t, "starbucks-seattle", deviceGrouproupAction.Updates[6].Val.GetStringVal())

			}

		default:
			t.Errorf("unexpected device group: %s", dgname)
		}
	}

	// Expecting actions 34-36 to be Traffic-Class - order is changeable
	for idx := 14; idx <= 16; idx++ {
		trafficClassAction := actions[idx]
		assert.Empty(t, trafficClassAction.DeletePrefix)
		assert.Equal(t, "traffic-class", trafficClassAction.UpdatePrefix.GetElem()[0].GetName(),
			"unexpected type for %d", idx)
		assert.Len(t, trafficClassAction.Deletes, 1)
		assert.Len(t, trafficClassAction.Updates, 5)

		tcID, ok := trafficClassAction.UpdatePrefix.GetElem()[1].GetKey()["id"]
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
			assert.Equal(t, "cs4", trafficClassAction.Updates[2].GetPath().GetTarget())
			assert.Equal(t, uint64(833), trafficClassAction.Updates[4].Val.GetUintVal())
		default:
			t.Errorf("Unexpected Traffic Class Profile ID %s for %d", tcID, idx)
		}
	}

	//Expecting actions 17-19 to be Applications - order is changeable
	for idx := 17; idx <= 19; idx++ {
		applicationAction := actions[idx]
		assert.Empty(t, applicationAction.DeletePrefix)
		assert.Equal(t, "application", applicationAction.UpdatePrefix.GetElem()[0].GetName(),
			"unexpected type for %d", idx)
		assert.Len(t, applicationAction.Deletes, 1)
		assert.Len(t, applicationAction.Updates, 7)
		appID, ok := applicationAction.UpdatePrefix.GetElem()[1].GetKey()["id"]
		assert.True(t, ok)
		switch appID {
		case "acme-dataacquisition":
			assert.Equal(t, "Data Acquisition", applicationAction.Updates[0].Val.GetStringVal())
			assert.Equal(t, "acme", applicationAction.Updates[2].Val.GetStringVal())
			assert.Equal(t, "da", applicationAction.Updates[3].GetPath().GetElem()[0].Key["endpoint-id"])
		case "starbucks-fidelio":
			assert.Equal(t, "Fidelio", applicationAction.Updates[1].Val.GetStringVal())
			assert.Equal(t, "fidelio.starbucks.com", applicationAction.Updates[6].Val.GetStringVal())
			assert.Equal(t, "fidelio", applicationAction.Updates[3].GetPath().GetElem()[0].Key["endpoint-id"])
		case "starbucks-nvr":
			assert.Equal(t, "NVR", applicationAction.Updates[1].Val.GetStringVal())
			assert.Equal(t, "endpoint", applicationAction.Updates[4].GetPath().GetElem()[0].GetName())
			assert.Equal(t, "cs4", applicationAction.Updates[2].GetPath().Target)
		default:
			t.Errorf("Unexpected Application Profile ID %s for %d", appID, idx)

		}
	}

	// Expecting actions 20-23 to be IP Domains - order is changeable
	for idx := 20; idx <= 23; idx++ {
		ipDomainAction := actions[idx]
		assert.Empty(t, ipDomainAction.DeletePrefix)
		assert.Equal(t, "ip-domain", ipDomainAction.UpdatePrefix.GetElem()[0].GetName(), "unexpected type for %d", idx)
		assert.Len(t, ipDomainAction.Deletes, 1)
		assert.Len(t, ipDomainAction.Updates, 8)
		ipdID, ok := ipDomainAction.UpdatePrefix.GetElem()[1].GetKey()["id"]
		assert.True(t, ok)
		switch ipdID {
		case "acme-chicago":
			assert.Equal(t, "Chicago", ipDomainAction.Updates[1].Val.GetStringVal())
			assert.Equal(t, uint64(12690), ipDomainAction.Updates[4].Val.GetUintVal())
			assert.Equal(t, "admin-status", ipDomainAction.Updates[6].GetPath().GetElem()[0].GetName())
		case "defaultent-defaultip":
			assert.Equal(t, "Global Default IP Domain", ipDomainAction.Updates[0].Val.GetStringVal())
			assert.Equal(t, "8.8.8.2", ipDomainAction.Updates[3].Val.GetStringVal())
			assert.Equal(t, "dns-primary", ipDomainAction.Updates[2].GetPath().GetElem()[0].GetName())
		case "starbucks-newyork":
			assert.Equal(t, "New York", ipDomainAction.Updates[1].Val.GetStringVal())
			assert.Equal(t, "254.186.117.251/31", ipDomainAction.Updates[5].Val.GetStringVal())
			assert.Equal(t, "cs4", ipDomainAction.Updates[4].GetPath().GetTarget())
		case "starbucks-seattle":
			assert.Equal(t, "ENABLE", ipDomainAction.Updates[6].Val.GetStringVal())
		default:
			t.Errorf("Unexpected ip-domain Profile ID %s", ipdID)
		}
	}

	// Expecting actions 24-30 Site - order is changeable
	for idx := 24; idx <= 27; idx++ {
		siteAction := actions[idx]
		assert.Empty(t, siteAction.DeletePrefix)
		assert.Equal(t, "site", siteAction.UpdatePrefix.GetElem()[0].GetName(), "unexpected type for %d", idx)
		assert.Len(t, siteAction.Deletes, 1)
		siteID, ok := siteAction.UpdatePrefix.GetElem()[1].GetKey()["id"]
		assert.True(t, ok)
		switch siteID {
		case "acme-chicago":
			assert.Len(t, siteAction.Updates, 7)
			assert.Equal(t, "acme", siteAction.Updates[2].Val.GetStringVal())
			assert.Equal(t, "imsi-definition", siteAction.Updates[5].GetPath().GetElem()[0].GetName())
		case "defaultent-defaultsite":
			assert.Len(t, siteAction.Updates, 6)
			assert.Equal(t, "SSSSSSSSSSSSSSS", siteAction.Updates[5].Val.GetStringVal())
			assert.Equal(t, "Global Default Site", siteAction.Updates[0].Val.GetStringVal())
		case "starbucks-newyork":
			assert.Len(t, siteAction.Updates, 7)
			assert.Equal(t, uint64(2), siteAction.Updates[6].Val.GetUintVal())
			assert.Equal(t, "mcc", siteAction.Updates[3].GetPath().GetElem()[1].GetName())
		case "starbucks-seattle":
			assert.Len(t, siteAction.Updates, 7)
			assert.Equal(t, "mnc", siteAction.Updates[4].GetPath().GetElem()[1].GetName())
		default:
			t.Errorf("Unexpected Site Profile ID %s for %d", siteID, idx)
		}
	}

	// Expecting actions 24-30 Site Small Cell - order is changeable
	for idx := 28; idx <= 30; idx++ {
		siteAction := actions[idx]
		assert.Empty(t, siteAction.DeletePrefix)
		assert.Equal(t, "site", siteAction.UpdatePrefix.GetElem()[0].GetName(), "unexpected type for %d", idx)
		assert.Len(t, siteAction.Deletes, 1)
		siteID, ok := siteAction.UpdatePrefix.GetElem()[1].GetKey()["id"]
		assert.True(t, ok)
		switch siteID {
		case "acme-chicago":
			assert.Len(t, siteAction.Deletes, 1)
			assert.Len(t, siteAction.Updates, 3)
			assert.Equal(t, "ap2.chicago.acme.com", siteAction.Updates[0].Val.GetStringVal())
			assert.Equal(t, "enable", siteAction.Updates[2].GetPath().GetElem()[1].GetName())
		case "defaultent-defaultsite":
			assert.Len(t, siteAction.Deletes, 0)
			assert.Len(t, siteAction.Updates, 0)
		case "starbucks-newyork":
			assert.Len(t, siteAction.Deletes, 1)
			assert.Len(t, siteAction.Updates, 3)
			assert.Equal(t, "8002", siteAction.Updates[1].Val.GetStringVal())
		case "starbucks-seattle":
			assert.Len(t, siteAction.Deletes, 1)
			assert.Len(t, siteAction.Updates, 6)
			assert.Equal(t, "tac", siteAction.Updates[1].GetPath().GetElem()[1].GetName())
		default:
			t.Errorf("Unexpected Site Profile ID %s for %d", siteID, idx)
		}
	}

	// Expecting actions 31-31 Template - order is changeable
	for idx := 31; idx <= 32; idx++ {
		templateAction := actions[idx]
		assert.Empty(t, templateAction.DeletePrefix)
		assert.Equal(t, "template", templateAction.UpdatePrefix.GetElem()[0].GetName(), "unexpected type for %d", idx)
		assert.Len(t, templateAction.Deletes, 1)
		assert.Len(t, templateAction.Updates, 7)
		templateID, ok := templateAction.UpdatePrefix.GetElem()[1].GetKey()["id"]
		assert.True(t, ok)
		switch templateID {
		case "template-1":
			assert.Equal(t, "VCS Template 1", templateAction.Updates[0].Val.GetStringVal())
			assert.Equal(t, "default-behavior", templateAction.Updates[4].GetPath().GetElem()[0].GetName())
			assert.Equal(t, "mbr", templateAction.Updates[5].GetPath().GetElem()[1].GetName())
		case "template-2":
			assert.Equal(t, "slice", templateAction.Updates[6].GetPath().GetElem()[0].GetName())
			assert.Equal(t, "DENY-ALL", templateAction.Updates[4].Val.GetStringVal())
			assert.Equal(t, "sst", templateAction.Updates[2].GetPath().GetElem()[0].GetName())
		default:
			t.Errorf("Unexpected Site Profile ID %s for %d", templateID, idx)
		}
	}

	// Expecting actions 33-36 to be Upf - order is changeable
	for idx := 33; idx <= 36; idx++ {
		upfAction := actions[idx]
		assert.Empty(t, upfAction.DeletePrefix)
		assert.Equal(t, "upf", upfAction.UpdatePrefix.GetElem()[0].GetName(), "expected upf for %d", idx)
		upfID, ok := upfAction.UpdatePrefix.GetElem()[1].GetKey()["id"]
		assert.Len(t, upfAction.Deletes, 1)
		assert.Len(t, upfAction.Updates, 6)
		assert.True(t, ok)
		switch upfID {
		case "acme-chicago-robots":
			assert.Equal(t, "Chicago Robots UPF", upfAction.Updates[0].Val.GetStringVal())
			assert.Equal(t, "enterprise", upfAction.Updates[4].GetPath().GetElem()[0].GetName())
			assert.Equal(t, uint64(6161), upfAction.Updates[3].Val.GetUintVal())
		case "starbucks-newyork-cameras":
			assert.Equal(t, "newyork.cameras-upf.starbucks.com", upfAction.Updates[2].Val.GetStringVal())
			assert.Equal(t, "New York Cameras", upfAction.Updates[1].Val.GetStringVal())
			assert.Equal(t, "port", upfAction.Updates[3].GetPath().GetElem()[0].GetName())
		case "starbucks-newyork-pos":
			assert.Equal(t, "NewYork POS UPF", upfAction.Updates[0].Val.GetStringVal())
			assert.Equal(t, "address", upfAction.Updates[2].GetPath().GetElem()[0].GetName())
			assert.Equal(t, "starbucks", upfAction.Updates[4].Val.GetStringVal())
		case "starbucks-seattle-cameras":
			assert.Equal(t, "seattle.cameras-upf.starbucks.com", upfAction.Updates[2].Val.GetStringVal())
			assert.Equal(t, "cs4", upfAction.Updates[0].GetPath().GetTarget())
			assert.Equal(t, uint64(9229), upfAction.Updates[3].Val.GetUintVal())
		default:
			t.Errorf("Unexpected UPF Profile ID %s", upfID)
		}
	}

	// Expecting actions 37-39 to be VCS - order is changeable
	for idx := 37; idx <= 39; idx++ {
		vcsAction := actions[idx]
		assert.Empty(t, vcsAction.DeletePrefix)
		assert.Equal(t, "vcs", vcsAction.UpdatePrefix.GetElem()[0].GetName(), "expected upf for %d", idx)
		vcsID, ok := vcsAction.UpdatePrefix.GetElem()[1].GetKey()["id"]
		assert.Len(t, vcsAction.Deletes, 1)
		assert.Len(t, vcsAction.Updates, 14)
		assert.True(t, ok)
		switch vcsID {
		case "acme-chicago-robots":
			assert.Equal(t, "Chicago Robots VCS", vcsAction.Updates[1].Val.GetStringVal())
			assert.Equal(t, "upf", vcsAction.Updates[2].GetPath().GetElem()[0].GetName())
			assert.Equal(t, "mbr", vcsAction.Updates[12].GetPath().GetElem()[1].GetName())
		case "starbucks-newyork-cameras":
			assert.Equal(t, "starbucks-newyork-cameras", vcsAction.Updates[2].Val.GetStringVal())
			assert.Equal(t, "sst", vcsAction.Updates[4].GetPath().GetElem()[0].GetName())
			assert.Equal(t, "DENY-ALL", vcsAction.Updates[6].Val.GetStringVal())
		case "starbucks-seattle-cameras":
			assert.Equal(t, "starbucks", vcsAction.Updates[3].Val.GetStringVal())
			assert.Equal(t, "cs4", vcsAction.Updates[11].GetPath().GetTarget())
			assert.Equal(t, "device-group", vcsAction.Updates[8].GetPath().GetElem()[0].GetName())
		default:
			t.Errorf("Unexpected UPF Profile ID %s", vcsID)
		}
	}

	//Expecting actions 40-42 to be Device Group (device and traffic-class migration)- order is changeable
	for idx := 40; idx <= 42; idx++ {
		deviceGrouproupAction := actions[idx]
		assert.Empty(t, deviceGrouproupAction.DeletePrefix)
		assert.Equal(t, "device-group", deviceGrouproupAction.UpdatePrefix.GetElem()[0].GetName(),
			"unexpected type for %d", idx)
		dgname, ok := deviceGrouproupAction.UpdatePrefix.Elem[1].Key["id"]
		assert.True(t, ok)
		assert.Len(t, deviceGrouproupAction.Deletes, 1)
		assert.Len(t, deviceGrouproupAction.Updates, 3)
		switch dgname {
		case "acme-chicago-robots":
			assert.Equal(t, "device", deviceGrouproupAction.Updates[0].GetPath().GetElem()[0].GetName())
			assert.Equal(t, "class-2", deviceGrouproupAction.Updates[2].Val.GetStringVal())
		case "starbucks-newyork-cameras":
			assert.Equal(t, "mbr", deviceGrouproupAction.Updates[1].GetPath().GetElem()[1].GetName())
			assert.Equal(t, "class-1", deviceGrouproupAction.Updates[2].Val.GetStringVal())
		case "starbucks-seattle-cameras":
			assert.Equal(t, "cs4", deviceGrouproupAction.Updates[1].GetPath().GetTarget())
			assert.Equal(t, "device", deviceGrouproupAction.Updates[2].GetPath().GetElem()[0].GetName())
		default:
			t.Errorf("unexpected device group: %s", dgname)
		}
	}

	//Expecting actions 43-45 to be Applications (traffic-class migration)- order is changeable
	for idx := 43; idx <= 45; idx++ {
		applicationAction := actions[idx]
		assert.Empty(t, applicationAction.DeletePrefix)
		assert.Len(t, applicationAction.Deletes, 1)
		assert.Len(t, applicationAction.Updates, 0)
	}

	// Expecting actions 46-48 to be Upf (site)  - order is changeable
	for idx := 46; idx <= 48; idx++ {
		upfAction := actions[idx]
		assert.Empty(t, upfAction.DeletePrefix)
		upfID, ok := upfAction.UpdatePrefix.GetElem()[1].GetKey()["id"]
		assert.True(t, ok)
		assert.Len(t, upfAction.Deletes, 1)
		assert.Len(t, upfAction.Updates, 1)
		assert.Equal(t, "site", upfAction.Updates[0].GetPath().GetElem()[0].GetName())
		switch upfID {
		case "acme-chicago-robots":
			assert.Equal(t, "acme-chicago", upfAction.Updates[0].Val.GetStringVal())
		case "starbucks-newyork-cameras":
			assert.Equal(t, "starbucks-newyork", upfAction.Updates[0].Val.GetStringVal())
		case "starbucks-seattle-cameras":
			assert.Equal(t, "starbucks-seattle", upfAction.Updates[0].Val.GetStringVal())
		default:
			t.Errorf("Unexpected UPF Profile ID %s", upfID)
		}
	}
}
