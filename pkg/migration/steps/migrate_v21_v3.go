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
	"fmt"
	modelsv21 "github.com/onosproject/config-models/modelplugin/aether-2.1.0/aether_2_1_0"
	modelsv3 "github.com/onosproject/config-models/modelplugin/aether-3.0.0/aether_3_0_0"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/gnmiclient"
	"github.com/onosproject/sdcore-adapter/pkg/migration"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"strconv"
	"strings"
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

	if srcDevice.ConnectivityService != nil {
		for _, profile := range srcDevice.ConnectivityService.ConnectivityService {
			log.Infof("Migrating Connectivity Service %s", gnmiclient.StrDeref(profile.Id))
			actions, err := migrateV21V3ConnectivityService(fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, actions)
		}
	}

	if srcDevice.Enterprise != nil {
		action := createDefaultEnterprise(fromTarget, toTarget)
		var err error
		allActions = append(allActions, action)
		for _, profile := range srcDevice.Enterprise.Enterprise {
			log.Infof("Migrating Enterprise %s", gnmiclient.StrDeref(profile.Id))
			action, err = migrateV21V3Enterprise(fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, action)
		}
	}

	if srcDevice.UpProfile != nil {
		for _, profile := range srcDevice.UpProfile.UpProfile {
			log.Infof("Migrating Up Profile %s", gnmiclient.StrDeref(profile.Id))
			action, err := migrateV21V3UpProfile(fromTarget, toTarget, profile, srcDevice.Subscriber)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, action)
		}
	}

	if srcDevice.AccessProfile != nil {
		log.Warn("Migrating AccessProfile is not supported")
	}

	if srcDevice.ApnProfile != nil {
		for _, profile := range srcDevice.ApnProfile.ApnProfile {
			log.Infof("Migrating APN Profile %s", gnmiclient.StrDeref(profile.Id))
			action, err := migrateV21V3ApnProfileToIPDomain(fromTarget, toTarget, profile, srcDevice.Subscriber)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, action)
		}
	}

	if srcDevice.QosProfile != nil {
		for _, profile := range srcDevice.QosProfile.QosProfile {
			log.Infof("Migrating QOS Profile %s", gnmiclient.StrDeref(profile.Id))
			action, err := migrateV21V3QosProfileToTrafficClass(fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
			} else {
				allActions = append(allActions, action)
			}
			action, err = migrateV21V3QosProfileToTemplate(fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, action)
		}
	}

	if srcDevice.Subscriber != nil {
		for _, profile := range srcDevice.Subscriber.Ue {
			log.Infof("Migrating Subscriber Ue %s", gnmiclient.StrDeref(profile.Id))
			action, err := migrateV21V3SubscriberToSite(fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
			} else {
				allActions = append(allActions, action)
			}
			action, err = migrateV21V3SubscriberToDeviceGroup(fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, action)
		}
	}

	if srcDevice.SecurityProfile != nil {
		log.Warn("Migrating SecurityProfile is not supported")
	}

	if srcDevice.ServiceGroup != nil {
		log.Warn("Migrating ServiceGroup is not supported")
	}

	if srcDevice.ServicePolicy != nil {
		log.Warn("Migrating ServicePolicy is not supported")
	}

	if srcDevice.ServiceRule != nil {
		log.Warn("Migrating ServiceRule is not supported")
	}

	return allActions, nil
}

func migrateV21V3SubscriberToSite(fromTarget string, toTarget string, ue *modelsv21.AetherSubscriber_Subscriber_Ue) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, ue.DisplayName))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("enterprise", toTarget, ue.Enterprise))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt32("imsi-definition/mcc", toTarget, ue.ServingPlmn.Mcc))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt32("imsi-definition/mnc", toTarget, ue.ServingPlmn.Mnc))
	genericFormat := "SSSSSSSSSSSSSSS"
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("imsi-definition/format", toTarget, &genericFormat))

	prefix := gnmiclient.StringToPath(fmt.Sprintf("site/site[id=%s]", truncateUUID(ue.Id)), toTarget) // Only 32 chars allowed
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("subscriber/ue[id=%s]", *ue.Id), fromTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV21V3SubscriberToDeviceGroup(fromTarget string, toTarget string, ue *modelsv21.AetherSubscriber_Subscriber_Ue) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	if ue.ImsiRangeFrom == nil {
		return nil, errors.NewInternal("UE %s has no Imsi ranges", *ue.Id)
	}
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, ue.DisplayName))
	truncatedSiteUUID := truncateUUID(ue.Id)
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("site", toTarget, &truncatedSiteUUID))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("ip-domain", toTarget, ue.Profiles.ApnProfile))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64("imsis[name=range-1]/imsi-range-from", toTarget, ue.ImsiRangeFrom))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64("imsis[name=range-1]/imsi-range-to", toTarget, ue.ImsiRangeTo))

	prefix := gnmiclient.StringToPath(fmt.Sprintf("device-group/device-group[id=%s]", truncateUUID(ue.Id)), toTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{}}, nil
}

func migrateV21V3ApnProfileToIPDomain(fromTarget string, toTarget string, profile *modelsv21.ApnProfile_ApnProfile_ApnProfile,
	subscriber *modelsv21.AetherSubscriber_Subscriber) (*migration.MigrationActions, error) {

	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, profile.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, profile.DisplayName))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("dns-primary", toTarget, profile.DnsPrimary))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("dns-secondary", toTarget, profile.DnsSecondary))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("dnn", toTarget, profile.ApnName))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt32("mtu", toTarget, profile.Mtu))
	status := "DISABLE"
	if profile.GxEnabled != nil && *profile.GxEnabled {
		status = "ENABLE"
	}
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("admin-status", toTarget, &status))
	defaultSubnet := "255.255.255.255/32"
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("subnet", toTarget, &defaultSubnet))

	// Look for the UE where this is used and extract its enterprise
	enterprise := "defaultent"
	for _, ue := range subscriber.Ue {
		if ue.Enterprise == nil || ue.Profiles.ApnProfile == nil || *ue.Profiles.ApnProfile != *profile.Id {
			continue
		}
		enterprise = *ue.Enterprise
	}
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("enterprise", toTarget, &enterprise))

	prefix := gnmiclient.StringToPath(fmt.Sprintf("ip-domain/ip-domain[id=%s]", *profile.Id), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("apn-profile/apn-profile[id=%s]", *profile.Id), fromTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV21V3QosProfileToTemplate(fromTarget string, toTarget string, profile *modelsv21.QosProfile_QosProfile_QosProfile) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, profile.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, profile.DisplayName))
	if profile.ApnAmbr.Uplink != nil {
		uplink := (*profile.ApnAmbr.Uplink) / 1e6 // Floor, not round
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt32("uplink", toTarget, &uplink))
	}
	if profile.ApnAmbr.Downlink != nil {
		downlink := (*profile.ApnAmbr.Downlink) / 1e6 // Floor, not round
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt32("downlink", toTarget, &downlink))
	}
	if profile.Qci != nil {
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("traffic-class", toTarget, profile.Id))
	}

	prefix := gnmiclient.StringToPath(fmt.Sprintf("template/template[id=%s]", *profile.Id), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("qos-profile/qos-profile[id=%s]", *profile.Id), fromTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV21V3QosProfileToTrafficClass(fromTarget string, toTarget string, profile *modelsv21.QosProfile_QosProfile_QosProfile) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	if profile.Qci == nil {
		return nil, errors.NewInternal("Not creating Traffic-Class from QOS Profile %s - no qci", *profile.Id)
	}
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, profile.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, profile.DisplayName))
	scaledQci := (*profile.Qci) / 3 // Old range is 0-85 - new range is 0-32
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt32("qci", toTarget, &scaledQci))

	prefix := gnmiclient.StringToPath(fmt.Sprintf("traffic-class/traffic-class[id=%s]", *profile.Id), toTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{}}, nil
}

func migrateV21V3UpProfile(fromTarget string, toTarget string, profile *modelsv21.UpProfile_UpProfile_UpProfile,
	subscriber *modelsv21.AetherSubscriber_Subscriber) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, profile.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, profile.DisplayName))
	if profile.UserPlane != nil {
		upAddress := *profile.UserPlane
		upAddrParts := strings.Split(upAddress, ":")
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("address", toTarget, &upAddrParts[0]))
		port := uint32(8080)
		if len(upAddrParts) == 2 {
			portInt, err := strconv.ParseInt(upAddrParts[1], 10, 16)
			if err != nil {
				log.Warn("unable to parse port in %s for UP Profile %s", upAddrParts[1])
			} else {
				port = uint32(portInt)
			}
		}
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt32("port", toTarget, &port))
	}

	// Look for the UE where this is used and extract its enterprise
	enterprise := "default-enterprise"
	for _, ue := range subscriber.Ue {
		if ue.Enterprise == nil || ue.Profiles.UpProfile == nil || *ue.Profiles.UpProfile != *profile.Id {
			continue
		}
		enterprise = *ue.Enterprise
	}
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("enterprise", toTarget, &enterprise))

	prefix := gnmiclient.StringToPath(fmt.Sprintf("upf/upf[id=%s]", *profile.Id), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("up-profile/up-profile[id=%s]", *profile.Id), fromTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV21V3ConnectivityService(fromTarget string, toTarget string, cs *modelsv21.ConnectivityService_ConnectivityService_ConnectivityService) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, cs.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, cs.DisplayName))
	// Only copy over the Description and ID - the other endpoints
	// do not have a role in the 3.0.0 models even for 4G
	prefix := gnmiclient.StringToPath(fmt.Sprintf("connectivity-service/connectivity-service[id=%s]", *cs.Id), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("connectivity-service/connectivity-service[id=%s]", *cs.Id), fromTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func createDefaultEnterprise(fromTarget string, toTarget string) *migration.MigrationActions {
	var updates []*gpb.Update
	dispName := "Default Enterprise"
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, &dispName))

	prefix := gnmiclient.StringToPath("enterprise/enterprise[id=defaultent]", toTarget)
	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{}}
}

func migrateV21V3Enterprise(fromTarget string, toTarget string, ent *modelsv21.Enterprise_Enterprise_Enterprise) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, ent.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, ent.DisplayName))
	for _, cs := range ent.ConnectivityService {
		updStr := fmt.Sprintf("connectivity-service[connectivity-service=%s]/enabled", *cs.ConnectivityService)
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateBool(updStr, toTarget, cs.Enabled))
	}

	prefix := gnmiclient.StringToPath(fmt.Sprintf("enterprise/enterprise[id=%s]", *ent.Id), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("enterprise/enterprise[id=%s]", *ent.Id), fromTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func truncateUUID(uuid *string) string {
	if uuid == nil {
		return ""
	}
	firstChar := (*uuid)[:1]
	// if first char is numeric, replace it with letter
	_, err := strconv.ParseInt(firstChar, 10, 8)
	if err != nil {
		return (*uuid)[:32]
	}
	return fmt.Sprintf("a%s", *uuid)[:32]
}
