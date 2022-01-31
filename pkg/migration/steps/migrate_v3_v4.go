// SPDX-FileCopyrightText: 2021-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

/*
 * Implements the migration function from v3.0.0 aether models to v4.0.0 aether models. This
 * involves migrating each of the profiles (EnterpriseProfile, DeviceProfile, etc) and then
 * the UE.
 */

package steps

import (
	"fmt"
	modelsv3 "github.com/onosproject/config-models/modelplugin/aether-3.0.0/aether_3_0_0"
	modelsv4 "github.com/onosproject/config-models/modelplugin/aether-4.0.0/aether_4_0_0"
	"github.com/onosproject/sdcore-adapter/pkg/gnmiclient"
	"github.com/onosproject/sdcore-adapter/pkg/migration"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
)

// MigrateV3V4 - top level migration entry
func MigrateV3V4(step *migration.MigrationStep, fromTarget string, toTarget string, srcVal *gpb.TypedValue, destVal *gpb.TypedValue) ([]*migration.MigrationActions, error) {
	srcJSONBytes := srcVal.GetJsonVal()
	srcDevice := &modelsv3.Device{}

	if len(srcJSONBytes) > 0 {
		if err := step.FromModels.Unmarshal(srcJSONBytes, srcDevice); err != nil {
			return nil, err
		}
	}

	destJSONBytes := destVal.GetJsonVal()
	destDevice := &modelsv4.Device{}
	if len(destJSONBytes) > 0 {
		if err := step.ToModels.Unmarshal(destJSONBytes, destDevice); err != nil {
			return nil, err
		}
	}

	allActions := make([]*migration.MigrationActions, 0)

	if srcDevice.ConnectivityService != nil {
		for _, profile := range srcDevice.ConnectivityService.ConnectivityService {
			log.Infof("Migrating Connectivity Service %s", gnmiclient.StrDeref(profile.Id))
			actions, err := migrateV3V4ConnectivityService(fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, actions)
		}
	}

	if srcDevice.Enterprise != nil {
		for _, profile := range srcDevice.Enterprise.Enterprise {
			log.Infof("Migrating Enterprise %s", gnmiclient.StrDeref(profile.Id))
			action, err := migrateV3V4Enterprise(fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, action)
		}
	}

	if srcDevice.DeviceGroup != nil {
		for _, profile := range srcDevice.DeviceGroup.DeviceGroup {
			log.Infof("Migrating Device Group %s", gnmiclient.StrDeref(profile.Id))
			action, err := migrateV3V4DeviceGroup(fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, action)
		}
	}

	if srcDevice.TrafficClass != nil {
		for _, profile := range srcDevice.TrafficClass.TrafficClass {
			log.Infof("Migrating Traffic Class %s", gnmiclient.StrDeref(profile.Id))
			action, err := migrateV3V4TrafficClass(fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, action)
		}
	}

	if srcDevice.Application != nil {
		for _, profile := range srcDevice.Application.Application {
			log.Infof("Migrating Application Profile %s", gnmiclient.StrDeref(profile.Id))
			action, err := migrateV3V4Application(fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, action)
		}
	}

	if srcDevice.IpDomain != nil {
		for _, profile := range srcDevice.IpDomain.IpDomain {
			log.Infof("Migrating IpDomain Profile %s", gnmiclient.StrDeref(profile.Id))
			action, err := migrateV3V4IpDomain(fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, action)
		}
	}

	if srcDevice.Site != nil {
		for _, profile := range srcDevice.Site.Site {
			log.Infof("Migrating Site Profile %s", gnmiclient.StrDeref(profile.Id))
			action, err := migrateV3V4Site(fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, action)
		}
	}

	if srcDevice.ApList != nil {
		log.Warn("Migrating Access Points to Site")
		if srcDevice.Vcs != nil {
			for _, v := range srcDevice.Vcs.Vcs {
				for _, dg := range v.DeviceGroup {
					site := srcDevice.DeviceGroup.DeviceGroup[*dg.DeviceGroup].Site
					ap := srcDevice.ApList.ApList[*v.Ap]
					action, err := migrateV3V4ApListToSite(fromTarget, toTarget, site, ap)
					if err != nil {
						log.Warn(err.Error())
						continue
					}
					allActions = append(allActions, action)
				}
			}
		}
	}

	if srcDevice.Template != nil {
		for _, profile := range srcDevice.Template.Template {
			log.Infof("Migrating Template Profile %s", gnmiclient.StrDeref(profile.Id))
			action, err := migrateV3V4Template(fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, action)
		}
	}

	if srcDevice.Upf != nil {
		for _, profile := range srcDevice.Upf.Upf {
			log.Infof("Migrating Upf Profile %s", gnmiclient.StrDeref(profile.Id))
			action, err := migrateV3V4Upf(fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, action)
		}
	}

	if srcDevice.Vcs != nil {
		for _, profile := range srcDevice.Vcs.Vcs {
			log.Infof("Migrating Vcs Profile %s", gnmiclient.StrDeref(profile.Id))
			action, err := migrateV3V4Vcs(fromTarget, toTarget, profile, srcDevice.DeviceGroup)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, action)
		}
	}

	if srcDevice.DeviceGroup != nil {
		log.Warn("Migrating vcs mbr and traffic-class to Device Group")
		if srcDevice.Vcs != nil {
			for _, v := range srcDevice.Vcs.Vcs {
				for _, dg := range v.DeviceGroup {
					dgID := srcDevice.DeviceGroup.DeviceGroup[*dg.DeviceGroup].Id
					action, err := migrateV3V4VcsMbrTcToDG(fromTarget, toTarget, dgID, v)
					if err != nil {
						log.Warn(err.Error())
						continue
					}
					allActions = append(allActions, action)
				}
			}
		}
	}

	if srcDevice.Application != nil {
		log.Warn("Migrating vcs traffic-class to Application")
		if srcDevice.Vcs != nil {
			for _, v := range srcDevice.Vcs.Vcs {
				for _, ap := range v.Application {
					apID := ap.Application
					action, err := migrateV3V4VcsTcToApplication(fromTarget, toTarget, apID, v)
					if err != nil {
						log.Warn(err.Error())
						continue
					}
					allActions = append(allActions, action)
				}
			}
		}
	}

	if srcDevice.Upf != nil {
		log.Warn("Migrating vcs device-group's site to Upf")
		if srcDevice.Vcs != nil {
			for _, v := range srcDevice.Vcs.Vcs {
				var dgID *string
				for _, dg := range v.DeviceGroup {
					dgID = dg.DeviceGroup
					break
				}
				siteID := srcDevice.DeviceGroup.DeviceGroup[*dgID].Site
				upfID := v.Upf
				action, err := migrateV3V4VcssiteToUpf(fromTarget, toTarget, upfID, siteID)
				if err != nil {
					log.Warn(err.Error())
					continue
				}
				allActions = append(allActions, action)
			}
		}
	}
	return allActions, nil
}

func migrateV3V4ConnectivityService(fromTarget string, toTarget string, cs *modelsv3.ConnectivityService_ConnectivityService_ConnectivityService) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, cs.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, cs.DisplayName))
	// Not mapping spgwc-endpoint, hss-endpoint, pcrf-endpoint
	// Nothing to migrate for acc-prometheus-url
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("core-5g-endpoint", toTarget, cs.Core_5GEndpoint))

	prefix := gnmiclient.StringToPath(fmt.Sprintf("connectivity-service/connectivity-service[id=%s]", *cs.Id), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("connectivity-service/connectivity-service[id=%s]", *cs.Id), fromTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV3V4Enterprise(fromTarget string, toTarget string, ent *modelsv3.Enterprise_Enterprise_Enterprise) (*migration.MigrationActions, error) {
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

func migrateV3V4DeviceGroup(fromTarget string, toTarget string, dg *modelsv3.DeviceGroup_DeviceGroup_DeviceGroup) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, dg.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, dg.DisplayName))
	for _, dgi := range dg.Imsis {
		updStr := fmt.Sprintf("imsis[imsi-id=%s]/imsi-range-from", *dgi.Name)
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64(updStr, toTarget, dgi.ImsiRangeFrom))
		updStr = fmt.Sprintf("imsis[imsi-id=%s]/imsi-range-to", *dgi.Name)
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64(updStr, toTarget, dgi.ImsiRangeTo))
	}
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("site", toTarget, dg.Site))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("ip-domain", toTarget, dg.IpDomain))

	prefix := gnmiclient.StringToPath(fmt.Sprintf("device-group/device-group[id=%s]", *dg.Id), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("device-group/device-group[id=%s]", *dg.Id), fromTarget)
	deletePath.Origin = "site" // Used to build the "unchanged" attributes
	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil

}

func migrateV3V4TrafficClass(fromTarget string, toTarget string, tc *modelsv3.TrafficClass_TrafficClass_TrafficClass) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, tc.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, tc.DisplayName))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt8("qci", toTarget, tc.Qci))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateInt8("pelr", toTarget, tc.Pelr))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt16("pdb", toTarget, tc.Pdb))

	prefix := gnmiclient.StringToPath(fmt.Sprintf("traffic-class/traffic-class[id=%s]", *tc.Id), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("traffic-class/traffic-class[id=%s]", *tc.Id), fromTarget)
	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV3V4Application(fromTarget string, toTarget string, app *modelsv3.Application_Application_Application) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, app.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, app.DisplayName))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("enterprise", toTarget, app.Enterprise))
	for _, ap := range app.Endpoint {
		updStr := fmt.Sprintf("endpoint[endpoint-id=%s]/protocol", *ap.Name)
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString(updStr, toTarget, ap.Protocol))
		updStr = fmt.Sprintf("endpoint[endpoint-id=%s]/port-start", *ap.Name)
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt16(updStr, toTarget, ap.PortStart))
		updStr = fmt.Sprintf("endpoint[endpoint-id=%s]/port-end", *ap.Name)
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt16(updStr, toTarget, ap.PortEnd))
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("address", toTarget, ap.Address))
	}

	prefix := gnmiclient.StringToPath(fmt.Sprintf("application/application[id=%s]", *app.Id), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("application/application[id=%s]", *app.Id), fromTarget)
	deletePath.Origin = "enterprise" // Used to build the "unchanged" attributes

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV3V4IpDomain(fromTarget string, toTarget string, ipd *modelsv3.IpDomain_IpDomain_IpDomain) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, ipd.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, ipd.DisplayName))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("dnn", toTarget, ipd.Dnn))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("dns-primary", toTarget, ipd.DnsPrimary))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("dns-secondary", toTarget, ipd.DnsSecondary))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt16("mtu", toTarget, ipd.Mtu))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("subnet", toTarget, ipd.Subnet))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("admin-status", toTarget, ipd.AdminStatus))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("enterprise", toTarget, ipd.Enterprise))

	prefix := gnmiclient.StringToPath(fmt.Sprintf("ip-domain/ip-domain[id=%s]", *ipd.Id), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("ip-domain/ip-domain[id=%s]", *ipd.Id), fromTarget)
	deletePath.Origin = "enterprise,subnet" // Used to build the "unchanged" attributes

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV3V4Site(fromTarget string, toTarget string, st *modelsv3.Site_Site_Site) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, st.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, st.DisplayName))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("enterprise", toTarget, st.Enterprise))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("imsi-definition/mcc", toTarget, st.ImsiDefinition.Mcc))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("imsi-definition/mnc", toTarget, st.ImsiDefinition.Mnc))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("imsi-definition/format", toTarget, st.ImsiDefinition.Format))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt32("imsi-definition/enterprise", toTarget, st.ImsiDefinition.Enterprise))

	prefix := gnmiclient.StringToPath(fmt.Sprintf("site/site[id=%s]", *st.Id), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("site/site[id=%s]", *st.Id), fromTarget)
	deletePath.Origin = "enterprise" // Used to build the "unchanged" attributes

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV3V4ApListToSite(fromTarget string, toTarget string, site *string, apList *modelsv3.ApList_ApList_ApList) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	for _, aap := range apList.AccessPoints {
		updStr := fmt.Sprintf("small-cell[small-cell-id=%s]/address", *apList.Id)
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString(updStr, toTarget, aap.Address))
		updStr = fmt.Sprintf("small-cell[small-cell-id=%s]/tac", *apList.Id)
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString(updStr, toTarget, aap.Tac))
		updStr = fmt.Sprintf("small-cell[small-cell-id=%s]/enable", *apList.Id)
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateBool(updStr, toTarget, aap.Enable))
	}
	prefix := gnmiclient.StringToPath(fmt.Sprintf("site/site[id=%s]", *site), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("ap-list/ap-list[id=%s]", *apList.Id), fromTarget)
	deletePath.Origin = "enterprise" // Used to build the "unchanged" attributes
	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV3V4Template(fromTarget string, toTarget string, te *modelsv3.Template_Template_Template) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, te.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, te.DisplayName))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt8("sst", toTarget, te.Sst))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt32("sd", toTarget, te.Sd))
	defaultBehaviorDefault := "DENY-ALL"
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("default-behavior", toTarget, &defaultBehaviorDefault))

	uplink := uint64(*te.Uplink * 1000000)
	downlink := uint64(*te.Downlink * 1000000)
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64("slice/mbr/uplink", toTarget, &uplink))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64("slice/mbr/downlink", toTarget, &downlink))
	// nothing to migrate for downlink-burst-size or uplink-burst-size

	prefix := gnmiclient.StringToPath(fmt.Sprintf("template/template[id=%s]", *te.Id), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("template/template[id=%s]", *te.Id), fromTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV3V4Upf(fromTarget string, toTarget string, up *modelsv3.Upf_Upf_Upf) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, up.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, up.DisplayName))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("address", toTarget, up.Address))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt16("port", toTarget, up.Port))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("enterprise", toTarget, up.Enterprise))
	defaultSite := "defaultent-defaultsite"
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("site", toTarget, &defaultSite))

	//config-endpoint should be left blank
	prefix := gnmiclient.StringToPath(fmt.Sprintf("upf/upf[id=%s]", *up.Id), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("upf/upf[id=%s]", *up.Id), fromTarget)
	deletePath.Origin = "address,port,enterprise" // Used to build the "unchanged" attributes

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV3V4Vcs(fromTarget string, toTarget string, vc *modelsv3.Vcs_Vcs_Vcs, dgs *modelsv3.DeviceGroup_DeviceGroup) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, vc.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, vc.DisplayName))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("upf", toTarget, vc.Upf))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("enterprise", toTarget, vc.Enterprise))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt8("sst", toTarget, vc.Sst))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt32("sd", toTarget, vc.Sd))
	defaultBehaviorDefault := "DENY-ALL"
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("default-behavior", toTarget, &defaultBehaviorDefault))
	site := "defaultent-defaultsite" // In case the VCS has no DGs
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("site", toTarget, &site))

	// Get the site from the corresponding DG (if any)
	for _, vcd := range vc.DeviceGroup {
		updStr := fmt.Sprintf("device-group[device-group=%s]/enable", *vcd.DeviceGroup)
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateBool(updStr, toTarget, vcd.Enable))
		dgObject, ok := dgs.DeviceGroup[*vcd.DeviceGroup]
		if ok {
			site = *dgObject.Site
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("site", toTarget, &site))
		}
	}

	defaultPriority := uint8(5)
	for _, vcApp := range vc.Application {
		updStr := fmt.Sprintf("filter[application=%s]/allow", *vcApp.Application)
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateBool(updStr, toTarget, vcApp.Allow))
		updStr = fmt.Sprintf("filter[application=%s]/priority", *vcApp.Application)
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt8(updStr, toTarget, &defaultPriority))
	}

	uplink := uint64(*vc.Uplink * 1000000)
	downlink := uint64(*vc.Downlink * 1000000)
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64("slice/mbr/uplink", toTarget, &uplink))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64("slice/mbr/downlink", toTarget, &downlink))

	prefix := gnmiclient.StringToPath(fmt.Sprintf("vcs/vcs[id=%s]", *vc.Id), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("vcs/vcs[id=%s]", *vc.Id), fromTarget)
	deletePath.Origin = "enterprise,sst,sd,traffic-class" // Used to build the "unchanged" attributes

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV3V4VcsMbrTcToDG(fromTarget string, toTarget string, dg *string, vc *modelsv3.Vcs_Vcs_Vcs) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	uplink := uint64(*vc.Uplink * 1000000)
	downlink := uint64(*vc.Downlink * 1000000)
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64("device/mbr/uplink", toTarget, &uplink))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64("device/mbr/downlink", toTarget, &downlink))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("device/traffic-class", toTarget, vc.TrafficClass))
	prefix := gnmiclient.StringToPath(fmt.Sprintf("device-group/device-group[id=%s]", *dg), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("device-group/device-group[id=%s]", *dg), fromTarget)
	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV3V4VcsTcToApplication(fromTarget string, toTarget string, ap *string, vc *modelsv3.Vcs_Vcs_Vcs) (*migration.MigrationActions, error) {
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("application/application[id=%s]", *ap), fromTarget)
	return &migration.MigrationActions{Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV3V4VcssiteToUpf(fromTarget string, toTarget string, upfID *string, siteID *string) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("site", toTarget, siteID))
	prefix := gnmiclient.StringToPath(fmt.Sprintf("upf/upf[id=%s]", *upfID), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("upf/upf[id=%s]", *upfID), fromTarget)
	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}
