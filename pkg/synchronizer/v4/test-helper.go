// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package synchronizerv4

// helpers that are useful to multiple test cases

import (
	models "github.com/onosproject/config-models/modelplugin/aether-4.0.0/aether_4_0_0"
)

// TODO: Refactor to use mockgen and/or sample data files ?

// MakeEnterprise populates an Enterprise structure for unit tests
func MakeEnterprise(desc string, displayName string, id string, cs []string) *models.OnfEnterprise_Enterprise_Enterprise {
	csList := map[string]*models.OnfEnterprise_Enterprise_Enterprise_ConnectivityService{}

	for _, csID := range cs {
		csList[csID] = &models.OnfEnterprise_Enterprise_Enterprise_ConnectivityService{
			ConnectivityService: aStr(csID),
			Enabled:             aBool(true),
		}
	}

	ent := models.OnfEnterprise_Enterprise_Enterprise{
		Description:         aStr(desc),
		DisplayName:         aStr(displayName),
		Id:                  aStr(id),
		ConnectivityService: csList,
	}

	return &ent
}

// MakeCs makes a connectivity service structure for unit tests
func MakeCs(desc string, displayName string, id string) *models.OnfConnectivityService_ConnectivityService_ConnectivityService {
	cs := models.OnfConnectivityService_ConnectivityService_ConnectivityService{
		Description:     aStr(desc),
		DisplayName:     aStr(displayName),
		Id:              aStr(id),
		Core_5GEndpoint: aStr("http://5gcore"),
	}

	return &cs
}

// BuildSampleDeviceGroup builds a sample device group for unit testing
func BuildSampleDeviceGroup() (
	*models.OnfEnterprise_Enterprise_Enterprise,
	*models.OnfConnectivityService_ConnectivityService_ConnectivityService,
	map[string]*models.OnfTrafficClass_TrafficClass_TrafficClass,
	*models.OnfIpDomain_IpDomain_IpDomain,
	*models.OnfSite_Site_Site,
	*models.OnfDeviceGroup_DeviceGroup_DeviceGroup) {
	ent := MakeEnterprise("sample-ent-desc", "sample-ent-dn", "sample-ent", []string{"sample-cs"})
	cs := MakeCs("sample-cs-desc", "sample-cs-dn", "sample-cs")

	tc := &models.OnfTrafficClass_TrafficClass_TrafficClass{
		Id:          aStr("sample-traffic-class"),
		Description: aStr("sample-traffic-class-desc"),
		DisplayName: aStr("sample-traffic-class-dn"),
		Qci:         aUint8(55),
		Arp:         aUint8(3),
	}

	ipd := &models.OnfIpDomain_IpDomain_IpDomain{
		Description: aStr("sample-ipd-desc"),
		DisplayName: aStr("sample-ipd-dn"),
		Id:          aStr("sample-ipd"),
		Subnet:      aStr("1.2.3.4/24"),
		DnsPrimary:  aStr("8.8.8.8"),
		Mtu:         aUint16(1492),
		Dnn:         aStr("5ginternet"),
	}
	imsiDef := &models.OnfSite_Site_Site_ImsiDefinition{
		Mcc:        aStr("123"),
		Mnc:        aStr("456"),
		Enterprise: aUint32(789),
		Format:     aStr("CCCNNNEEESSSSSS"),
	}
	sc := &models.OnfSite_Site_Site_SmallCell{
		Name:    aStr("myradio"),
		Address: aStr("6.7.8.9"),
		Enable:  aBool(true),
		Tac:     aStr("77AB"),
	}
	site := &models.OnfSite_Site_Site{
		Description:    aStr("sample-site-desc"),
		DisplayName:    aStr("sample-site-dn"),
		Id:             aStr("sample-site"),
		Enterprise:     aStr("sample-ent"),
		ImsiDefinition: imsiDef,
		SmallCell:      map[string]*models.OnfSite_Site_Site_SmallCell{"myradio": sc},
	}
	imsi := models.OnfDeviceGroup_DeviceGroup_DeviceGroup_Imsis{
		ImsiRangeFrom: aUint64(1),
	}
	dgDevMbr := &models.OnfDeviceGroup_DeviceGroup_DeviceGroup_Device_Mbr{
		Downlink: aUint64(4321),
		Uplink:   aUint64(8765),
	}
	dg := &models.OnfDeviceGroup_DeviceGroup_DeviceGroup{
		//Description: aStr("sample-dg-desc"),
		DisplayName: aStr("sample-dg-dn"),
		Id:          aStr("sample-dg"),
		Site:        aStr("sample-site"),
		IpDomain:    aStr("sample-ipd"),
		Imsis:       map[string]*models.OnfDeviceGroup_DeviceGroup_DeviceGroup_Imsis{"sample-imsi": &imsi},
		Device:      &models.OnfDeviceGroup_DeviceGroup_DeviceGroup_Device{Mbr: dgDevMbr, TrafficClass: tc.Id},
	}

	tcList := map[string]*models.OnfTrafficClass_TrafficClass_TrafficClass{*tc.Id: tc}

	return ent, cs, tcList, ipd, site, dg
}

// BuildSampleVcs builds a sample vcs for testing
func BuildSampleVcs() (
	map[string]*models.OnfApplication_Application_Application,
	*models.OnfTemplate_Template_Template,
	*models.OnfUpf_Upf_Upf,
	*models.OnfVcs_Vcs_Vcs) {

	ep := &models.OnfApplication_Application_Application_Endpoint{
		Name:      aStr("sample-app-ep"),
		PortStart: aUint16(123),
		PortEnd:   aUint16(124),
		Protocol:  aStr("UDP"),
	}

	app1 := &models.OnfApplication_Application_Application{
		Id:          aStr("sample-app"),
		Description: aStr("sample-app-desc"),
		DisplayName: aStr("sample-app-dn"),
		Address:     aStr("1.2.3.4"),
		Endpoint:    map[string]*models.OnfApplication_Application_Application_Endpoint{"sample-app-ep": ep},
		Enterprise:  aStr("sample-ent"),
	}

	mbr := &models.OnfApplication_Application_Application_Endpoint_Mbr{
		Uplink:   aUint64(11223344),
		Downlink: aUint64(55667788),
	}

	ep2 := &models.OnfApplication_Application_Application_Endpoint{
		Name:         aStr("sample-app-ep"),
		PortStart:    aUint16(123),
		PortEnd:      aUint16(124),
		Protocol:     aStr("UDP"),
		Mbr:          mbr,
		TrafficClass: aStr("sample-traffic-class"),
	}

	app2 := &models.OnfApplication_Application_Application{
		Id:          aStr("sample-app2"),
		Description: aStr("sample-app2-desc"),
		DisplayName: aStr("sample-app2-dn"),
		Address:     aStr("1.2.3.5"),
		Endpoint:    map[string]*models.OnfApplication_Application_Application_Endpoint{"sample-app2-ep": ep2},
		Enterprise:  aStr("sample-ent"),
	}

	appLink := &models.OnfVcs_Vcs_Vcs_Filter{
		Allow:       aBool(true),
		Priority:    aUint8(7),
		Application: aStr("sample-app"),
	}

	app2Link := &models.OnfVcs_Vcs_Vcs_Filter{
		Allow:       aBool(false),
		Priority:    aUint8(8),
		Application: aStr("sample-app2"),
	}

	dgLink := &models.OnfVcs_Vcs_Vcs_DeviceGroup{
		DeviceGroup: aStr("sample-dg"),
		Enable:      aBool(true),
	}

	tp := &models.OnfTemplate_Template_Template{
		Id:          aStr("sample-template"),
		Description: aStr("sample-template-desc"),
		DisplayName: aStr("sample-template-dn"),
		Sd:          aUint32(111),
		Sst:         aUint8(222),
	}

	upf := &models.OnfUpf_Upf_Upf{
		Id:             aStr("sample-upf"),
		Address:        aStr("2.3.4.5"),
		ConfigEndpoint: aStr("http://upf"),
		Description:    aStr("sample-upf-desc"),
		DisplayName:    aStr("sample-upf-dn"),
		Port:           aUint16(66),
	}

	sliceQosMbr := &models.OnfVcs_Vcs_Vcs_Slice_Mbr{Uplink: aUint64(333), Downlink: aUint64(444)}
	sliceQos := &models.OnfVcs_Vcs_Vcs_Slice{Mbr: sliceQosMbr}

	vcs := &models.OnfVcs_Vcs_Vcs{
		Filter:          map[string]*models.OnfVcs_Vcs_Vcs_Filter{"sample-app": appLink, "sample-app2": app2Link},
		Description:     aStr("sample-vcs-desc"),
		DeviceGroup:     map[string]*models.OnfVcs_Vcs_Vcs_DeviceGroup{"sample-dg": dgLink},
		DisplayName:     aStr("sample-app-dn"),
		Id:              aStr("sample-vcs"),
		Sd:              aUint32(111),
		Sst:             aUint8(222),
		Slice:           sliceQos,
		Upf:             aStr("sample-upf"),
		DefaultBehavior: aStr("DENY-ALL"),
	}

	apps := map[string]*models.OnfApplication_Application_Application{*app1.Id: app1, *app2.Id: app2}

	return apps, tp, upf, vcs
}
