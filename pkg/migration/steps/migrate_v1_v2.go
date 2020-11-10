// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

/*
 * Implements the migration function from v1 aether models to v2 aether models. This
 * involves migrating each of the profiles (APNProfile, AccessProfile, etc) and then
 * the UE.
 */

package steps

import (
	"fmt"
	"github.com/google/uuid"
	models_v1 "github.com/onosproject/config-models/modelplugin/aether-1.0.0/aether_1_0_0"
	models_v2 "github.com/onosproject/config-models/modelplugin/aether-2.0.0/aether_2_0_0"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/migration"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"strconv"
	"strings"
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
	deletePath := migration.StringToPath(fmt.Sprintf("apn-profile/apn-profile[id=%s]", *profile.Id), fromTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func MigrateV1V2QosProfile(step migration.MigrationStep, fromTarget string, toTarget string, profile *models_v1.QosProfile_QosProfile_QosProfile) (*migration.MigrationActions, error) {
	updates := []*gpb.Update{}
	updates = migration.AddUpdate(updates, migration.UpdateString("description", toTarget, profile.Description))
	if profile.ApnAmbr != nil {
		updates = migration.AddUpdate(updates, migration.UpdateUInt32("apn-ambr/uplink", toTarget, profile.ApnAmbr.Uplink))
		updates = migration.AddUpdate(updates, migration.UpdateUInt32("apn-ambr/downlink", toTarget, profile.ApnAmbr.Downlink))
	}

	prefix := migration.StringToPath(fmt.Sprintf("qos-profile/qos-profile[id=%s]", *profile.Id), toTarget)
	deletePath := migration.StringToPath(fmt.Sprintf("qos-profile/qos-profile[id=%s]", *profile.Id), fromTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func MigrateV1V2UpProfile(step migration.MigrationStep, fromTarget string, toTarget string, profile *models_v1.UpProfile_UpProfile_UpProfile) (*migration.MigrationActions, error) {
	updates := []*gpb.Update{}
	updates = migration.AddUpdate(updates, migration.UpdateString("description", toTarget, profile.Description))
	updates = migration.AddUpdate(updates, migration.UpdateString("user-plane", toTarget, profile.UserPlane))
	updates = migration.AddUpdate(updates, migration.UpdateString("access-control", toTarget, profile.AccessControl))

	prefix := migration.StringToPath(fmt.Sprintf("up-profile/up-profile[id=%s]", *profile.Id), toTarget)
	deletePath := migration.StringToPath(fmt.Sprintf("up-profile/up-profile[id=%s]", *profile.Id), fromTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func MigrateV1V2AccessProfile(step migration.MigrationStep, fromTarget string, toTarget string, profile *models_v1.AccessProfile_AccessProfile_AccessProfile) (*migration.MigrationActions, error) {
	updates := []*gpb.Update{}
	updates = migration.AddUpdate(updates, migration.UpdateString("description", toTarget, profile.Description))
	updates = migration.AddUpdate(updates, migration.UpdateString("type", toTarget, profile.Type))
	updates = migration.AddUpdate(updates, migration.UpdateString("filter", toTarget, profile.Filter))

	prefix := migration.StringToPath(fmt.Sprintf("access-profile/access-profile[id=%s]", *profile.Id), toTarget)
	deletePath := migration.StringToPath(fmt.Sprintf("access-profile/access-profile[id=%s]", *profile.Id), fromTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

// Parse a V1 UEID and return the first and last IMSIs in the range
func ParseV1UEID(s string) (uint64, uint64, error) {
	// TODO: Bug in onos-config causes strings that are all digits to be
	// converted into integers. I've been using the workaround of substituting
	// an "e" for the first "3" in the IMSI.
	s = strings.Replace(s, "e", "3", -1)

	if strings.Contains(s, "-") {
		parts := strings.SplitN(s, "-", 2)

		from, err := strconv.ParseUint(parts[0], 0, 64)
		if err != nil {
			return 0, 0, err
		}

		to, err := strconv.ParseUint(parts[1], 0, 64)
		if err != nil {
			return 0, 0, err
		}

		return from, to, nil
	} else {
		imsi, err := strconv.ParseUint(s, 0, 64)
		if err != nil {
			return 0, 0, err
		}

		return imsi, imsi, nil
	}
}

// Given a V1 UE, check and see if the V2 models already contain a UE
// that matches the keys. If so, then return the UUID of the V2 model.
func FindExistingUE(destDevice *models_v2.Device, ue *models_v1.AetherSubscriber_Subscriber_Ue) (string, error) {
	if destDevice.Subscriber == nil {
		// there is nothing to search
		return "", nil
	}

	if ue.Ueid == nil {
		// the device we're searching from doesn't have a ueid
		// this probably can't happen...
		return "", nil
	}

	first, last, err := ParseV1UEID(*ue.Ueid)
	if err != nil {
		return "", err
	}

	for _, candidateUe := range destDevice.Subscriber.Ue {
		if candidateUe.ImsiRangeFrom == nil {
			continue
		}
		if candidateUe.ImsiRangeTo == nil {
			continue
		}
		if (*candidateUe.ImsiRangeFrom == first) && (*candidateUe.ImsiRangeTo == last) {
			return *candidateUe.Id, nil
		}
	}

	return "", nil
}

func MigrateV1V2Subscriber(step migration.MigrationStep, fromTarget string, toTarget string, ue *models_v1.AetherSubscriber_Subscriber_Ue, destDevice *models_v2.Device) (*migration.MigrationActions, error) {
	updates := []*gpb.Update{}
	updates = migration.AddUpdate(updates, migration.UpdateUInt32("priority", toTarget, ue.Priority))
	updates = migration.AddUpdate(updates, migration.UpdateBool("enabled", toTarget, ue.Enabled))
	updates = migration.AddUpdate(updates, migration.UpdateString("profiles/apn-profile", toTarget, ue.Profiles.ApnProfile))
	updates = migration.AddUpdate(updates, migration.UpdateString("profiles/qos-profile", toTarget, ue.Profiles.QosProfile))
	updates = migration.AddUpdate(updates, migration.UpdateString("profiles/up-profile", toTarget, ue.Profiles.UpProfile))

	updates = migration.AddUpdate(updates, migration.UpdateString("requested-apn", toTarget, ue.RequestedApn))
	if ue.ServingPlmn != nil {
		updates = migration.AddUpdate(updates, migration.UpdateUInt32("serving-plmn/mcc", toTarget, ue.ServingPlmn.Mcc))
		updates = migration.AddUpdate(updates, migration.UpdateUInt32("serving-plmn/mnc", toTarget, ue.ServingPlmn.Mnc))
		updates = migration.AddUpdate(updates, migration.UpdateUInt32("serving-plmn/tac", toTarget, ue.ServingPlmn.Tac))
	}
	if ue.Ueid != nil {
		first, last, err := ParseV1UEID(*ue.Ueid)
		if err != nil {
			return nil, err
		}

		// TODO: Compensates for a bug in aether-config, by exploiting a different bug in aether-config
		firstStr := strconv.FormatUint(first, 10)
		lastStr := strconv.FormatUint(last, 10)
		updates = migration.AddUpdate(updates, migration.UpdateString("imsi-range-from", toTarget, &firstStr))
		updates = migration.AddUpdate(updates, migration.UpdateString("imsi-range-to", toTarget, &lastStr))
	}

	/*
	 * V1 Aether models had a different key (imsi range) than V2 aether models (uuid). So what we do is
	 * to use the v1 imsi to see if the object already exists in the v2 device. If it does, then we know
	 * a migration has been previously partially completed. If no such imsi exists, then we know the
	 * object has not been previously migrated, and we generate a new uuid for the new v2 object.
	 */

	ueUuid, err := FindExistingUE(destDevice, ue)
	if err != nil {
		return nil, err
	}
	if ueUuid == "" {
		ueUuid = uuid.New().String()
	}

	prefix := migration.StringToPath(fmt.Sprintf("subscriber/ue[id=%s]", ueUuid), toTarget)

	deletePrefix := migration.StringToPath(fmt.Sprintf("subscriber/ue[ueid=%s]", *ue.Ueid), fromTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePrefix}}, nil
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

	if srcDevice.Subscriber != nil {
		for _, ue := range srcDevice.Subscriber.Ue {
			actions, err := MigrateV1V2Subscriber(step, fromTarget, toTarget, ue, destDevice)
			if err != nil {
				return nil, err
			}
			allActions = append(allActions, actions)
		}
	}

	return allActions, nil
}
