// SPDX-FileCopyrightText: 2021-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

/*
 * Implements the migration function from v4.0.0 aether models to v2.0.0 aether models. This
 * involves migrating each of the profiles (EnterpriseProfile, DeviceProfile, etc) and then
 * the UE.
 */

package steps

import (
	"fmt"
	modelsv2 "github.com/onosproject/config-models/modelplugin/aether-2.0.0/aether_2_0_0"
	modelsv4 "github.com/onosproject/config-models/modelplugin/aether-4.0.0/aether_4_0_0"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/gnmiclient"
	"github.com/onosproject/sdcore-adapter/pkg/migration"
	"github.com/onosproject/sdcore-adapter/pkg/synchronizer"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
)

var log = logging.GetLogger("migration.steps")

// MigrateV4V2 - top level migration entry
func MigrateV4V2(step *migration.MigrationStep, fromTarget string, toTarget string, srcVal *gpb.TypedValue, destVal *gpb.TypedValue) ([]*migration.MigrationActions, error) {
	srcJSONBytes := srcVal.GetJsonVal()
	srcDevice := &modelsv4.Device{}

	if len(srcJSONBytes) > 0 {
		if err := step.FromModels.Unmarshal(srcJSONBytes, srcDevice); err != nil {
			return nil, err
		}
	}

	destJSONBytes := destVal.GetJsonVal()
	destDevice := &modelsv2.Device{}
	if len(destJSONBytes) > 0 {
		if err := step.ToModels.Unmarshal(destJSONBytes, destDevice); err != nil {
			return nil, err
		}
	}

	allActions := make([]*migration.MigrationActions, 0)

	//migrating Connectivity Service to Connectivity Service
	if srcDevice.ConnectivityService != nil {
		for _, profile := range srcDevice.ConnectivityService.ConnectivityService {
			log.Infof("Migrating Connectivity Service %s", gnmiclient.StrDeref(profile.Id))
			actions, err := migrateV4V2ConnectivityService(fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, actions)
		}
	}

	//migrating Enterprise to Enterprise
	if srcDevice.Enterprise != nil {
		for _, profile := range srcDevice.Enterprise.Enterprise {
			log.Infof("Migrating Enterprise %s", gnmiclient.StrDeref(profile.Id))
			action, err := migrateV4V2Enterprise(fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, action)
		}
	}

	//migrating Application to Enterprise->Application
	if srcDevice.Application != nil {
		for _, profile := range srcDevice.Application.Application {
			log.Infof("Migrating Application Profile %s", gnmiclient.StrDeref(profile.Id))
			action, err := migrateV4V2Application(fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, action)
		}
	}

	//migrating TrafficClass to Enterprise->TrafficClass
	if srcDevice.TrafficClass != nil {
		for _, profile := range srcDevice.Enterprise.Enterprise {
			entID := profile.Id
			for _, tc := range srcDevice.TrafficClass.TrafficClass {
				log.Infof("Migrating Traffic Class %s", gnmiclient.StrDeref(tc.Id))
				action, err := migrateV4V2TrafficClass(fromTarget, toTarget, entID, tc)
				if err != nil {
					log.Warn(err.Error())
					continue
				}
				allActions = append(allActions, action)
			}
		}
	}

	//migrating Template to Enterprise->Template
	if srcDevice.Template != nil {
		for _, profile := range srcDevice.Enterprise.Enterprise {
			entID := profile.Id
			for _, tm := range srcDevice.Template.Template {
				log.Infof("Migrating Template Profile %s", gnmiclient.StrDeref(tm.Id))
				action, err := migrateV4V2Template(fromTarget, toTarget, entID, tm)
				if err != nil {
					log.Warn(err.Error())
					continue
				}
				allActions = append(allActions, action)
			}
		}
	}

	//migrating Site to Enterprise->Site
	if srcDevice.Site != nil {
		for _, profile := range srcDevice.Site.Site {
			log.Infof("Migrating Site Profile %s", gnmiclient.StrDeref(profile.Id))
			action, err := migrateV4V2Site(fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, action)
		}
	}

	//migrating Upf to Enterprise->Site->Upf
	if srcDevice.Upf != nil {
		for _, profile := range srcDevice.Upf.Upf {
			log.Infof("Migrating Upf Profile %s", gnmiclient.StrDeref(profile.Id))
			action, err := migrateV4V2Upf(fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, action)
		}
	}

	//migrating Vcs to Enterprise->Site->Slice
	if srcDevice.Vcs != nil {
		for _, profile := range srcDevice.Vcs.Vcs {
			log.Infof("Migrating Vcs Profile %s", gnmiclient.StrDeref(profile.Id))
			action, err := migrateV4V2Vcs(fromTarget, toTarget, profile)
			if err != nil {
				log.Warn(err.Error())
				continue
			}
			allActions = append(allActions, action)
		}
	}

	//migrating DeviceGroup to Enterprise->Site->DeviceGroup
	if srcDevice.DeviceGroup != nil {
		for _, profile := range srcDevice.Vcs.Vcs {
			entID := profile.Enterprise
			siteID := profile.Site

			for _, dg := range profile.DeviceGroup {
				dgID := dg.DeviceGroup
				dg := srcDevice.DeviceGroup.DeviceGroup[*dgID]
				log.Infof("Migrating Device Group %s", gnmiclient.StrDeref(dg.Id))
				action, err := migrateV4V2DeviceGroup(fromTarget, toTarget, entID, siteID, dg)
				if err != nil {
					log.Warn(err.Error())
					continue
				}
				allActions = append(allActions, action)
			}

		}
	}

	//migrating IpDomain to Enterprise->Site->IpDomain
	if srcDevice.IpDomain != nil {
		for _, profile := range srcDevice.Vcs.Vcs {
			entID := profile.Enterprise
			siteID := profile.Site
			for _, dg := range profile.DeviceGroup {
				dgID := dg.DeviceGroup
				ipdID := srcDevice.DeviceGroup.DeviceGroup[*dgID].IpDomain
				ipd := srcDevice.IpDomain.IpDomain[*ipdID]
				log.Infof("Migrating IpDomain Profile %s", gnmiclient.StrDeref(ipd.Id))
				action, err := migrateV4V2IpDomain(fromTarget, toTarget, entID, siteID, ipd)
				if err != nil {
					log.Warn(err.Error())
					continue
				}
				allActions = append(allActions, action)
			}
		}
	}

	//migrate DeviceGroup->imsis to Enterprise->Site->device and Enterprise->Site->sim
	if srcDevice.DeviceGroup != nil {
		for _, profile := range srcDevice.Vcs.Vcs {
			entID := profile.Enterprise
			siteID := profile.Site
			imsiDefination := srcDevice.Site.Site[*siteID].ImsiDefinition

			for _, dg := range profile.DeviceGroup {
				dgID := dg.DeviceGroup
				dg := srcDevice.DeviceGroup.DeviceGroup[*dgID]
				log.Infof("Migrating Device Group's imsis %s", gnmiclient.StrDeref(dg.Id))
				for _, im := range dg.Imsis {
					action, err := migrateV4V2DeviceGroupImsis(fromTarget, toTarget, entID, siteID, imsiDefination, im)
					if err != nil {
						log.Warn(err.Error())
						continue
					}
					allActions = append(allActions, action)
				}
			}

		}
	}

	return allActions, nil
}

func migrateV4V2ConnectivityService(fromTarget string, toTarget string, cs *modelsv4.OnfConnectivityService_ConnectivityService_ConnectivityService) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, cs.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, cs.DisplayName))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("core-5g-endpoint", toTarget, cs.Core_5GEndpoint))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("acc-prometheus-url", toTarget, cs.AccPrometheusUrl))

	prefix := gnmiclient.StringToPath(fmt.Sprintf("connectivity-services/connectivity-service[id=%s]", *cs.Id), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("connectivity-service/connectivity-service[id=%s]", *cs.Id), fromTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV4V2Enterprise(fromTarget string, toTarget string, ent *modelsv4.OnfEnterprise_Enterprise_Enterprise) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, ent.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, ent.DisplayName))

	if ent.ConnectivityService != nil {
		for _, cs := range ent.ConnectivityService {
			updStr := fmt.Sprintf("connectivity-service[connectivity-service=%s]/enabled", *cs.ConnectivityService)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateBool(updStr, toTarget, cs.Enabled))
		}
	}

	prefix := gnmiclient.StringToPath(fmt.Sprintf("enterprises/enterprise[ent-id=%s]", *ent.Id), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("enterprise/enterprise[id=%s]", *ent.Id), fromTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV4V2Application(fromTarget string, toTarget string, app *modelsv4.OnfApplication_Application_Application) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, app.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, app.DisplayName))

	if app.Endpoint != nil {
		for _, ap := range app.Endpoint {
			updStr := fmt.Sprintf("endpoint[endpoint-id=%s]/display-name", *ap.EndpointId)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString(updStr, toTarget, ap.DisplayName))
			updStr = fmt.Sprintf("endpoint[endpoint-id=%s]/protocol", *ap.EndpointId)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString(updStr, toTarget, ap.Protocol))
			updStr = fmt.Sprintf("endpoint[endpoint-id=%s]/port-start", *ap.EndpointId)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt16(updStr, toTarget, ap.PortStart))
			updStr = fmt.Sprintf("endpoint[endpoint-id=%s]/port-end", *ap.EndpointId)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt16(updStr, toTarget, ap.PortEnd))
			updStr = fmt.Sprintf("endpoint[endpoint-id=%s]/mbr/uplink", *ap.EndpointId)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64(updStr, toTarget, ap.Mbr.Uplink))
			updStr = fmt.Sprintf("endpoint[endpoint-id=%s]/mbr/downlink", *ap.EndpointId)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64(updStr, toTarget, ap.Mbr.Downlink))
			updStr = fmt.Sprintf("endpoint[endpoint-id=%s]/traffic-class", *ap.EndpointId)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString(updStr, toTarget, ap.TrafficClass))
		}
	}

	prefix := gnmiclient.StringToPath(fmt.Sprintf("enterprises/enterprise[id=%s]/application[app-id=%s]", *app.Enterprise, *app.Id), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("application/application[id=%s]", *app.Id), fromTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV4V2TrafficClass(fromTarget string, toTarget string, entID *string, tc *modelsv4.OnfTrafficClass_TrafficClass_TrafficClass) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, tc.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, tc.DisplayName))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt8("qci", toTarget, tc.Qci))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateInt8("pelr", toTarget, tc.Pelr))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt16("pdb", toTarget, tc.Pdb))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt8("arp", toTarget, tc.Arp))

	prefix := gnmiclient.StringToPath(fmt.Sprintf("enterprises/enterprise[id=%s]/traffic-class[tc-id=%s]", *entID, *tc.Id), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("traffic-class/traffic-class[id=%s]", *tc.Id), fromTarget)
	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV4V2Template(fromTarget string, toTarget string, entID *string, te *modelsv4.OnfTemplate_Template_Template) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, te.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, te.DisplayName))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt8("sst", toTarget, te.Sst))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt32("sd", toTarget, te.Sd))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("default-behavior", toTarget, te.DefaultBehavior))

	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64("slice/mbr/uplink", toTarget, te.Slice.Mbr.Uplink))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64("slice/mbr/downlink", toTarget, te.Slice.Mbr.Downlink))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt32("slice/mbr/uplink-burst-size", toTarget, te.Slice.Mbr.UplinkBurstSize))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt32("slice/mbr/downlink-burst-size", toTarget, te.Slice.Mbr.DownlinkBurstSize))

	prefix := gnmiclient.StringToPath(fmt.Sprintf("enterprises/enterprise[id=%s]/template[tp-id=%s]", *entID, *te.Id), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("template/template[id=%s]", *te.Id), fromTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV4V2Site(fromTarget string, toTarget string, st *modelsv4.OnfSite_Site_Site) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, st.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, st.DisplayName))

	if st.SmallCell != nil {
		for _, sc := range st.SmallCell {
			updStr := fmt.Sprintf("small-cell[small-cell-id=%s]/display-name", *sc.SmallCellId)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString(updStr, toTarget, sc.DisplayName))
			updStr = fmt.Sprintf("small-cell[small-cell-id=%s]/address", *sc.SmallCellId)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString(updStr, toTarget, sc.Address))
			updStr = fmt.Sprintf("small-cell[small-cell-id=%s]/tac", *sc.SmallCellId)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString(updStr, toTarget, sc.Tac))
			updStr = fmt.Sprintf("small-cell[small-cell-id=%s]/enable", *sc.SmallCellId)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateBool(updStr, toTarget, sc.Enable))
		}
	}

	if st.Monitoring != nil {
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("monitoring/edge-cluster-prometheus-url", toTarget, st.Monitoring.EdgeClusterPrometheusUrl))
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("monitoring/edge-monitoring-prometheus-url", toTarget, st.Monitoring.EdgeMonitoringPrometheusUrl))
		for _, ed := range st.Monitoring.EdgeDevice {
			updStr := fmt.Sprintf("monitoring/edge-device[edge-device-id=%s]/display-name", *ed.EdgeDeviceId)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString(updStr, toTarget, ed.DisplayName))
			updStr = fmt.Sprintf("monitoring/edge-device[edge-device-id=%s]/description", *ed.EdgeDeviceId)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString(updStr, toTarget, ed.Description))
		}
	}

	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("imsi-definition/mcc", toTarget, st.ImsiDefinition.Mcc))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("imsi-definition/mnc", toTarget, st.ImsiDefinition.Mnc))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("imsi-definition/format", toTarget, st.ImsiDefinition.Format))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt32("imsi-definition/enterprise", toTarget, st.ImsiDefinition.Enterprise))

	prefix := gnmiclient.StringToPath(fmt.Sprintf("enterprises/enterprise[ent-id=%s]/site[site-id=%s]", *st.Enterprise, *st.Id), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("site/site[id=%s]", *st.Id), fromTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV4V2Upf(fromTarget string, toTarget string, up *modelsv4.OnfUpf_Upf_Upf) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, up.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, up.DisplayName))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("address", toTarget, up.Address))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt16("port", toTarget, up.Port))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("config-endpoint", toTarget, up.ConfigEndpoint))

	prefix := gnmiclient.StringToPath(fmt.Sprintf("enterprises/enterprise[ent-id=%s]/site[site-id=%s]/upf[upf-id=%s]",
		*up.Enterprise, *up.Site, *up.Id), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("upf/upf[id=%s]", *up.Id), fromTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV4V2Vcs(fromTarget string, toTarget string, vc *modelsv4.OnfVcs_Vcs_Vcs) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, vc.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, vc.DisplayName))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("upf", toTarget, vc.Upf))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt8("sst", toTarget, vc.Sst))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt32("sd", toTarget, vc.Sd))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("default-behavior", toTarget, vc.DefaultBehavior))

	if vc.DeviceGroup != nil {
		for _, dg := range vc.DeviceGroup {
			updStr := fmt.Sprintf("device-group[device-group=%s]/enable", *dg.DeviceGroup)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateBool(updStr, toTarget, dg.Enable))
		}
	}

	if vc.Filter != nil {
		for _, fi := range vc.Filter {
			updStr := fmt.Sprintf("filter[application=%s]/priority", *fi.Application)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt8(updStr, toTarget, fi.Priority))
			updStr = fmt.Sprintf("filter[application=%s]/allow", *fi.Application)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateBool(updStr, toTarget, fi.Allow))
		}
	}

	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64("/mbr/uplink", toTarget, vc.Slice.Mbr.Uplink))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64("/mbr/downlink", toTarget, vc.Slice.Mbr.Downlink))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt32("/mbr/uplink-burst-size", toTarget, vc.Slice.Mbr.UplinkBurstSize))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt32("/mbr/downlink-burst-size", toTarget, vc.Slice.Mbr.DownlinkBurstSize))

	prefix := gnmiclient.StringToPath(fmt.Sprintf("enterprises/enterprise[ent-id=%s]/site[site-id=%s]/slice[slice-id=%s]",
		*vc.Enterprise, *vc.Site, *vc.Id), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("vcs/vcs[id=%s]", *vc.Id), fromTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV4V2DeviceGroup(fromTarget string, toTarget string, entID *string, siteID *string, dg *modelsv4.OnfDeviceGroup_DeviceGroup_DeviceGroup) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, dg.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, dg.DisplayName))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("site", toTarget, dg.Site))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("ip-domain", toTarget, dg.IpDomain))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64("mbr/uplink", toTarget, dg.Device.Mbr.Uplink))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64("mbr/downlink", toTarget, dg.Device.Mbr.Downlink))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("traffic-class", toTarget, dg.Device.TrafficClass))

	if dg.Imsis != nil {
		for _, im := range dg.Imsis {

			for dev := *im.ImsiRangeFrom; dev <= *im.ImsiRangeTo; dev++ {
				devID := fmt.Sprintf(*im.ImsiId+"-imsi-range-%d", dev)
				updStr := fmt.Sprintf("device[dev-id=%s]/enable", devID)
				enable := true
				updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateBool(updStr, toTarget, &enable))
			}

		}
	}

	prefix := gnmiclient.StringToPath(fmt.Sprintf("enterprises/enterprise[ent-id=%s]/site[site-id=%s]/device-group[dg-id=%s]",
		*entID, *siteID, *dg.Id), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("device-group/device-group[id=%s]", *dg.Id), fromTarget)
	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV4V2IpDomain(fromTarget string, toTarget string, entID *string, siteID *string, ipd *modelsv4.OnfIpDomain_IpDomain_IpDomain) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, ipd.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, ipd.DisplayName))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("dnn", toTarget, ipd.Dnn))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("dns-primary", toTarget, ipd.DnsPrimary))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("dns-secondary", toTarget, ipd.DnsSecondary))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt16("mtu", toTarget, ipd.Mtu))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("subnet", toTarget, ipd.Subnet))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("admin-status", toTarget, ipd.AdminStatus))

	prefix := gnmiclient.StringToPath(fmt.Sprintf("enterprises/enterprise[ent-id=%s]/site[site-id=%s]/ip-domain[ip-id=%s]",
		*entID, *siteID, *ipd.Id), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("ip-domain/ip-domain[id=%s]", *ipd.Id), fromTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV4V2DeviceGroupImsis(fromTarget string, toTarget string, entID *string, siteID *string, imDef *modelsv4.OnfSite_Site_Site_ImsiDefinition, im *modelsv4.OnfDeviceGroup_DeviceGroup_DeviceGroup_Imsis) (*migration.MigrationActions, error) {
	var updates []*gpb.Update

	for dev := *im.ImsiRangeFrom; dev <= *im.ImsiRangeTo; dev++ {
		devID := fmt.Sprintf(*im.ImsiId+"-%d", dev)
		displayName := fmt.Sprintf(*im.ImsiId+" %d", dev)
		simCardID := fmt.Sprintf("sim-"+*im.ImsiId+"-%d", dev)

		updStr := fmt.Sprintf("device[dev-id=%s]/display-name", devID)
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString(updStr, toTarget, &displayName))
		updStr = fmt.Sprintf("device[dev-id=%s]/sim-card", devID)
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString(updStr, toTarget, &simCardID))

		simDisplayName := fmt.Sprintf("Sim "+*im.ImsiId+" %d", dev)
		imsi, err := synchronizer.FormatImsi(*imDef.Format, *imDef.Mcc, *imDef.Mnc, *imDef.Enterprise, dev)
		if err != nil {
			log.Warn(err.Error())
		}

		updStr = fmt.Sprintf("sim-card[sim-id=%s]/display-name", simCardID)
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString(updStr, toTarget, &simDisplayName))
		updStr = fmt.Sprintf("sim-card[sim-id=%s]/imsi", simCardID)
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64(updStr, toTarget, &imsi))

	}

	prefix := gnmiclient.StringToPath(fmt.Sprintf("enterprises/enterprise[ent-id=%s]/site[site-id=%s]", *entID, *siteID), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("device-group/device-group[id=%s]/imsi[imsi-id]", *im.ImsiId), fromTarget)
	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}
