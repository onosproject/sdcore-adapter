// SPDX-FileCopyrightText: 2022-present Intel Corporation
// SPDX-FileCopyrightText: 2022-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package steps

import (
	modelsv2 "github.com/onosproject/aether-models/models/aether-2.0.x/api"
	modelsv21 "github.com/onosproject/aether-models/models/aether-2.1.x/api"
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

func Test_MigrateV2V21(t *testing.T) {
	srcJSON, err := ioutil.ReadFile("./testdata/mega_patch_200_src.json")
	assert.NoError(t, err)
	srcValJSON := &gpb.TypedValue{
		Value: &gpb.TypedValue_JsonVal{
			JsonVal: srcJSON,
		},
	}

	dstValJSON := &gpb.TypedValue{
		Value: &gpb.TypedValue_JsonVal{},
	}

	v2Models := gnmi.NewModel(modelsv2.ModelData(),
		reflect.TypeOf((*modelsv2.Device)(nil)),
		modelsv2.SchemaTree["Device"],
		modelsv2.Unmarshal,
		//models.ΛEnum  // NOTE: There is no Enum in the aether models? So use a blank map.
		map[string]map[int64]ygot.EnumDefinition{},
	)

	v21Models := gnmi.NewModel(modelsv21.ModelData(),
		reflect.TypeOf((*modelsv21.Device)(nil)),
		modelsv21.SchemaTree["Device"],
		modelsv21.Unmarshal,
		//models.ΛEnum  // NOTE: There is no Enum in the aether models? So use a blank map.
		map[string]map[int64]ygot.EnumDefinition{},
	)

	migrateV2V21Step := &migration.MigrationStep{
		FromVersion:   "2.0.0",
		FromModels:    v2Models,
		ToVersion:     "2.1.0",
		ToModels:      v21Models,
		MigrationFunc: MigrateV2V21,
		Migrator:      nil,
	}

	actions, err := MigrateV2V21(migrateV2V21Step, "cs2", "notused", srcValJSON, dstValJSON)
	assert.NoError(t, err)
	assert.Len(t, actions, 55, "unexpected: actions is %d items", len(actions))

	for _, a := range actions {
		switch updPfx := strings.ReplaceAll(a.UpdatePrefix.String(), "  ", " "); updPfx {
		case `elem:{name:"traffic-class" key:{key:"traffic-class-id" value:"class-1"}} target:"defaultent"`:
			for _, upd := range a.Updates {
				switch updStr := strings.ReplaceAll(upd.Path.String(), "  ", " "); updStr {
				case `elem:{name:"description"} target:"defaultent"`:
					assert.Equal(t, `High Priority TC`, upd.GetVal().GetStringVal())
				case `elem:{name:"display-name"} target:"defaultent"`:
					assert.Equal(t, `Class 1`, upd.GetVal().GetStringVal())
				case `elem:{name:"qci"} target:"defaultent"`:
					assert.Equal(t, uint64(10), upd.GetVal().GetUintVal())
				case `elem:{name:"pelr"} target:"defaultent"`:
					assert.Equal(t, uint64(10), upd.GetVal().GetUintVal())
				case `elem:{name:"arp"} target:"defaultent"`:
					assert.Equal(t, uint64(1), upd.GetVal().GetUintVal())
				case `elem:{name:"pdb"} target:"defaultent"`:
					assert.Equal(t, uint64(100), upd.GetVal().GetUintVal())
				default:
					t.Errorf("unexpected update %s", updStr)
					t.FailNow()
				}
			}
		case `elem:{name:"traffic-class" key:{key:"traffic-class-id" value:"class-1"}} target:"acme"`:
			for _, upd := range a.Updates {
				switch updStr := strings.ReplaceAll(upd.Path.String(), "  ", " "); updStr {
				case `elem:{name:"description"} target:"acme"`:
					assert.Equal(t, `High Priority TC`, upd.GetVal().GetStringVal())
				case `elem:{name:"display-name"} target:"acme"`:
					assert.Equal(t, `Class 1`, upd.GetVal().GetStringVal())
				case `elem:{name:"qci"} target:"acme"`:
					assert.Equal(t, uint64(10), upd.GetVal().GetUintVal())
				case `elem:{name:"pelr"} target:"acme"`:
					assert.Equal(t, uint64(10), upd.GetVal().GetUintVal())
				case `elem:{name:"arp"} target:"acme"`:
					assert.Equal(t, uint64(1), upd.GetVal().GetUintVal())
				case `elem:{name:"pdb"} target:"acme"`:
					assert.Equal(t, uint64(100), upd.GetVal().GetUintVal())
				default:
					t.Errorf("unexpected update %s", updStr)
					t.FailNow()
				}
			}
		case `elem:{name:"traffic-class" key:{key:"traffic-class-id" value:"class-2"}} target:"acme"`:
			for _, upd := range a.Updates {
				switch updStr := strings.ReplaceAll(upd.Path.String(), "  ", " "); updStr {
				case `elem:{name:"description"} target:"acme"`:
					assert.Equal(t, `Medium Priority TC`, upd.GetVal().GetStringVal())
				case `elem:{name:"display-name"} target:"acme"`:
					assert.Equal(t, `Class 2`, upd.GetVal().GetStringVal())
				case `elem:{name:"qci"} target:"acme"`:
					assert.Equal(t, uint64(20), upd.GetVal().GetUintVal())
				case `elem:{name:"pelr"} target:"acme"`:
					assert.Equal(t, uint64(10), upd.GetVal().GetUintVal())
				case `elem:{name:"arp"} target:"acme"`:
					assert.Equal(t, uint64(1), upd.GetVal().GetUintVal())
				case `elem:{name:"pdb"} target:"acme"`:
					assert.Equal(t, uint64(100), upd.GetVal().GetUintVal())
				default:
					t.Errorf("unexpected update %s", updStr)
					t.FailNow()
				}
			}
		case `elem:{name:"traffic-class" key:{key:"traffic-class-id" value:"class-3"}} target:"acme"`:
			for _, upd := range a.Updates {
				switch updStr := strings.ReplaceAll(upd.Path.String(), "  ", " "); updStr {
				case `elem:{name:"description"} target:"acme"`:
					assert.Equal(t, `Low Priority TC`, upd.GetVal().GetStringVal())
				case `elem:{name:"display-name"} target:"acme"`:
					assert.Equal(t, `Class 3`, upd.GetVal().GetStringVal())
				case `elem:{name:"qci"} target:"acme"`:
					assert.Equal(t, uint64(30), upd.GetVal().GetUintVal())
				case `elem:{name:"pelr"} target:"acme"`:
					assert.Equal(t, uint64(10), upd.GetVal().GetUintVal())
				case `elem:{name:"arp"} target:"acme"`:
					assert.Equal(t, uint64(1), upd.GetVal().GetUintVal())
				case `elem:{name:"pdb"} target:"acme"`:
					assert.Equal(t, uint64(100), upd.GetVal().GetUintVal())
				default:
					t.Errorf("unexpected update %s", updStr)
					t.FailNow()
				}
			}
		case `elem:{name:"template" key:{key:"template-id" value:"template-1"}} target:"acme"`:
			for _, upd := range a.Updates {
				switch updStr := strings.ReplaceAll(upd.Path.String(), "  ", " "); updStr {
				case `elem:{name:"description"} target:"acme"`:
					assert.Equal(t, `Slice Template 1`, upd.GetVal().GetStringVal())
				case `elem:{name:"display-name"} target:"acme"`:
					assert.Equal(t, `Template 1`, upd.GetVal().GetStringVal())
				case `elem:{name:"sd"} target:"acme"`:
					assert.Equal(t, uint64(10886763), upd.GetVal().GetUintVal())
				case `elem:{name:"sst"} target:"acme"`:
					assert.Equal(t, uint64(158), upd.GetVal().GetUintVal())
				case `elem:{name:"default-behavior"} target:"acme"`:
					assert.Equal(t, `DENY-ALL`, upd.GetVal().GetStringVal())
				case `elem:{name:"mbr"} elem:{name:"uplink"} target:"acme"`:
					assert.Equal(t, uint64(10000000), upd.GetVal().GetUintVal())
				case `elem:{name:"mbr"} elem:{name:"downlink"} target:"acme"`:
					assert.Equal(t, uint64(5000000), upd.GetVal().GetUintVal())
				case `elem:{name:"mbr"} elem:{name:"uplink-burst-size"} target:"acme"`:
					assert.Equal(t, uint64(600000), upd.GetVal().GetUintVal())
				case `elem:{name:"mbr"} elem:{name:"downlink-burst-size"} target:"acme"`:
					assert.Equal(t, uint64(600000), upd.GetVal().GetUintVal())
				default:
					t.Errorf("unexpected update %s", updStr)
					t.FailNow()
				}
			}
		case `elem:{name:"application" key:{key:"application-id" value:"acme-dataacquisition"}} target:"acme"`:
			for _, upd := range a.Updates {
				switch updStr := strings.ReplaceAll(upd.Path.String(), "  ", " "); updStr {
				case `elem:{name:"description"} target:"acme"`:
					assert.Equal(t, `Data Acquisition`, upd.GetVal().GetStringVal())
				case `elem:{name:"display-name"} target:"acme"`:
					assert.Equal(t, `DA`, upd.GetVal().GetStringVal())
				case `elem:{name:"address"} target:"acme"`:
					assert.Equal(t, `da.acme.com`, upd.GetVal().GetStringVal())
				case `elem:{name:"endpoint" key:{key:"endpoint-id" value:"da"}} elem:{name:"display-name"} target:"acme"`:
					assert.Equal(t, `data acquisition endpoint`, upd.GetVal().GetStringVal())
				case `elem:{name:"endpoint" key:{key:"endpoint-id" value:"da"}} elem:{name:"protocol"} target:"acme"`:
					assert.Equal(t, `TCP`, upd.GetVal().GetStringVal())
				case `elem:{name:"endpoint" key:{key:"endpoint-id" value:"da"}} elem:{name:"traffic-class"} target:"acme"`:
					assert.Equal(t, `class-2`, upd.GetVal().GetStringVal())
				case `elem:{name:"endpoint" key:{key:"endpoint-id" value:"da"}} elem:{name:"port-start"} target:"acme"`:
					assert.Equal(t, uint64(7585), upd.GetVal().GetUintVal())
				case `elem:{name:"endpoint" key:{key:"endpoint-id" value:"da"}} elem:{name:"port-end"} target:"acme"`:
					assert.Equal(t, uint64(7588), upd.GetVal().GetUintVal())
				case `elem:{name:"endpoint" key:{key:"endpoint-id" value:"da"}} elem:{name:"mbr"} elem:{name:"uplink"} target:"acme"`:
					assert.Equal(t, uint64(2000000), upd.GetVal().GetUintVal())
				case `elem:{name:"endpoint" key:{key:"endpoint-id" value:"da"}} elem:{name:"mbr"} elem:{name:"downlink"} target:"acme"`:
					assert.Equal(t, uint64(1000000), upd.GetVal().GetUintVal())
				default:
					t.Errorf("unexpected update %s", updStr)
					t.FailNow()
				}
			}
		case `elem:{name:"application" key:{key:"application-id" value:"starbucks-nvr"}} target:"starbucks"`:
			for _, upd := range a.Updates {
				switch updStr := strings.ReplaceAll(upd.Path.String(), "  ", " "); updStr {
				case `elem:{name:"description"} target:"starbucks"`:
					assert.Equal(t, `Network Video Recorder`, upd.GetVal().GetStringVal())
				case `elem:{name:"display-name"} target:"starbucks"`:
					assert.Equal(t, `NVR`, upd.GetVal().GetStringVal())
				case `elem:{name:"address"} target:"starbucks"`:
					assert.Equal(t, `nvr.starbucks.com`, upd.GetVal().GetStringVal())
				case `elem:{name:"endpoint" key:{key:"endpoint-id" value:"rtsp"}} elem:{name:"display-name"} target:"starbucks"`:
					assert.Equal(t, `rtsp port`, upd.GetVal().GetStringVal())
				case `elem:{name:"endpoint" key:{key:"endpoint-id" value:"rtsp"}} elem:{name:"protocol"} target:"starbucks"`:
					assert.Equal(t, `UDP`, upd.GetVal().GetStringVal())
				case `elem:{name:"endpoint" key:{key:"endpoint-id" value:"rtsp"}} elem:{name:"traffic-class"} target:"starbucks"`:
					assert.Equal(t, `class-1`, upd.GetVal().GetStringVal())
				case `elem:{name:"endpoint" key:{key:"endpoint-id" value:"rtsp"}} elem:{name:"port-start"} target:"starbucks"`:
					assert.Equal(t, uint64(3316), upd.GetVal().GetUintVal())
				case `elem:{name:"endpoint" key:{key:"endpoint-id" value:"rtsp"}} elem:{name:"port-end"} target:"starbucks"`:
					assert.Equal(t, uint64(3330), upd.GetVal().GetUintVal())
				case `elem:{name:"endpoint" key:{key:"endpoint-id" value:"rtsp"}} elem:{name:"mbr"} elem:{name:"uplink"} target:"starbucks"`:
					assert.Equal(t, uint64(1000000), upd.GetVal().GetUintVal())
				case `elem:{name:"endpoint" key:{key:"endpoint-id" value:"rtsp"}} elem:{name:"mbr"} elem:{name:"downlink"} target:"starbucks"`:
					assert.Equal(t, uint64(1000000), upd.GetVal().GetUintVal())
				default:
					t.Errorf("unexpected update %s", updStr)
					t.FailNow()
				}
			}
		case `elem:{name:"site" key:{key:"site-id" value:"defaultent-defaultsite"}} target:"defaultent"`:
			for _, upd := range a.Updates {
				switch updStr := strings.ReplaceAll(upd.Path.String(), "  ", " "); updStr {
				case `elem:{name:"description"} target:"defaultent"`:
					assert.Equal(t, `Global Default Site`, upd.GetVal().GetStringVal())
				case `elem:{name:"display-name"} target:"defaultent"`:
					assert.Equal(t, `Global Default Site`, upd.GetVal().GetStringVal())
				case `elem:{name:"imsi-definition"} elem:{name:"mcc"} target:"defaultent"`:
					assert.Equal(t, `000`, upd.GetVal().GetStringVal())
				case `elem:{name:"imsi-definition"} elem:{name:"mnc"} target:"defaultent"`:
					assert.Equal(t, `00`, upd.GetVal().GetStringVal())
				case `elem:{name:"imsi-definition"} elem:{name:"format"} target:"defaultent"`:
					assert.Equal(t, `SSSSSSSSSSSSSSS`, upd.GetVal().GetStringVal())
				case `elem:{name:"imsi-definition"} elem:{name:"enterprise"} target:"defaultent"`:
					assert.Equal(t, uint64(0), upd.GetVal().GetUintVal())
				default:
					t.Errorf("unexpected update %s", updStr)
					t.FailNow()
				}
			}

		case `elem:{name:"site" key:{key:"site-id" value:"defaultent-defaultsite"}} elem:{name:"device-group" key:{key:"device-group-id" value:"defaultent-defaultsite-default"}} target:"defaultent"`:
			for _, upd := range a.Updates {
				switch updStr := strings.ReplaceAll(upd.Path.String(), "  ", " "); updStr {
				case `elem:{name:"description"} target:"defaultent"`:
					assert.Equal(t, `Global Default Site`, upd.GetVal().GetStringVal())
				case `elem:{name:"display-name"} target:"defaultent"`:
					assert.Equal(t, `Unknown Inventory`, upd.GetVal().GetStringVal())
				case `elem:{name:"ip-domain"} target:"defaultent"`:
					assert.Equal(t, `defaultent-defaultip`, upd.GetVal().GetStringVal())
				case `elem:{name:"traffic-class"} target:"defaultent"`:
					assert.Equal(t, `class-1`, upd.GetVal().GetStringVal())
				case `elem:{name:"mbr"} elem:{name:"uplink"} target:"defaultent"`:
					assert.Equal(t, uint64(1000000), upd.GetVal().GetUintVal())
				case `elem:{name:"mbr"} elem:{name:"downlink"} target:"defaultent"`:
					assert.Equal(t, uint64(1000000), upd.GetVal().GetUintVal())
				default:
					t.Errorf("unexpected update %s", updStr)
					t.FailNow()
				}
			}

		case `elem:{name:"site" key:{key:"site-id" value:"defaultent-defaultsite"}} elem:{name:"ip-domain" key:{key:"ip-domain-id" value:"defaultent-defaultip"}} target:"defaultent"`:
			for _, upd := range a.Updates {
				switch updStr := strings.ReplaceAll(upd.Path.String(), "  ", " "); updStr {
				case `elem:{name:"description"} target:"defaultent"`:
					assert.Equal(t, `Global Default IP Domain`, upd.GetVal().GetStringVal())
				case `elem:{name:"display-name"} target:"defaultent"`:
					assert.Equal(t, `Global Default IP Domain`, upd.GetVal().GetStringVal())
				case `elem:{name:"admin-status"} target:"defaultent"`:
					assert.Equal(t, `ENABLE`, upd.GetVal().GetStringVal())
				case `elem:{name:"dns-primary"} target:"defaultent"`:
					assert.Equal(t, `8.8.8.1`, upd.GetVal().GetStringVal())
				case `elem:{name:"dns-secondary"} target:"defaultent"`:
					assert.Equal(t, `8.8.8.2`, upd.GetVal().GetStringVal())
				case `elem:{name:"dnn"} target:"defaultent"`:
					assert.Equal(t, `dnnglobal`, upd.GetVal().GetStringVal())
				case `elem:{name:"ip-domain-id"} target:"defaultent"`:
					assert.Equal(t, `defaultent-defaultip`, upd.GetVal().GetStringVal())
				case `elem:{name:"mtu"} target:"defaultent"`:
					assert.Equal(t, uint64(57600), upd.GetVal().GetUintVal())
				case `elem:{name:"subnet"} target:"defaultent"`:
					assert.Equal(t, `192.168.0.0/24`, upd.GetVal().GetStringVal())
				default:
					t.Errorf("unexpected update %s", updStr)
					t.FailNow()
				}
			}

		default:
			//t.Errorf("unexpected update prefix %s", updPfx)
			//t.FailNow()
		}
	}

	// Expecting action 0-1 to be connectivity-service order is changeable
	for idx := 0; idx <= 1; idx++ {
		csAction := actions[idx]
		assert.Empty(t, csAction.DeletePrefix)
		assert.Len(t, csAction.Deletes, 1)
	}

}
