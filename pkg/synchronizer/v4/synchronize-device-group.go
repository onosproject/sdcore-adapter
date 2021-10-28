// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizerv4 implements a synchronizer for converting sdcore gnmi to json
package synchronizerv4

import (
	"encoding/json"
	"fmt"

	models "github.com/onosproject/config-models/modelplugin/aether-4.0.0/aether_4_0_0"
	"github.com/onosproject/sdcore-adapter/pkg/synchronizer"
)

// SynchronizeDeviceGroup synchronizes a device group
func (s *Synchronizer) SynchronizeDeviceGroup(device *models.Device, dg *models.OnfDeviceGroup_DeviceGroup_DeviceGroup, cs *models.OnfConnectivityService_ConnectivityService_ConnectivityService, validEnterpriseIds map[string]bool) (int, error) {
	err := validateDeviceGroup(dg)
	if err != nil {
		return 0, fmt.Errorf("DeviceGroup %s failed validation: %v", *dg.Id, err)
	}

	site, err := s.GetDeviceGroupSite(device, dg)
	if err != nil {
		return 0, fmt.Errorf("DeviceGroup %s unable to determine site: %s", *dg.Id, err)
	}

	valid, okay := validEnterpriseIds[*site.Enterprise]
	if (!okay) || (!valid) {
		return 0, fmt.Errorf("DeviceGroup %s is not part of ConnectivityService %s", *dg.Id, *cs.Id)
	}

	dgCore := deviceGroup{
		IPDomainName: *dg.Id,
		SiteInfo:     *dg.Site,
	}

	if site.ImsiDefinition == nil {
		return 0, fmt.Errorf("DeviceGroup %s site has nil ImsiDefinition", *dg.Id)
	}
	err = validateImsiDefinition(site.ImsiDefinition)
	if err != nil {
		return 0, fmt.Errorf("DeviceGroup %s unable to determine Site.ImsiDefinition: %s", *dg.Id, err)
	}

	// populate the imsi list
	for _, imsiBlock := range dg.Imsis {
		if imsiBlock.ImsiRangeFrom == nil {
			return 0, fmt.Errorf("imsiBlock has blank ImsiRangeFrom: %v", imsiBlock)
		}
		var firstImsi uint64
		firstImsi, err = FormatImsiDef(site.ImsiDefinition, *imsiBlock.ImsiRangeFrom)
		if err != nil {
			return 0, fmt.Errorf("Failed to format IMSI in dg %s: %v", *dg.Id, err)
		}
		var lastImsi uint64
		if imsiBlock.ImsiRangeTo == nil {
			lastImsi = firstImsi
		} else {
			lastImsi, err = FormatImsiDef(site.ImsiDefinition, *imsiBlock.ImsiRangeTo)
			if err != nil {
				return 0, fmt.Errorf("Failed to format IMSI in dg %s: %v", *dg.Id, err)
			}

		}
		for i := firstImsi; i <= lastImsi; i++ {
			dgCore.Imsis = append(dgCore.Imsis, fmt.Sprintf("%015d", i))
		}
	}

	ipd, err := s.GetIPDomain(device, dg.IpDomain)
	if err != nil {
		return 0, fmt.Errorf("DeviceGroup %s failed to get IpDomain: %s", *dg.Id, err)
	}

	err = validateIPDomain(ipd)
	if err != nil {
		return 0, fmt.Errorf("DeviceGroup %s invalid: %s", *dg.Id, err)
	}

	dgCore.IPDomainName = *ipd.Id
	ipdCore := ipDomain{
		Dnn:          synchronizer.DerefStrPtr(ipd.Dnn, "internet"),
		Pool:         *ipd.Subnet,
		DNSPrimary:   synchronizer.DerefStrPtr(ipd.DnsPrimary, ""),
		DNSSecondary: synchronizer.DerefStrPtr(ipd.DnsSecondary, ""),
		Mtu:          synchronizer.DerefUint16Ptr(ipd.Mtu, DefaultMTU),
		Qos:          &ipdQos{Uplink: *dg.Device.Mbr.Uplink, Downlink: *dg.Device.Mbr.Downlink, Unit: aStr(DefaultBitrateUnit)},
	}
	dgCore.IPDomain = ipdCore

	rocTrafficClass, err := s.GetTrafficClass(device, dg.Device.TrafficClass)
	if err != nil {
		return 0, fmt.Errorf("DG %s unable to determine traffic class: %s", *dg.Id, err)
	}
	tcCore := &trafficClass{Name: *rocTrafficClass.Id,
		PDB:  synchronizer.DerefUint16Ptr(rocTrafficClass.Pdb, 300),
		PELR: uint8(synchronizer.DerefInt8Ptr(rocTrafficClass.Pelr, 6)),
		QCI:  synchronizer.DerefUint8Ptr(rocTrafficClass.Qci, 9),
		ARP:  synchronizer.DerefUint8Ptr(rocTrafficClass.Arp, 9)}
	dgCore.IPDomain.Qos.TrafficClass = tcCore

	data, err := json.MarshalIndent(dgCore, "", "  ")
	if err != nil {
		return 0, fmt.Errorf("DeviceGroup %s failed to Marshal Json: %s", *dg.Id, err)
	}

	url := fmt.Sprintf("%s/v1/device-group/%s", *cs.Core_5GEndpoint, *dg.Id)
	err = s.pusher.PushUpdate(url, data)
	if err != nil {
		return 1, fmt.Errorf("DeviceGroup %s failed to Push update: %s", *dg.Id, err)
	}
	return 0, nil
}

// SynchronizeAllDeviceGroups synchronizes the device groups
func (s *Synchronizer) SynchronizeAllDeviceGroups(device *models.Device, cs *models.OnfConnectivityService_ConnectivityService_ConnectivityService, validEnterpriseIds map[string]bool) int {
	pushFailures := 0
deviceGroupLoop:
	// All errors are treated as nonfatal, logged, and synchronization continues with the next device-group.
	// PushFailures are counted and reported to the caller, who can decide whether to retry.
	for _, dg := range device.DeviceGroup.DeviceGroup {
		dgFailures, err := s.SynchronizeDeviceGroup(device, dg, cs, validEnterpriseIds)
		pushFailures += dgFailures
		if err != nil {
			log.Warnf("DG %s failed to synchronize Core: %s", *dg.Id, err)
			continue deviceGroupLoop
		}
	}
	return pushFailures
}
