// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package synchronizer

// helpers that are useful to multiple test cases

import (
	"fmt"
	models "github.com/onosproject/aether-models/models/aether-2.0.x/api"
)

// TODO: Refactor to use mockgen and/or sample data files ?

// MakeEnterprise populates an Enterprise structure for unit tests
func MakeEnterprise(desc string, displayName string, id string, cs []string) *Enterprise {
	csList := map[string]*EnterpriseConnectivityService{}

	for _, csID := range cs {
		csList[csID] = &EnterpriseConnectivityService{
			ConnectivityService: aStr(csID),
			Enabled:             aBool(true),
		}
	}

	ent := Enterprise{
		Description:         aStr(desc),
		DisplayName:         aStr(displayName),
		EnterpriseId:        aStr(id),
		ConnectivityService: csList,
	}

	return &ent
}

// MakeCs makes a connectivity service structure for unit tests
func MakeCs(desc string, displayName string, id string) *ConnectivityService {
	cs := ConnectivityService{
		Description:           aStr(desc),
		DisplayName:           aStr(displayName),
		ConnectivityServiceId: aStr(id),
		Core_5GEndpoint:       aStr("http://5gcore"),
	}

	return &cs
}

// BuildSampleDeviceGroup builds a sample device group for unit testing
func BuildSampleDeviceGroup() (
	*Enterprise,
	*ConnectivityService,
	map[string]*TrafficClass,
	*IpDomain,
	*Site,
	*DeviceGroup) {
	ent := MakeEnterprise("sample-ent-desc", "sample-ent-dn", "sample-ent", []string{"sample-cs"})
	cs := MakeCs("sample-cs-desc", "sample-cs-dn", "sample-cs")

	tc := &TrafficClass{
		TrafficClassId: aStr("sample-traffic-class"),
		Description:    aStr("sample-traffic-class-desc"),
		DisplayName:    aStr("sample-traffic-class-dn"),
		Qci:            aUint8(55),
		Arp:            aUint8(3),
	}

	ipd := &IpDomain{
		Description: aStr("sample-ipd-desc"),
		DisplayName: aStr("sample-ipd-dn"),
		IpDomainId:  aStr("sample-ipd"),
		Subnet:      aStr("1.2.3.4/24"),
		DnsPrimary:  aStr("8.8.8.8"),
		Mtu:         aUint16(1492),
		Dnn:         aStr("5ginternet"),
	}
	imsiDef := &ImsiDefinition{
		Mcc:        aStr("123"),
		Mnc:        aStr("456"),
		Enterprise: aUint32(789),
		Format:     aStr("CCCNNNEEESSSSSS"),
	}
	sc := &SmallCell{
		SmallCellId: aStr("myradio"),
		Address:     aStr("6.7.8.9"),
		Enable:      aBool(true),
		Tac:         aStr("77AB"),
	}
	simCard := &SimCard{
		SimId: aStr("sample-sim"),
		Imsi:  aUint64(1),
	}
	device := &Device{
		DeviceId: aStr("sample-device"),
		SimCard:  aStr("sample-sim"),
	}
	dgDevice := &DeviceGroupDevice{
		DeviceId: aStr("sample-device"),
		// SimCard: ...
	}
	dgDevMbr := &DeviceGroupMbr{
		Downlink: aUint64(4321),
		Uplink:   aUint64(8765),
	}
	dg := &DeviceGroup{
		//Description: aStr("sample-dg-desc"),
		DisplayName:   aStr("sample-dg-dn"),
		DeviceGroupId: aStr("sample-dg"),
		IpDomain:      aStr("sample-ipd"),
		Device:        map[string]*DeviceGroupDevice{*dgDevice.DeviceId: dgDevice},
		Mbr:           dgDevMbr,
		TrafficClass:  tc.TrafficClassId,
	}
	site := &Site{
		Description:    aStr("sample-site-desc"),
		DisplayName:    aStr("sample-site-dn"),
		SiteId:         aStr("sample-site"),
		ImsiDefinition: imsiDef,
		Device:         map[string]*Device{*device.DeviceId: device},
		SimCard:        map[string]*SimCard{*simCard.SimId: simCard},
		SmallCell:      map[string]*SmallCell{*sc.SmallCellId: sc},
		DeviceGroup:    map[string]*DeviceGroup{*dg.DeviceGroupId: dg},
		IpDomain:       map[string]*IpDomain{*ipd.IpDomainId: ipd},
	}

	tcList := map[string]*TrafficClass{*tc.TrafficClassId: tc}

	ent.TrafficClass = tcList
	ent.Site = map[string]*Site{*site.SiteId: site}

	return ent, cs, tcList, ipd, site, dg
}

// BuildSampleSlice builds a sample slice and application for testing. The
// Slice and its related artifacts are attached to the Enterprise and Site
// that are passed as arguments.
func BuildSampleSlice(ent *Enterprise, site *Site) (
	map[string]*Application,
	*Template,
	*Upf,
	*Slice) {

	ep := &ApplicationEndpoint{
		EndpointId: aStr("sample-app-ep"),
		PortStart:  aUint16(123),
		PortEnd:    aUint16(124),
		Protocol:   aStr("UDP"),
	}

	app1 := &Application{
		ApplicationId: aStr("sample-app"),
		Description:   aStr("sample-app-desc"),
		DisplayName:   aStr("sample-app-dn"),
		Address:       aStr("1.2.3.4"),
		Endpoint:      map[string]*ApplicationEndpoint{"sample-app-ep": ep},
	}

	mbr := &ApplicationEndpointMbr{
		Uplink:   aUint64(11223344),
		Downlink: aUint64(55667788),
	}

	ep2 := &ApplicationEndpoint{
		EndpointId:   aStr("sample-app-ep"),
		PortStart:    aUint16(123),
		PortEnd:      aUint16(124),
		Protocol:     aStr("UDP"),
		Mbr:          mbr,
		TrafficClass: aStr("sample-traffic-class"),
	}

	app2 := &Application{
		ApplicationId: aStr("sample-app2"),
		Description:   aStr("sample-app2-desc"),
		DisplayName:   aStr("sample-app2-dn"),
		Address:       aStr("1.2.3.5"),
		Endpoint:      map[string]*ApplicationEndpoint{"sample-app2-ep": ep2},
	}

	appLink := &SliceFilter{
		Allow:       aBool(true),
		Priority:    aUint8(7),
		Application: aStr("sample-app"),
	}

	app2Link := &SliceFilter{
		Allow:       aBool(false),
		Priority:    aUint8(8),
		Application: aStr("sample-app2"),
	}

	dgLink := &SliceDeviceGroup{
		DeviceGroup: aStr("sample-dg"),
		Enable:      aBool(true),
	}

	tp := &Template{
		TemplateId:  aStr("sample-template"),
		Description: aStr("sample-template-desc"),
		DisplayName: aStr("sample-template-dn"),
		Sd:          aUint32(111),
		Sst:         aUint8(222),
	}

	upf := &Upf{
		UpfId:          aStr("sample-upf"),
		Address:        aStr("2.3.4.5"),
		ConfigEndpoint: aStr("http://upf"),
		Description:    aStr("sample-upf-desc"),
		DisplayName:    aStr("sample-upf-dn"),
		Port:           aUint16(66),
	}

	sliceMbr := &SliceMbr{Uplink: aUint64(333), Downlink: aUint64(444)}

	slice := &Slice{
		Filter:          map[string]*SliceFilter{"sample-app": appLink, "sample-app2": app2Link},
		Description:     aStr("sample-slice-desc"),
		DeviceGroup:     map[string]*SliceDeviceGroup{"sample-dg": dgLink},
		DisplayName:     aStr("sample-app-dn"),
		SliceId:         aStr("sample-slice"),
		Sd:              aUint32(111),
		Sst:             aUint8(222),
		Mbr:             sliceMbr,
		Upf:             aStr("sample-upf"),
		DefaultBehavior: aStr("DENY-ALL"),
	}

	apps := map[string]*Application{*app1.ApplicationId: app1, *app2.ApplicationId: app2}

	ent.Application = apps
	site.Slice = map[string]*Slice{*slice.SliceId: slice}
	site.Upf = map[string]*Upf{*upf.UpfId: upf}

	return apps, tp, upf, slice
}

// BuildSampleDevice builds a sample device, with VCS and Device-Group
func BuildSampleDevice() *RootDevice {
	ent, cs, _, _, site, _ := BuildSampleDeviceGroup() // nolint dogsled
	_, _, _, _ = BuildSampleSlice(ent, site)           // nolint dogsled

	device := &RootDevice{
		Enterprises:          &models.OnfEnterprise_Enterprises{Enterprise: map[string]*Enterprise{"sample-ent": ent}},
		ConnectivityServices: &models.OnfConnectivityService_ConnectivityServices{ConnectivityService: map[string]*ConnectivityService{"sample-cs": cs}},
	}

	return device
}

// BuildScope creates a scope for the sample device
func BuildScope(device *RootDevice, entID string, siteID string, csID string) (*AetherScope, error) {
	enterprise, okay := device.Enterprises.Enterprise[entID]
	if !okay {
		return nil, fmt.Errorf("Failed to find enterprise %s", entID)
	}

	site, okay := enterprise.Site[siteID]
	if !okay {
		return nil, fmt.Errorf("Failed to find enterprise %s", siteID)
	}

	cs, okay := device.ConnectivityServices.ConnectivityService[csID]
	if !okay {
		return nil, fmt.Errorf("Failed to find cs %s", csID)
	}

	return &AetherScope{
		RootDevice:          device,
		ConnectivityService: cs,
		Enterprise:          enterprise,
		Site:                site,
	}, nil

}
