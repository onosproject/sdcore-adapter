// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package steps

import (
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
	assert.Len(t, actions, 0)
}
