// SPDX-FileCopyrightText: 2022-present Intel Corporation
// SPDX-FileCopyrightText: 2022-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

/*
 * Implements the migration function from v2.0.0 aether models to v2.1.0 aether models. This
 * involves migrating each of the profiles (EnterpriseProfile, DeviceProfile, etc) and then
 * the UE.
 */

package steps

import (
	"fmt"
	modelsv2 "github.com/onosproject/aether-models/models/aether-2.0.x/v2/api"
	modelsv2_1 "github.com/onosproject/aether-models/models/aether-2.1.x/v2/api"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/gnmiclient"
	"github.com/onosproject/sdcore-adapter/pkg/migration"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"regexp"
	"strings"
)

var log = logging.GetLogger("migration.steps")

type enterpriseHierarchy struct {
	entID string
	site  []siteHierarchy
}

type siteHierarchy struct {
	siteID    string
	slicesIDs []string
}

// MigrateV2V21 - top level migration entry
func MigrateV2V21(step *migration.MigrationStep, fromTarget string, unused string, srcVal *gpb.TypedValue, destVal *gpb.TypedValue) ([]*migration.MigrationActions, error) {
	srcJSONBytes := srcVal.GetJsonVal()
	srcDevice := &modelsv2.Device{}

	if len(srcJSONBytes) > 0 {
		if err := step.FromModels.Unmarshal(srcJSONBytes, srcDevice); err != nil {
			return nil, err
		}
	}

	destJSONBytes := destVal.GetJsonVal()
	destDevice := &modelsv2_1.Device{}
	if len(destJSONBytes) > 0 {
		if err := step.ToModels.Unmarshal(destJSONBytes, destDevice); err != nil {
			return nil, err
		}
	}

	if len(srcDevice.Enterprises.Enterprise) == 0 {
		return nil, fmt.Errorf("no enterprises to migrate")
	}

	allActions := make([]*migration.MigrationActions, 0)

	entHierarchies := make([]enterpriseHierarchy, 0)
	for entID, e := range srcDevice.Enterprises.Enterprise {
		entHierarchy := enterpriseHierarchy{
			entID: entID,
			site:  []siteHierarchy{},
		}
		log.Infof("Please add an Entity to onos-topo for each enterprise %s", entID)

		//migrating Application to Enterprise->Application
		for _, app := range e.Application {
			log.Infof("Migrating Application Profile %s", gnmiclient.StrDeref(app.ApplicationId))
			action, err := migrateV2V21Application(fromTarget, entID, app)
			if err != nil {
				return nil, err
			}
			allActions = append(allActions, action)
		}

		for _, tc := range e.TrafficClass {
			log.Infof("Migrating Traffic Class %s", gnmiclient.StrDeref(tc.TrafficClassId))
			action, err := migrateV2V21TrafficClass(fromTarget, entID, tc)
			if err != nil {
				return nil, err
			}
			allActions = append(allActions, action)
		}

		for _, tm := range e.Template {
			log.Infof("Migrating Template Profile %s", gnmiclient.StrDeref(tm.TemplateId))
			action, err := migrateV2V21Template(fromTarget, entID, tm)
			if err != nil {
				return nil, err
			}
			allActions = append(allActions, action)
		}

		for siteID, site := range e.Site {
			sites := siteHierarchy{
				siteID:    siteID,
				slicesIDs: []string{},
			}

			log.Infof("Migrating Site %s", siteID)
			action, err := migrateV2V21Site(fromTarget, entID, site)
			if err != nil {
				return nil, err
			}
			allActions = append(allActions, action)

			for upfID, upf := range site.Upf {
				log.Infof("Migrating Upf Profile %s", upfID)
				action, err := migrateV2V21Upf(fromTarget, entID, siteID, upf)
				if err != nil {
					return nil, err
				}
				allActions = append(allActions, action)
			}

			for sliceID, slice := range site.Slice {
				log.Infof("Migrating Slice %s", sliceID)
				sites.slicesIDs = append(sites.slicesIDs, sliceID)
				action, err := migrateV2V21Slice(fromTarget, entID, siteID, slice)
				if err != nil {
					return nil, err
				}
				allActions = append(allActions, action)
			}

			for dgID, dg := range site.DeviceGroup {
				log.Infof("Migrating Device Group %s", dgID)
				action, err := migrateV2V21DeviceGroup(fromTarget, entID, siteID, dg)
				if err != nil {
					return nil, err
				}
				allActions = append(allActions, action)
			}

			for ipID, ipd := range site.IpDomain {
				log.Infof("Migrating Device Group %s", ipID)
				action, err := migrateV2V21IpDomain(fromTarget, entID, siteID, ipd)
				if err != nil {
					return nil, err
				}
				allActions = append(allActions, action)
			}

			for scID, sc := range site.SimCard {
				log.Infof("Migrating Sim Card %s", scID)
				action, err := migrateV2V21SimCard(fromTarget, entID, siteID, sc)
				if err != nil {
					return nil, err
				}
				allActions = append(allActions, action)
			}

			for devID, dev := range site.Device {
				log.Infof("Migrating Device %s", devID)
				action, err := migrateV2V21Device(fromTarget, entID, siteID, dev)
				if err != nil {
					return nil, err
				}
				allActions = append(allActions, action)
			}
			entHierarchy.site = append(entHierarchy.site, sites)
		}
		entHierarchies = append(entHierarchies, entHierarchy)
	}

	//migrating Connectivity Service to Connectivity Service
	if srcDevice.ConnectivityServices != nil {
		for _, profile := range srcDevice.ConnectivityServices.ConnectivityService {
			log.Infof("Migrating Connectivity Service %s", gnmiclient.StrDeref(profile.ConnectivityServiceId))
			actions, err := migrateV2V21ConnectivityService(entHierarchies, profile)
			if err != nil {
				return nil, err
			}
			allActions = append(allActions, actions)
		}
	}

	return allActions, nil
}

func migrateV2V21ConnectivityService(siteHierarchies []enterpriseHierarchy, cs *modelsv2.OnfConnectivityService_ConnectivityServices_ConnectivityService) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	for _, ent := range siteHierarchies {
		for _, site := range ent.site {
			sliceType := "5g"
			if cs.Core_5GEndpoint != nil &&
				(strings.Contains(strings.ToLower(*cs.Core_5GEndpoint), "4g") || strings.Contains(strings.ToLower(*cs.DisplayName), "4g")) {
				updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString(
					fmt.Sprintf("site[site-id=%s]/connectivity-service/core-4g/endpoint", site.siteID), ent.entID, cs.Core_5GEndpoint))
				updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString(
					fmt.Sprintf("site[site-id=%s]/connectivity-service/core-4g/acc-prometheus-url", site.siteID), ent.entID, cs.AccPrometheusUrl))
				sliceType = "4g"
			} else {
				updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString(
					fmt.Sprintf("site[site-id=%s]/connectivity-service/core-5g/endpoint", site.siteID), ent.entID, cs.Core_5GEndpoint))
				updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString(
					fmt.Sprintf("site[site-id=%s]/connectivity-service/core-5g/acc-prometheus-url", site.siteID), ent.entID, cs.AccPrometheusUrl))
			}
			for _, slice := range site.slicesIDs {
				updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString(
					fmt.Sprintf("site[site-id=%s]/slice[slice-id=%s]/connectivity-service", site.siteID, slice), ent.entID, &sliceType))
			}
		}
	}

	return &migration.MigrationActions{UpdatePrefix: nil, Updates: updates, Deletes: []*gpb.Path{}}, nil
}

func migrateV2V21Application(fromTarget string, toTarget string, app *modelsv2.OnfEnterprise_Enterprises_Enterprise_Application) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, app.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, app.DisplayName))

	matched, err := regexp.Match(`^\d*\.\d*\.\d*\.\d*/\d*`, []byte(*app.Address))
	if err != nil {
		return nil, err
	}
	if !matched {
		log.Warnf("application %s address is not an IP address '%s' - only IP address is allowed in 2.1.x",
			*app.ApplicationId, *app.Address)
		fakeAddress := "0.0.0.0/32"
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("address", toTarget, &fakeAddress))
	} else {
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("address", toTarget, app.Address))
	}

	if app.Endpoint != nil {
		for _, ap := range app.Endpoint {
			epID := *ap.EndpointId
			if !isValidIdentifier(epID) {
				epID = convertIdentifier(epID)
			}
			updStr := fmt.Sprintf("endpoint[endpoint-id=%s]/display-name", epID)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString(updStr, toTarget, ap.DisplayName))
			updStr = fmt.Sprintf("endpoint[endpoint-id=%s]/protocol", epID)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString(updStr, toTarget, ap.Protocol))
			updStr = fmt.Sprintf("endpoint[endpoint-id=%s]/port-start", epID)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt16(updStr, toTarget, ap.PortStart))
			updStr = fmt.Sprintf("endpoint[endpoint-id=%s]/port-end", epID)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt16(updStr, toTarget, ap.PortEnd))
			updStr = fmt.Sprintf("endpoint[endpoint-id=%s]/mbr/uplink", epID)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64(updStr, toTarget, ap.Mbr.Uplink))
			updStr = fmt.Sprintf("endpoint[endpoint-id=%s]/mbr/downlink", epID)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64(updStr, toTarget, ap.Mbr.Downlink))
			updStr = fmt.Sprintf("endpoint[endpoint-id=%s]/traffic-class", epID)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString(updStr, toTarget, ap.TrafficClass))
		}
	}

	appID := *app.ApplicationId
	if !isValidIdentifier(appID) {
		appID = convertIdentifier(appID)
	}

	prefix := gnmiclient.StringToPath(fmt.Sprintf("application[application-id=%s]", appID), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("enterprises/enterprise[enterprise-id=%s]/application/application[id=%s]",
		fromTarget, *app.ApplicationId), fromTarget)
	deletePath.Origin = "enterprise,address"

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV2V21TrafficClass(fromTarget string, toTarget string, tc *modelsv2.OnfEnterprise_Enterprises_Enterprise_TrafficClass) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, tc.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, tc.DisplayName))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt8("qci", toTarget, tc.Qci))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateInt8("pelr", toTarget, tc.Pelr))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt16("pdb", toTarget, tc.Pdb))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt8("arp", toTarget, tc.Arp))

	tcID := *tc.TrafficClassId
	if !isValidIdentifier(tcID) {
		tcID = convertIdentifier(tcID)
	}

	prefix := gnmiclient.StringToPath(fmt.Sprintf("traffic-class[traffic-class-id=%s]", tcID), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("enterprises/enterprise[enterprise-id=%s]/traffic-class[id=%s]",
		fromTarget, *tc.TrafficClassId), fromTarget)
	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV2V21Template(fromTarget string, toTarget string, te *modelsv2.OnfEnterprise_Enterprises_Enterprise_Template) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, te.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, te.DisplayName))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt8("sst", toTarget, te.Sst))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt32("sd", toTarget, te.Sd))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("default-behavior", toTarget, te.DefaultBehavior))

	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64("mbr/uplink", toTarget, te.Mbr.Uplink))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64("mbr/downlink", toTarget, te.Mbr.Downlink))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt32("mbr/uplink-burst-size", toTarget, te.Mbr.UplinkBurstSize))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt32("mbr/downlink-burst-size", toTarget, te.Mbr.DownlinkBurstSize))

	teID := *te.TemplateId
	if !isValidIdentifier(teID) {
		teID = convertIdentifier(teID)
	}

	prefix := gnmiclient.StringToPath(fmt.Sprintf("template[template-id=%s]", teID), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("enterprises/enterprise[enterprise-id=%s]/template/template[id=%s]",
		fromTarget, *te.TemplateId), fromTarget)
	deletePath.Origin = "default-behavior"

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV2V21Site(fromTarget string, toTarget string, st *modelsv2.OnfEnterprise_Enterprises_Enterprise_Site) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, st.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, st.DisplayName))

	if st.SmallCell != nil {
		for _, sc := range st.SmallCell {
			smcID := *sc.SmallCellId
			if !isValidIdentifier(smcID) {
				smcID = convertIdentifier(smcID)
			}
			updStr := fmt.Sprintf("small-cell[small-cell-id=%s]/display-name", smcID)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString(updStr, toTarget, sc.DisplayName))
			updStr = fmt.Sprintf("small-cell[small-cell-id=%s]/address", smcID)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString(updStr, toTarget, sc.Address))
			updStr = fmt.Sprintf("small-cell[small-cell-id=%s]/tac", smcID)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString(updStr, toTarget, sc.Tac))
			updStr = fmt.Sprintf("small-cell[small-cell-id=%s]/enable", smcID)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateBool(updStr, toTarget, sc.Enable))
		}
	}

	if st.Monitoring != nil {
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("monitoring/edge-cluster-prometheus-url", toTarget, st.Monitoring.EdgeClusterPrometheusUrl))
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("monitoring/edge-monitoring-prometheus-url", toTarget, st.Monitoring.EdgeMonitoringPrometheusUrl))
		for _, ed := range st.Monitoring.EdgeDevice {
			edID := *ed.EdgeDeviceId
			if !isValidIdentifier(edID) {
				edID = convertIdentifier(edID)
			}
			updStr := fmt.Sprintf("monitoring/edge-device[edge-device-id=%s]/display-name", edID)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString(updStr, toTarget, ed.DisplayName))
			updStr = fmt.Sprintf("monitoring/edge-device[edge-device-id=%s]/description", edID)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString(updStr, toTarget, ed.Description))
		}
	}

	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("imsi-definition/mcc", toTarget, st.ImsiDefinition.Mcc))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("imsi-definition/mnc", toTarget, st.ImsiDefinition.Mnc))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("imsi-definition/format", toTarget, st.ImsiDefinition.Format))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt32("imsi-definition/enterprise", toTarget, st.ImsiDefinition.Enterprise))

	stID := *st.SiteId
	if !isValidIdentifier(stID) {
		stID = convertIdentifier(stID)
	}

	prefix := gnmiclient.StringToPath(fmt.Sprintf("site[site-id=%s]", stID), toTarget)
	deletePath := gnmiclient.StringToPath(fmt.Sprintf("enterprises/enterprise[enterprise-id=%s]/site/site[id=%s]",
		fromTarget, *st.SiteId), fromTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}, nil
}

func migrateV2V21Upf(fromTarget string, toTarget string, siteID string, up *modelsv2.OnfEnterprise_Enterprises_Enterprise_Site_Upf) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, up.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, up.DisplayName))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("address", toTarget, up.Address))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt16("port", toTarget, up.Port))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("config-endpoint", toTarget, up.ConfigEndpoint))

	upfID := *up.UpfId
	if !isValidIdentifier(upfID) {
		upfID = convertIdentifier(upfID)
	}

	prefix := gnmiclient.StringToPath(fmt.Sprintf("site[site-id=%s]/upf[upf-id=%s]",
		siteID, upfID), toTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{}}, nil
}

func migrateV2V21Slice(fromTarget string, toTarget string, siteID string, slice *modelsv2.OnfEnterprise_Enterprises_Enterprise_Site_Slice) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, slice.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, slice.DisplayName))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("upf", toTarget, slice.Upf))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt8("sst", toTarget, slice.Sst))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt32("sd", toTarget, slice.Sd))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("default-behavior", toTarget, slice.DefaultBehavior))

	if slice.DeviceGroup != nil {
		for _, dg := range slice.DeviceGroup {
			dgID := *dg.DeviceGroup
			if !isValidIdentifier(dgID) {
				dgID = convertIdentifier(dgID)
			}
			updStr := fmt.Sprintf("device-group[device-group=%s]/enable", dgID)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateBool(updStr, toTarget, dg.Enable))
		}
	}

	if slice.Filter != nil {
		for _, fi := range slice.Filter {
			appID := *fi.Application
			if !isValidIdentifier(appID) {
				appID = convertIdentifier(appID)
			}
			updStr := fmt.Sprintf("filter[application=%s]/priority", appID)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt8(updStr, toTarget, fi.Priority))
			updStr = fmt.Sprintf("filter[application=%s]/allow", appID)
			updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateBool(updStr, toTarget, fi.Allow))
		}
	}

	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64("/mbr/uplink", toTarget, slice.Mbr.Uplink))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64("/mbr/downlink", toTarget, slice.Mbr.Downlink))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt32("/mbr/uplink-burst-size", toTarget, slice.Mbr.UplinkBurstSize))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt32("/mbr/downlink-burst-size", toTarget, slice.Mbr.DownlinkBurstSize))

	slID := *slice.SliceId
	if !isValidIdentifier(slID) {
		slID = convertIdentifier(slID)
	}

	prefix := gnmiclient.StringToPath(fmt.Sprintf("site[site-id=%s]/slice[slice-id=%s]",
		siteID, slID), toTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{}}, nil
}

func migrateV2V21DeviceGroup(fromTarget string, toTarget string, siteID string, dg *modelsv2.OnfEnterprise_Enterprises_Enterprise_Site_DeviceGroup) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, dg.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, dg.DisplayName))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("ip-domain", toTarget, dg.IpDomain))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("traffic-class", toTarget, dg.TrafficClass))
	if dg.Mbr != nil {
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64("mbr/uplink", toTarget, dg.Mbr.Uplink))
		updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64("mbr/downlink", toTarget, dg.Mbr.Downlink))
	}

	stID := siteID
	if !isValidIdentifier(stID) {
		stID = convertIdentifier(stID)
	}

	dgID := *dg.DeviceGroupId
	if !isValidIdentifier(dgID) {
		dgID = convertIdentifier(dgID)
	}

	prefix := gnmiclient.StringToPath(fmt.Sprintf("site[site-id=%s]/device-group[device-group-id=%s]",
		stID, dgID), toTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{}}, nil
}

func migrateV2V21IpDomain(fromTarget string, toTarget string, siteID string, ipd *modelsv2.OnfEnterprise_Enterprises_Enterprise_Site_IpDomain) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, ipd.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, ipd.DisplayName))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("dnn", toTarget, ipd.Dnn))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("dns-primary", toTarget, ipd.DnsPrimary))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("dns-secondary", toTarget, ipd.DnsSecondary))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt16("mtu", toTarget, ipd.Mtu))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("subnet", toTarget, ipd.Subnet))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("admin-status", toTarget, ipd.AdminStatus))

	stID := siteID
	if !isValidIdentifier(stID) {
		stID = convertIdentifier(stID)
	}

	ipdID := *ipd.IpDomainId
	if !isValidIdentifier(ipdID) {
		ipdID = convertIdentifier(ipdID)
	}

	prefix := gnmiclient.StringToPath(fmt.Sprintf("site[site-id=%s]/ip-domain[ip-domain-id=%s]",
		stID, ipdID), toTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{}}, nil
}

func migrateV2V21Device(fromTarget string, toTarget string, siteID string, dev *modelsv2.OnfEnterprise_Enterprises_Enterprise_Site_Device) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, dev.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, dev.DisplayName))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("imei", toTarget, dev.Imei))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("sim-card", toTarget, dev.SimCard))

	stID := siteID
	if !isValidIdentifier(stID) {
		stID = convertIdentifier(stID)
	}

	ipdID := *dev.DeviceId
	if !isValidIdentifier(ipdID) {
		ipdID = convertIdentifier(ipdID)
	}

	prefix := gnmiclient.StringToPath(fmt.Sprintf("site[site-id=%s]/device[device-id=%s]",
		stID, ipdID), toTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{}}, nil
}

func migrateV2V21SimCard(fromTarget string, toTarget string, siteID string, sim *modelsv2.OnfEnterprise_Enterprises_Enterprise_Site_SimCard) (*migration.MigrationActions, error) {
	var updates []*gpb.Update
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("description", toTarget, sim.Description))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("display-name", toTarget, sim.DisplayName))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateString("iccid", toTarget, sim.Iccid))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64("imsi", toTarget, sim.Imsi))

	stID := siteID
	if !isValidIdentifier(stID) {
		stID = convertIdentifier(stID)
	}

	ipdID := *sim.SimId
	if !isValidIdentifier(ipdID) {
		ipdID = convertIdentifier(ipdID)
	}

	prefix := gnmiclient.StringToPath(fmt.Sprintf("site[site-id=%s]/sim-card[sim-id=%s]",
		stID, ipdID), toTarget)

	return &migration.MigrationActions{UpdatePrefix: prefix, Updates: updates, Deletes: []*gpb.Path{}}, nil
}
