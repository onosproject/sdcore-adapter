// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package steps

import (
	"fmt"
	modelsv21 "github.com/onosproject/config-models/modelplugin/aether-2.1.0/aether_2_1_0"
	modelpluginv21 "github.com/onosproject/config-models/modelplugin/aether-2.1.0/modelplugin"
	modelsv3 "github.com/onosproject/config-models/modelplugin/aether-3.0.0/aether_3_0_0"
	modelpluginv3 "github.com/onosproject/config-models/modelplugin/aether-3.0.0/modelplugin"
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	"github.com/onosproject/sdcore-adapter/pkg/migration"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

func Test_MigrateV21V3(t *testing.T) {
	srcJSON, err := ioutil.ReadFile("./testdata/mega_patch_210_src.json")
	assert.NoError(t, err)
	srcValJSON := &gpb.TypedValue{
		Value: &gpb.TypedValue_JsonVal{
			JsonVal: srcJSON,
		},
	}
	destJSON, err := ioutil.ReadFile("./testdata/mega_patch_300_dest.json")
	assert.NoError(t, err)
	dstValJSON := &gpb.TypedValue{
		Value: &gpb.TypedValue_JsonVal{
			JsonVal: destJSON,
		},
	}

	v21Models := gnmi.NewModel(modelpluginv21.ModelData,
		reflect.TypeOf((*modelsv21.Device)(nil)),
		modelsv21.SchemaTree["Device"],
		modelsv21.Unmarshal,
		//models.ΛEnum  // NOTE: There is no Enum in the aether models? So use a blank map.
		map[string]map[int64]ygot.EnumDefinition{},
	)

	v3Models := gnmi.NewModel(modelpluginv3.ModelData,
		reflect.TypeOf((*modelsv3.Device)(nil)),
		modelsv3.SchemaTree["Device"],
		modelsv3.Unmarshal,
		//models.ΛEnum  // NOTE: There is no Enum in the aether models? So use a blank map.
		map[string]map[int64]ygot.EnumDefinition{},
	)

	migrateV21V3Step := &migration.MigrationStep{
		FromVersion:   "2.1.0",
		FromModels:    v21Models,
		ToVersion:     "3.0.0",
		ToModels:      v3Models,
		MigrationFunc: MigrateV21V3,
		Migrator:      nil,
	}

	actions, err := MigrateV21V3(migrateV21V3Step, "cs2", "cs3", srcValJSON, dstValJSON)
	assert.NoError(t, err)
	assert.Len(t, actions, 60, "unexpected: actions is %d items", len(actions))

	csAction := actions[0]
	assert.Empty(t, csAction.DeletePrefix)
	assert.Len(t, csAction.Deletes, 1)

	assert.Equal(t, `elem:{name:"connectivity-service"} elem:{name:"connectivity-service" key:{key:"id" value:"cs-4g"}} target:"cs3"`,
		strings.ReplaceAll(csAction.UpdatePrefix.String(), "  ", " "))
	assert.Len(t, csAction.Updates, 2)
	assert.Equal(t, `path:{elem:{name:"description"} target:"cs3"} val:{string_val:"Connectivity service endpoints"}`,
		strings.ReplaceAll(csAction.Updates[0].String(), "  ", " "))
	assert.Equal(t, `path:{elem:{name:"display-name"} target:"cs3"} val:{string_val:"4G Connectivity Service"}`,
		strings.ReplaceAll(csAction.Updates[1].String(), "  ", " "))

	// Expecting action 1 to be enterprise defaultent
	defaultEnterpriseAction := actions[1]
	assert.Empty(t, defaultEnterpriseAction.DeletePrefix)
	assert.Equal(t, "enterprise", defaultEnterpriseAction.UpdatePrefix.GetElem()[0].GetName(),
		"unexpected type for %d", 1)
	defaultEntName, ok := defaultEnterpriseAction.UpdatePrefix.Elem[1].Key["id"]
	assert.True(t, ok)
	assert.Equal(t, "defaultent", defaultEntName)

	// Expecting action 2 to be ip-domain defaultent-defaultip
	defaultIPDomainAction := actions[2]
	assert.Empty(t, defaultIPDomainAction.DeletePrefix)
	assert.Equal(t, "ip-domain", defaultIPDomainAction.UpdatePrefix.GetElem()[0].GetName(),
		"unexpected type for %d", 1)
	defaultIPDomainName, ok := defaultIPDomainAction.UpdatePrefix.Elem[1].Key["id"]
	assert.True(t, ok)
	assert.Equal(t, "defaultent-defaultip", defaultIPDomainName)

	// Expecting action 3 to be ip-domain defaultent-defaultsite
	defaultSiteAction := actions[3]
	assert.Empty(t, defaultSiteAction.DeletePrefix)
	assert.Equal(t, "site", defaultSiteAction.UpdatePrefix.GetElem()[0].GetName(),
		"unexpected type for %d", 3)
	defaultSiteName, ok := defaultSiteAction.UpdatePrefix.Elem[1].Key["id"]
	assert.True(t, ok)
	assert.Equal(t, "defaultent-defaultsite", defaultSiteName)

	// Expecting action 4 to be ip-domain defaultent-defaultent-defaultdg
	defaultDeviceGroupAction := actions[4]
	assert.Empty(t, defaultDeviceGroupAction.DeletePrefix)
	assert.Equal(t, "device-group", defaultDeviceGroupAction.UpdatePrefix.GetElem()[0].GetName(),
		"unexpected type for %d", 4)
	defaultDeviceGroupName, ok := defaultDeviceGroupAction.UpdatePrefix.Elem[1].Key["id"]
	assert.True(t, ok)
	assert.Equal(t, "defaultent-defaultdg", defaultDeviceGroupName)

	// Expecting actions 5-13 to be enterprises - order is changeable
	for idx := 5; idx < 13; idx++ {
		enterpriseAction := actions[idx]
		assert.Empty(t, enterpriseAction.DeletePrefix)
		assert.Equal(t, "enterprise", enterpriseAction.UpdatePrefix.GetElem()[0].GetName(),
			"unexpected type for %d", idx)

		entname, ok := enterpriseAction.UpdatePrefix.Elem[1].Key["id"]
		assert.True(t, ok)
		switch entname {
		case "aether-ciena", "aether-ntt", "aether-intel", "aether-onf", "aether-tef",
			"pronto-cornell", "pronto-princeton", "pronto-stanford":
			assert.Len(t, enterpriseAction.Deletes, 1)
			assert.Len(t, enterpriseAction.Updates, 3)
			assert.Equal(t, "description", enterpriseAction.Updates[0].Path.GetElem()[0].Name)
			assert.Equal(t, fmt.Sprintf("%s-description", entname), enterpriseAction.Updates[0].Val.GetStringVal())
			assert.Equal(t, "display-name", enterpriseAction.Updates[1].Path.GetElem()[0].Name)
			assert.Equal(t, fmt.Sprintf("%s-display-name", entname), enterpriseAction.Updates[1].Val.GetStringVal())
			assert.Equal(t, "connectivity-service", enterpriseAction.Updates[2].Path.GetElem()[0].Name)
			cs, ok := enterpriseAction.Updates[2].Path.GetElem()[0].Key["connectivity-service"]
			assert.True(t, ok)
			assert.Equal(t, "cs-4g", cs)
			assert.Equal(t, len(entname)%2 == 0, enterpriseAction.Updates[2].Val.GetBoolVal(),
				"expected connectivity-service for %s to be true", entname)
		default:
			t.Errorf("unexpected enterprise: %s", entname)
		}
	}

	// Expecting actions 10-19 to be Upf (from UP Profile) - order is changeable
	for idx := 14; idx < 23; idx++ {
		action := actions[idx]
		assert.Empty(t, action.DeletePrefix)
		assert.Equal(t, "upf", action.UpdatePrefix.GetElem()[0].GetName(), "expected upf for %d", idx)
		upID, ok := action.UpdatePrefix.GetElem()[1].GetKey()["id"]
		assert.Len(t, action.Deletes, 1, "Unexpected # deletes %d for upf %s", len(action.Deletes), upID)
		assert.Len(t, action.Updates, 5, "Unexpected # updates %d for upf %s", len(action.Updates), upID)
		assert.True(t, ok)
		switch upID {
		case "ciena":
			assert.Equal(t, "description", action.Updates[0].GetPath().GetElem()[0].GetName())
			assert.Equal(t, "User plane profile for Ciena", action.Updates[0].Val.GetStringVal())

			assert.Equal(t, "display-name", action.Updates[1].GetPath().GetElem()[0].GetName())
			assert.Equal(t, "Ciena", action.Updates[1].Val.GetStringVal())

			assert.Equal(t, "address", action.Updates[2].GetPath().GetElem()[0].GetName())
			assert.Equal(t, "pfcp-agent.omec.svc.prd.ciena.aetherproject.net", action.Updates[2].Val.GetStringVal())

			assert.Equal(t, "port", action.Updates[3].GetPath().GetElem()[0].GetName())
			assert.Equal(t, uint64(8080), action.Updates[3].Val.GetUintVal())

			assert.Equal(t, "enterprise", action.Updates[4].GetPath().GetElem()[0].GetName())
			assert.Equal(t, "aether-ciena", action.Updates[4].Val.GetStringVal())

		case "cornell1", "intel", "ntt", "menlo", "tucson", "princeton1", "stanford1", "stanford2", "tef":
		default:
			t.Errorf("Unexpected UP Profile ID %s", upID)
		}
	}

	// Expecting actions 21-30 to be IP Domains (from APN Profile) - order is changeable
	for idx := 24; idx < 34; idx++ {
		action := actions[idx]
		assert.Empty(t, action.DeletePrefix)
		assert.Equal(t, "ip-domain", action.UpdatePrefix.GetElem()[0].GetName(), "unexpected type for %d", idx)
		apnID, ok := action.UpdatePrefix.GetElem()[1].GetKey()["id"]
		assert.Len(t, action.Deletes, 1, "Unexpected # deletes %d for ip-domain %s", len(action.Deletes), apnID)
		assert.Len(t, action.Updates, 9, "Unexpected # updates %d for ip-domain %s", len(action.Updates), apnID)
		assert.True(t, ok)
		switch apnID {
		case "apn-internet-ciena":
			assert.Equal(t, "description", action.Updates[0].GetPath().GetElem()[0].GetName())
			assert.Equal(t, "Ciena Internet APN config", action.Updates[0].Val.GetStringVal())

			assert.Equal(t, "display-name", action.Updates[1].GetPath().GetElem()[0].GetName())
			assert.Equal(t, "Ciena Internet", action.Updates[1].Val.GetStringVal())

			assert.Equal(t, "dns-primary", action.Updates[2].GetPath().GetElem()[0].GetName())
			assert.Equal(t, "10.24.7.11", action.Updates[2].Val.GetStringVal())

			assert.Equal(t, "dns-secondary", action.Updates[3].GetPath().GetElem()[0].GetName())
			assert.Equal(t, "1.1.1.1", action.Updates[3].Val.GetStringVal())

			assert.Equal(t, "dnn", action.Updates[4].GetPath().GetElem()[0].GetName())
			assert.Equal(t, "internet", action.Updates[4].Val.GetStringVal())

			assert.Equal(t, "mtu", action.Updates[5].GetPath().GetElem()[0].GetName())
			assert.Equal(t, uint64(1460), action.Updates[5].Val.GetUintVal())

			assert.Equal(t, "admin-status", action.Updates[6].GetPath().GetElem()[0].GetName())
			assert.Equal(t, "ENABLE", action.Updates[6].Val.GetStringVal())

			assert.Equal(t, "subnet", action.Updates[7].GetPath().GetElem()[0].GetName())
			assert.Equal(t, "255.255.255.255/32", action.Updates[7].Val.GetStringVal())

			assert.Equal(t, "enterprise", action.Updates[8].GetPath().GetElem()[0].GetName())
			assert.Equal(t, "aether-ciena", action.Updates[8].Val.GetStringVal())

		case "apn-internet-cornell1", "apn-internet-default", "apn-internet-intel", "apn-internet-menlo",
			"apn-internet-tucson", "apn-internet-princeton1", "apn-internet-stanford1", "apn-internet-stanford2",
			"apn-internet-tef", "apn-profile1":
		default:
			t.Errorf("Unexpected APN Profile ID %s", apnID)
		}
	}

	// Expecting actions 34-36 to be Templates and Traffic-Class (from QOS Profile) - order is changeable
	for idx := 34; idx < 39; idx++ {
		action := actions[idx]
		assert.Empty(t, action.DeletePrefix)
		apnID, ok := action.UpdatePrefix.GetElem()[1].GetKey()["id"]
		assert.True(t, ok)
		switch apnID {
		case "qos-profile1":
			assert.Equal(t, "description", action.Updates[0].GetPath().GetElem()[0].GetName())
			assert.Equal(t, "low bitrate internet service", action.Updates[0].Val.GetStringVal())

			assert.Equal(t, "display-name", action.Updates[1].GetPath().GetElem()[0].GetName())
			assert.Equal(t, "QOS Profile 1", action.Updates[1].Val.GetStringVal())

			switch objType := action.UpdatePrefix.GetElem()[0].GetName(); objType {
			case "template":
				assert.Len(t, action.Deletes, 1, "Unexpected # deletes %d for template %s", len(action.Deletes), apnID)
				assert.Len(t, action.Updates, 5, "Unexpected # updates %d for template %s", len(action.Updates), apnID)

				assert.Equal(t, "uplink", action.Updates[2].GetPath().GetElem()[0].GetName())
				assert.Equal(t, uint64(12), action.Updates[2].Val.GetUintVal())

				assert.Equal(t, "downlink", action.Updates[3].GetPath().GetElem()[0].GetName())
				assert.Equal(t, uint64(12), action.Updates[3].Val.GetUintVal())

				assert.Equal(t, "traffic-class", action.Updates[4].GetPath().GetElem()[0].GetName())
				assert.Equal(t, "qos-profile1", action.Updates[4].Val.GetStringVal())

			case "traffic-class":
				assert.Len(t, action.Deletes, 0, "Unexpected # deletes %d for traffic-class %s", len(action.Deletes), apnID)
				assert.Len(t, action.Updates, 3, "Unexpected # updates %d for traffic-class %s", len(action.Updates), apnID)

				assert.Equal(t, "qci", action.Updates[2].GetPath().GetElem()[0].GetName())
				assert.Equal(t, uint64(23), action.Updates[2].Val.GetUintVal())

			default:
				t.Errorf("unexpected type %s when processing %s for %d", objType, apnID, idx)
			}

		case "sed", "culpa":
			switch objType := action.UpdatePrefix.GetElem()[0].GetName(); objType {
			case "template", "traffic-class":
			default:
				t.Errorf("unexpected type %s when processing %s for %d", objType, apnID, idx)
			}
		default:
			t.Errorf("Unexpected APN Profile ID %s for %d", apnID, idx)
		}
	}

	// Expecting actions 36-50 to be Sites and Device Groups from UE- order is changeable
	// Telefonica has no imsi-range, and so should not create a device group
	for idx := 39; idx < 59; idx++ {
		action := actions[idx]
		assert.Empty(t, action.DeletePrefix)
		uuid, ok := action.UpdatePrefix.GetElem()[1].GetKey()["id"]
		assert.True(t, ok)
		switch uuid {
		case "a4c814a64-c592-468e-9435-b60f225": // Take 1 as an example
			switch objType := action.UpdatePrefix.GetElem()[0].GetName(); objType {
			case "site":
				assert.Len(t, action.Deletes, 1)
				assert.Len(t, action.Updates, 5)
				assert.Equal(t, "Ciena subscriber match rule", action.Updates[0].Val.GetStringVal())
				assert.Equal(t, "enterprise", action.Updates[1].GetPath().GetElem()[0].GetName())
				assert.Equal(t, "aether-ciena", action.Updates[1].Val.GetStringVal())
				assert.Equal(t, "031", action.Updates[2].Val.GetStringVal())
				assert.Equal(t, "01", action.Updates[3].Val.GetStringVal())
				assert.Equal(t, "SSSSSSSSSSSSSSS", action.Updates[4].Val.GetStringVal())
			case "device-group":
				assert.Len(t, action.Deletes, 0)
				assert.Len(t, action.Updates, 5)
				assert.Equal(t, "Ciena subscriber match rule", action.Updates[0].Val.GetStringVal())
				assert.Equal(t, "site", action.Updates[1].GetPath().GetElem()[0].GetName())
				assert.Equal(t, "a4c814a64-c592-468e-9435-b60f225", action.Updates[1].Val.GetStringVal())
				assert.Equal(t, "ip-domain", action.Updates[2].GetPath().GetElem()[0].GetName())
				assert.Equal(t, "apn-internet-ciena", action.Updates[2].Val.GetStringVal())
				assert.Len(t, action.Updates[3].GetPath().GetElem(), 2)
				imsiRangeNameStart, imsiRangeStartKeyOk := action.Updates[3].GetPath().GetElem()[0].GetKey()["name"]
				assert.True(t, imsiRangeStartKeyOk)
				assert.Equal(t, "range-1", imsiRangeNameStart)
				assert.Equal(t, uint64(315010101000001), action.Updates[3].GetVal().GetUintVal())
				imsiRangeNameEnd, imsiRangeEndKeyOk := action.Updates[3].GetPath().GetElem()[0].GetKey()["name"]
				assert.True(t, imsiRangeEndKeyOk)
				assert.Equal(t, "range-1", imsiRangeNameEnd)
				assert.Equal(t, uint64(315010101000010), action.Updates[4].GetVal().GetUintVal())
			default:
				t.Errorf("unexpected object type %s for %d", objType, idx)
			}

		case "a1c6852e6-5b12-413a-9fa5-c631c64",
			"a0debf047-8416-4539-9abf-02a0d7e", "a554b4c5b-de49-4868-ba7e-f428aef",
			"f2ba8cc0-e593-403b-a130-f18a9901", "f5a0929f-b4a4-4f34-8bd5-52c57eeb",
			"e8b4f8ea-cd9c-4ae7-a1df-15ee82cc", "cbdb20c1-c3d7-47e3-a1a1-7465c8ad",
			"a30f77900-18b1-480c-a419-031956d", "a415d0496-6926-4a49-b0f1-69ef174",
			"c6711eb4-5210-4d94-b83c-0f890dc2":
			switch objType := action.UpdatePrefix.GetElem()[0].GetName(); objType {
			case "site", "device-group":
			default:
				t.Errorf("unexpected object type %s for %d", objType, idx)
			}
		default:
			t.Errorf("Unexpected UE ID %s for %d", uuid, idx)
		}
	}

}

func Test_truncateUuid(t *testing.T) {
	u1 := "4c814a64-c592-468e-9435-b60f225f97ff"
	assert.Equal(t, "a4c814a64-c592-468e-9435-b60f225", truncateUUID(&u1))
	u2 := "c6711eb4-5210-4d94-b83c-0f890dc21c31"
	assert.Equal(t, "c6711eb4-5210-4d94-b83c-0f890dc2", truncateUUID(&u2))
}
