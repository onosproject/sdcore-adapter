// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package gnmi implements a gnmi server to mock a device with YANG models.
package steps

import (
	models_v1 "github.com/onosproject/config-models/modelplugin/aether-1.0.0/aether_1_0_0"
	models_v2 "github.com/onosproject/config-models/modelplugin/aether-2.0.0/aether_2_0_0"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/migration"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
)

var log = logging.GetLogger("migration.steps")

func MigrateV1V2(step migration.MigrationStep, toTarget string, srcVal *gpb.TypedValue, destVal *gpb.TypedValue) error {
	srcJsonBytes := srcVal.GetJsonVal()
	srcDevice := &models_v1.Device{}
	if len(srcJsonBytes) > 0 {
		if err := step.FromModels.Unmarshal(srcJsonBytes, srcDevice); err != nil {
			return err
		}
	}

	destJsonBytes := destVal.GetJsonVal()
	destDevice := &models_v2.Device{}
	log.Infof("%v", destJsonBytes)
	if len(destJsonBytes) > 0 {
		if err := step.ToModels.Unmarshal(destJsonBytes, destDevice); err != nil {
			return err
		}
	}

	log.Infof("Migrate src=%v, dest=%v", srcDevice, destDevice)

	return nil
}
