// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

// Package synchronizer implements the synchronizer.
package synchronizer

import (
	"errors"
	"fmt"
	models "github.com/onosproject/aether-models/models/aether-2.1.x/v2/api"
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	"testing"
	"time"
)

var (
	mockSynchronizeDeviceCalls         []*gnmi.ConfigForest // list of calls to MockSynchronizeDevice that succeeded
	mockSynchronizeDeviceFails         []*gnmi.ConfigForest // list of calls to MockSynchronizeDevice that failed
	mockSynchronizeDevicePushFails     []*gnmi.ConfigForest // list of calls to MockSynchronizeDevice that had a push failure
	mockSynchronizeDeviceFailCount     int                  // Cause MockSynchronizeDevice to fail the specified number of times
	mockSynchronizeDevicePushFailCount int                  // Cause MockSynchronizeDevice to fail to push the specified number of times
	mockSynchronizeDeviceDelay         time.Duration        // Cause MockSynchronizeDevice to take some time
)

func mockSynchronizeDevice(config *gnmi.ConfigForest) (int, error) {
	time.Sleep(mockSynchronizeDeviceDelay)
	if mockSynchronizeDeviceFailCount > 0 {
		mockSynchronizeDeviceFailCount--
		mockSynchronizeDeviceFails = append(mockSynchronizeDeviceFails, config)
		return 0, errors.New("Mock error")
	}
	if mockSynchronizeDevicePushFailCount > 0 {
		mockSynchronizeDevicePushFailCount--
		mockSynchronizeDevicePushFails = append(mockSynchronizeDevicePushFails, config)
		return 1, nil
	}
	mockSynchronizeDeviceCalls = append(mockSynchronizeDeviceCalls, config)
	return 0, nil
}

// Reset mockSynchronizeDevice for a new set of tests
//
//	failCount = number of times to fail before returning success
//	pushFailCount = number of times to fail to push before returning success
//	delay = amount of time to delay before returning
func mockSynchronizeDeviceReset(failCount int, pushFailCount int, delay time.Duration) {
	mockSynchronizeDeviceCalls = nil
	mockSynchronizeDeviceFails = nil
	mockSynchronizeDevicePushFails = nil
	mockSynchronizeDeviceFailCount = failCount
	mockSynchronizeDevicePushFailCount = pushFailCount
	mockSynchronizeDeviceDelay = delay
}

// Wait for the synchronizer to be idle. Used in unit tests to perform asserts
// when a predictable state is reached.
func waitForSyncIdle(t *testing.T, s *Synchronizer, timeout time.Duration) {
	elapsed := 0 * time.Second
	for {
		if s.isIdle() {
			return
		}
		time.Sleep(100 * time.Millisecond)
		elapsed += 100 * time.Millisecond
		if elapsed > timeout {
			t.Fatal("waitForSyncIdle failed to complete")
		}
	}
}

// TODO: Refactor to use mockgen and/or sample data files ?

// BuildSampleDeviceGroup builds a sample device group for unit testing
func BuildSampleDeviceGroup() (
	map[string]*TrafficClass,
	*IpDomain,
	*Site,
	*DeviceGroup) {

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
		Imsi:  aStr("123456789012345"),
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
	cs := &ConnectivityService{
		Core_5G: &Core5G{Endpoint: aStr("http://5gcore")},
	}
	site := &Site{
		Description:         aStr("sample-site-desc"),
		DisplayName:         aStr("sample-site-dn"),
		SiteId:              aStr("sample-site"),
		ImsiDefinition:      imsiDef,
		Device:              map[string]*Device{*device.DeviceId: device},
		SimCard:             map[string]*SimCard{*simCard.SimId: simCard},
		SmallCell:           map[string]*SmallCell{*sc.SmallCellId: sc},
		DeviceGroup:         map[string]*DeviceGroup{*dg.DeviceGroupId: dg},
		IpDomain:            map[string]*IpDomain{*ipd.IpDomainId: ipd},
		ConnectivityService: cs,
	}

	tcList := map[string]*TrafficClass{*tc.TrafficClassId: tc}

	return tcList, ipd, site, dg
}

// BuildSampleSlice builds a sample slice and application for testing. The
// Slice and its related artifacts are attached to the Enterprise and Site
// that are passed as arguments.
func BuildSampleSlice(site *Site) (
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
		Sd:          aStr("111"),
		Sst:         aStr("222"),
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
		Filter:              map[string]*SliceFilter{"sample-app": appLink, "sample-app2": app2Link},
		Description:         aStr("sample-slice-desc"),
		DeviceGroup:         map[string]*SliceDeviceGroup{"sample-dg": dgLink},
		DisplayName:         aStr("sample-app-dn"),
		SliceId:             aStr("sample-slice"),
		Sd:                  aStr("111"),
		Sst:                 aStr("222"),
		Mbr:                 sliceMbr,
		Upf:                 aStr("sample-upf"),
		DefaultBehavior:     aStr("DENY-ALL"),
		ConnectivityService: ConnectivityService5G,
	}

	apps := map[string]*Application{*app1.ApplicationId: app1, *app2.ApplicationId: app2}

	site.Slice = map[string]*Slice{*slice.SliceId: slice}
	site.Upf = map[string]*Upf{*upf.UpfId: upf}

	return apps, tp, upf, slice
}

// BuildSampleDevice builds a sample device, with VCS and Device-Group
func BuildSampleDevice() *RootDevice {
	tcList, _, site, _ := BuildSampleDeviceGroup() // nolint dogsled
	apps, tp, _, _ := BuildSampleSlice(site)       // nolint dogsled

	device := &RootDevice{
		Site:         map[string]*models.OnfSite_Site{*site.SiteId: site},
		TrafficClass: tcList,
		Template:     map[string]*models.OnfTemplate_Template{*tp.TemplateId: tp},
		Application:  apps,
	}

	return device
}

func BuildSampleConfig() (*gnmi.ConfigForest, *RootDevice) {
	device := BuildSampleDevice()
	config := gnmi.NewConfigForest()
	config.Configs["sample-ent"] = device
	return config, device
}

// BuildScope creates a scope for the sample device
func BuildScope(device *RootDevice, entID string, siteID string) (*AetherScope, error) {
	site, okay := device.Site[siteID]
	if !okay {
		return nil, fmt.Errorf("Failed to find sites %s", siteID)
	}

	return &AetherScope{
		EnterpriseId: &entID,
		Enterprise:   device,
		Site:         site,
	}, nil

}
