// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package gnmi implements a gnmi server to mock a device with YANG models.
package steps

import (
	"fmt"
	models_v1 "github.com/onosproject/config-models/modelplugin/aether-1.0.0/aether_1_0_0"
	models_v2 "github.com/onosproject/config-models/modelplugin/aether-2.0.0/aether_2_0_0"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/migration"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
)

var log = logging.GetLogger("migration.steps")

func MigrateV1V2ApnProfile(step migration.MigrationStep, fromTarget string, toTarget string, profile *models_v1.ApnProfile_ApnProfile_ApnProfile) (*migration.MigrationActions, error) {
	updates := []*gpb.Update{}
	updates = migration.AddUpdate(updates, migration.UpdateString("description", toTarget, profile.Description))
	updates = migration.AddUpdate(updates, migration.UpdateString("apn-name", toTarget, profile.ApnName))
	updates = migration.AddUpdate(updates, migration.UpdateString("dns-primary", toTarget, profile.DnsPrimary))
	updates = migration.AddUpdate(updates, migration.UpdateString("dns-secondary", toTarget, profile.DnsSecondary))
	updates = migration.AddUpdate(updates, migration.UpdateUInt32("mtu", toTarget, profile.Mtu))
	updates = migration.AddUpdate(updates, migration.UpdateBool("gx-enabled", toTarget, profile.GxEnabled))

	prefix := migration.StringToPath(fmt.Sprintf("apn-profile/apn-profile[id=%s]", *profile.Id), toTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{prefix}}, nil
}

func MigrateV1V2QosProfile(step migration.MigrationStep, fromTarget string, toTarget string, profile *models_v1.QosProfile_QosProfile_QosProfile) (*migration.MigrationActions, error) {
	updates := []*gpb.Update{}
	updates = migration.AddUpdate(updates, migration.UpdateString("description", toTarget, profile.Description))
	if profile.ApnAmbr != nil {
		updates = migration.AddUpdate(updates, migration.UpdateUInt32("apn-ambr/uplink", toTarget, profile.ApnAmbr.Uplink))
		updates = migration.AddUpdate(updates, migration.UpdateUInt32("apn-ambr/downlink", toTarget, profile.ApnAmbr.Downlink))
	}

	prefix := migration.StringToPath(fmt.Sprintf("qos-profile/qos-profile[id=%s]", *profile.Id), toTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{prefix}}, nil
}

func MigrateV1V2UpProfile(step migration.MigrationStep, fromTarget string, toTarget string, profile *models_v1.UpProfile_UpProfile_UpProfile) (*migration.MigrationActions, error) {
	updates := []*gpb.Update{}
	updates = migration.AddUpdate(updates, migration.UpdateString("description", toTarget, profile.Description))
	updates = migration.AddUpdate(updates, migration.UpdateString("user-plane", toTarget, profile.UserPlane))
	updates = migration.AddUpdate(updates, migration.UpdateString("access-control", toTarget, profile.AccessControl))

	prefix := migration.StringToPath(fmt.Sprintf("up-profile/up-profile[id=%s]", *profile.Id), toTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{prefix}}, nil
}

func MigrateV1V2AccessProfile(step migration.MigrationStep, fromTarget string, toTarget string, profile *models_v1.AccessProfile_AccessProfile_AccessProfile) (*migration.MigrationActions, error) {
	updates := []*gpb.Update{}
	updates = migration.AddUpdate(updates, migration.UpdateString("description", toTarget, profile.Description))
	updates = migration.AddUpdate(updates, migration.UpdateString("type", toTarget, profile.Type))
	updates = migration.AddUpdate(updates, migration.UpdateString("filter", toTarget, profile.Filter))

	prefix := migration.StringToPath(fmt.Sprintf("access-profile/access-profile[id=%s]", *profile.Id), toTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{prefix}}, nil
}

func MigrateV1V2(step migration.MigrationStep, fromTarget string, toTarget string, srcVal *gpb.TypedValue, destVal *gpb.TypedValue) ([]*migration.MigrationActions, error) {
	srcJsonBytes := srcVal.GetJsonVal()
	srcDevice := &models_v1.Device{}
	if len(srcJsonBytes) > 0 {
		if err := step.FromModels.Unmarshal(srcJsonBytes, srcDevice); err != nil {
			return nil, err
		}
	}

	destJsonBytes := destVal.GetJsonVal()
	destDevice := &models_v2.Device{}
	log.Infof("%v", destJsonBytes)
	if len(destJsonBytes) > 0 {
		if err := step.ToModels.Unmarshal(destJsonBytes, destDevice); err != nil {
			return nil, err
		}
	}

	log.Infof("Migrate src=%v, dest=%v", srcDevice, destDevice)

	allActions := []*migration.MigrationActions{}

	if srcDevice.ApnProfile != nil {
		for _, profile := range srcDevice.ApnProfile.ApnProfile {
			actions, err := MigrateV1V2ApnProfile(step, fromTarget, toTarget, profile)
			if err != nil {
				return nil, err
			}
			allActions = append(allActions, actions)
		}
	}

	if srcDevice.QosProfile != nil {
		for _, profile := range srcDevice.QosProfile.QosProfile {
			actions, err := MigrateV1V2QosProfile(step, fromTarget, toTarget, profile)
			if err != nil {
				return nil, err
			}
			allActions = append(allActions, actions)
		}
	}

	if srcDevice.UpProfile != nil {
		for _, profile := range srcDevice.UpProfile.UpProfile {
			actions, err := MigrateV1V2UpProfile(step, fromTarget, toTarget, profile)
			if err != nil {
				return nil, err
			}
			allActions = append(allActions, actions)
		}
	}

	if srcDevice.AccessProfile != nil {
		for _, profile := range srcDevice.AccessProfile.AccessProfile {
			actions, err := MigrateV1V2AccessProfile(step, fromTarget, toTarget, profile)
			if err != nil {
				return nil, err
			}
			allActions = append(allActions, actions)
		}
	}

	return allActions, nil
}
