// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

/*
 * Implements the migration function from v2.1.0 aether models to v3.0.0 aether models. This
 * involves migrating each of the profiles (APNProfile, AccessProfile, etc) and then
 * the UE.
 */

package steps

import (
	modelsv21 "github.com/onosproject/config-models/modelplugin/aether-2.1.0/aether_2_1_0"
	modelsv3 "github.com/onosproject/config-models/modelplugin/aether-3.0.0/aether_3_0_0"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/migration"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
)

var log = logging.GetLogger("migration.steps")

// MigrateV21V3 - top level migration entry
func MigrateV21V3(step *migration.MigrationStep, fromTarget string, toTarget string, srcVal *gpb.TypedValue, destVal *gpb.TypedValue) ([]*migration.MigrationActions, error) {
	srcJSONBytes := srcVal.GetJsonVal()
	srcDevice := &modelsv21.Device{}
	if len(srcJSONBytes) > 0 {
		if err := step.FromModels.Unmarshal(srcJSONBytes, srcDevice); err != nil {
			return nil, err
		}
	}

	destJSONBytes := destVal.GetJsonVal()
	destDevice := &modelsv3.Device{}
	if len(destJSONBytes) > 0 {
		if err := step.ToModels.Unmarshal(destJSONBytes, destDevice); err != nil {
			return nil, err
		}
	}

	allActions := make([]*migration.MigrationActions, 0)
	if srcDevice.AccessProfile != nil {
		for _, profile := range srcDevice.AccessProfile.AccessProfile {
			log.Infof("Migrating Access Profile %s", migration.StrDeref(profile.Id))
			actions, err := migrateV21V3AccessProfile(step, fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, actions)
		}
	}

	if srcDevice.Subscriber != nil {
		for _, profile := range srcDevice.Subscriber.Ue {
			log.Infof("Migrating Subscriber Ue %s", migration.StrDeref(profile.Id))
			actions, err := migrateV21V3Subscriber(step, fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, actions)
		}
	}

	if srcDevice.ApnProfile != nil {
		for _, profile := range srcDevice.ApnProfile.ApnProfile {
			log.Infof("Migrating APN Profile %s", migration.StrDeref(profile.Id))
			actions, err := migrateV21V3ApnProfile(step, fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, actions)
		}
	}

	if srcDevice.QosProfile != nil {
		for _, profile := range srcDevice.QosProfile.QosProfile {
			log.Infof("Migrating QOS Profile %s", migration.StrDeref(profile.Id))
			actions, err := migrateV21V3QosProfile(step, fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, actions)
		}
	}

	if srcDevice.SecurityProfile != nil {
		for _, profile := range srcDevice.SecurityProfile.SecurityProfile {
			log.Infof("Migrating Security Profile %s", migration.StrDeref(profile.Id))
			actions, err := migrateV21V3SecurityProfile(step, fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, actions)
		}
	}

	if srcDevice.ServiceGroup != nil {
		for _, profile := range srcDevice.ServiceGroup.ServiceGroup {
			log.Infof("Migrating Service Group %s", migration.StrDeref(profile.Id))
			actions, err := migrateV21V3ServiceGroup(step, fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, actions)
		}
	}

	if srcDevice.ServicePolicy != nil {
		for _, profile := range srcDevice.ServicePolicy.ServicePolicy {
			log.Infof("Migrating Service Policy %s", migration.StrDeref(profile.Id))
			actions, err := migrateV21V3ServicePolicy(step, fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, actions)
		}
	}

	if srcDevice.ServiceRule != nil {
		for _, profile := range srcDevice.ServiceRule.ServiceRule {
			log.Infof("Migrating Service Rule %s", migration.StrDeref(profile.Id))
			actions, err := migrateV21V3ServiceRule(step, fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, actions)
		}
	}

	if srcDevice.UpProfile != nil {
		for _, profile := range srcDevice.UpProfile.UpProfile {
			log.Infof("Migrating Up Profile %s", migration.StrDeref(profile.Id))
			actions, err := migrateV21V3UpProfile(step, fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, actions)
		}
	}

	if srcDevice.ConnectivityService != nil {
		for _, profile := range srcDevice.ConnectivityService.ConnectivityService {
			log.Infof("Migrating Connectivity Service %s", migration.StrDeref(profile.Id))
			actions, err := migrateV21V3ConnectivityService(step, fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, actions)
		}
	}

	if srcDevice.Enterprise != nil {
		for _, profile := range srcDevice.Enterprise.Enterprise {
			log.Infof("Migrating Enterprise %s", migration.StrDeref(profile.Id))
			actions, err := migrateV21V3Enterprise(step, fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, actions)
		}
	}

	return allActions, nil
}

func migrateV21V3AccessProfile(step *migration.MigrationStep, fromTarget string, toTarget string, profile *modelsv21.AccessProfile_AccessProfile_AccessProfile) (*migration.MigrationActions, error) {
	return nil, errors.NewNotSupported("migrateV21V3AccessProfile not yet supported")
}

func migrateV21V3Subscriber(step *migration.MigrationStep, fromTarget string, toTarget string, profile *modelsv21.AetherSubscriber_Subscriber_Ue) (*migration.MigrationActions, error) {
	return nil, errors.NewNotSupported("migrateV21V3Subscriber not yet supported")
}

func migrateV21V3ApnProfile(step *migration.MigrationStep, fromTarget string, toTarget string, profile *modelsv21.ApnProfile_ApnProfile_ApnProfile) (*migration.MigrationActions, error) {
	return nil, errors.NewNotSupported("migrateV21V3ApnProfile not yet supported")
}

func migrateV21V3QosProfile(step *migration.MigrationStep, fromTarget string, toTarget string, profile *modelsv21.QosProfile_QosProfile_QosProfile) (*migration.MigrationActions, error) {
	return nil, errors.NewNotSupported("migrateV21V3QosProfile not yet supported")
}

func migrateV21V3SecurityProfile(step *migration.MigrationStep, fromTarget string, toTarget string, profile *modelsv21.SecurityProfile_SecurityProfile_SecurityProfile) (*migration.MigrationActions, error) {
	return nil, errors.NewNotSupported("migrateV21V3SecurityProfile not yet supported")
}

func migrateV21V3ServiceGroup(step *migration.MigrationStep, fromTarget string, toTarget string, profile *modelsv21.ServiceGroup_ServiceGroup_ServiceGroup) (*migration.MigrationActions, error) {
	return nil, errors.NewNotSupported("migrateV21V3ServiceGroup not yet supported")
}

func migrateV21V3ServicePolicy(step *migration.MigrationStep, fromTarget string, toTarget string, profile *modelsv21.ServicePolicy_ServicePolicy_ServicePolicy) (*migration.MigrationActions, error) {
	return nil, errors.NewNotSupported("migrateV21V3ServicePolicy not yet supported")
}

func migrateV21V3ServiceRule(step *migration.MigrationStep, fromTarget string, toTarget string, profile *modelsv21.ServiceRule_ServiceRule_ServiceRule) (*migration.MigrationActions, error) {
	return nil, errors.NewNotSupported("migrateV21V3ServiceRule not yet supported")
}

func migrateV21V3UpProfile(step *migration.MigrationStep, fromTarget string, toTarget string, profile *modelsv21.UpProfile_UpProfile_UpProfile) (*migration.MigrationActions, error) {
	return nil, errors.NewNotSupported("migrateV21V3UpProfile not yet supported")
}

func migrateV21V3ConnectivityService(step *migration.MigrationStep, fromTarget string, toTarget string, profile *modelsv21.ConnectivityService_ConnectivityService_ConnectivityService) (*migration.MigrationActions, error) {
	return nil, errors.NewNotSupported("migrateV21V3ConnectivityService not yet supported")
}

func migrateV21V3Enterprise(step *migration.MigrationStep, fromTarget string, toTarget string, profile *modelsv21.Enterprise_Enterprise_Enterprise) (*migration.MigrationActions, error) {
	return nil, errors.NewNotSupported("migrateV21V3Enterprise not yet supported")
}
