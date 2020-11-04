// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package gnmi implements a gnmi server to mock a device with YANG models.
package steps

import (
	"context"
	"fmt"
	models_v1 "github.com/onosproject/config-models/modelplugin/aether-1.0.0/aether_1_0_0"
	models_v2 "github.com/onosproject/config-models/modelplugin/aether-2.0.0/aether_2_0_0"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/migration"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
)

var log = logging.GetLogger("migration.steps")

func PathUpdateString(path string, target string, val *string) *gpb.Update {
	if val == nil {
		return nil
	}

	return &gpb.Update{
		Path: migration.StringToPath(path, target),
		Val:  &gpb.TypedValue{Value: &gpb.TypedValue_StringVal{StringVal: *val}},
	}
}

func AddUpdate(updates []*gpb.Update, update *gpb.Update) []*gpb.Update {
	if update != nil {
		updates = append(updates, update)
	}
	return updates
}

func MigrateV1V2APNProfile(step migration.MigrationStep, toTarget string, profile *models_v1.ApnProfile_ApnProfile_ApnProfile) error {
	updates := []*gpb.Update{}
	updates = AddUpdate(updates, PathUpdateString("apn-name", toTarget, profile.ApnName))

	prefix := migration.StringToPathWithKeys(fmt.Sprintf("apn-profile/apn-profile[id=%s]", *profile.Id), toTarget)

	err := migration.Update(prefix, toTarget, step.Migrator.AetherConfigAddr, updates, context.Background())

	if err != nil {
		return err
	}

	//deletes := []*gdb.Path{StringToPath("app-name")}

	err = migration.Delete(nil, toTarget, step.Migrator.AetherConfigAddr, []*gpb.Path{prefix} /*deletes,*/, context.Background())

	return nil

	/*
		ApnName:      apn.ApnName,
		DnsPrimary:   apn.DnsPrimary,
		DnsSecondary: apn.DnsSecondary,
		Mtu:          apn.Mtu,
		GxEnabled:    apn.GxEnabled,
	*/
}

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

	if srcDevice.ApnProfile != nil {
		for _, apn := range srcDevice.ApnProfile.ApnProfile {
			err := MigrateV1V2APNProfile(step, toTarget, apn)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
