// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizer implements a synchronizer for converting sdcore gnmi to json
package synchronizer

import (
	"encoding/json"
	"fmt"
	"sort"
)

// SynchronizeDeviceGroup synchronizes a device group
func (s *Synchronizer) SynchronizeDeviceGroup(scope *AetherScope, dg *DeviceGroup) (int, error) {
	err := validateDeviceGroup(dg)
	if err != nil {
		return 0, fmt.Errorf("DeviceGroup %s failed validation: %v", *dg.DgId, err)
	}

	dgCore := deviceGroup{
		IPDomainName: *dg.DgId,
		SiteInfo:     *scope.Site.SiteId,
	}

	if scope.Site.ImsiDefinition == nil {
		return 0, fmt.Errorf("DeviceGroup %s site has nil ImsiDefinition", *dg.DgId)
	}
	err = validateImsiDefinition(scope.Site.ImsiDefinition)
	if err != nil {
		return 0, fmt.Errorf("DeviceGroup %s unable to determine Site.ImsiDefinition: %s", *dg.DgId, err)
	}

	// be deterministic...
	deviceLinkKeys := []string{}
	for k := range dg.Device {
		deviceLinkKeys = append(deviceLinkKeys, k)
	}
	sort.Strings(deviceLinkKeys)

	// populate the imsi list
	for _, k := range deviceLinkKeys {
		deviceID := dg.Device[k].DeviceId

		device, err := s.GetDevice(scope, deviceID)
		if err != nil {
			return 0, fmt.Errorf("DeviceGroup %s failed to get Device: %s", *dg.DgId, err)
		}

		if (device.SimCard == nil) || (*device.SimCard == "") {
			continue
		}

		simCard, err := s.GetSimCard(scope, device.SimCard)
		if err != nil {
			return 0, fmt.Errorf("DeviceGroup %s failed to get SimCard: %s", *dg.DgId, err)
		}

		imsi, err := FormatImsiDef(scope.Site.ImsiDefinition, *simCard.Imsi)
		if err != nil {
			return 0, fmt.Errorf("Failed to format IMSI in dg %s: %v", *dg.DgId, err)
		}
		dgCore.Imsis = append(dgCore.Imsis, fmt.Sprintf("%015d", imsi))
	}

	ipd, err := s.GetIPDomain(scope, dg.IpDomain)
	if err != nil {
		return 0, fmt.Errorf("DeviceGroup %s failed to get IpDomain: %s", *dg.DgId, err)
	}

	err = validateIPDomain(ipd)
	if err != nil {
		return 0, fmt.Errorf("DeviceGroup %s invalid: %s", *dg.DgId, err)
	}

	dgCore.IPDomainName = *ipd.IpId
	ipdCore := ipDomain{
		Dnn:          DerefStrPtr(ipd.Dnn, "internet"),
		Pool:         *ipd.Subnet,
		DNSPrimary:   DerefStrPtr(ipd.DnsPrimary, ""),
		DNSSecondary: DerefStrPtr(ipd.DnsSecondary, ""),
		Mtu:          DerefUint16Ptr(ipd.Mtu, DefaultMTU),
		Qos:          &ipdQos{Uplink: *dg.Mbr.Uplink, Downlink: *dg.Mbr.Downlink, Unit: aStr(DefaultBitrateUnit)},
	}
	dgCore.IPDomain = ipdCore

	rocTrafficClass, err := s.GetTrafficClass(scope, dg.Mbr.TrafficClass)
	if err != nil {
		return 0, fmt.Errorf("DG %s unable to determine traffic class: %s", *dg.DgId, err)
	}
	tcCore := &trafficClass{Name: *rocTrafficClass.TcId,
		PDB:  DerefUint16Ptr(rocTrafficClass.Pdb, 300),
		PELR: uint8(DerefInt8Ptr(rocTrafficClass.Pelr, 6)),
		QCI:  DerefUint8Ptr(rocTrafficClass.Qci, 9),
		ARP:  DerefUint8Ptr(rocTrafficClass.Arp, 9)}
	dgCore.IPDomain.Qos.TrafficClass = tcCore

	if s.partialUpdateEnable && s.CacheCheck(CacheModelDeviceGroup, *dg.DgId, dgCore) {
		log.Infof("Core Device-Group %s has not changed", *dg.DgId)
		return 0, nil
	}

	data, err := json.MarshalIndent(dgCore, "", "  ")
	if err != nil {
		return 0, fmt.Errorf("DeviceGroup %s failed to Marshal Json: %s", *dg.DgId, err)
	}

	url := fmt.Sprintf("%s/v1/device-group/%s", *scope.ConnectivityService.Core_5GEndpoint, *dg.DgId)
	err = s.pusher.PushUpdate(url, data)
	if err != nil {
		return 1, fmt.Errorf("DeviceGroup %s failed to Push update: %s", *dg.DgId, err)
	}

	s.CacheUpdate(CacheModelDeviceGroup, *dg.DgId, dgCore)

	return 0, nil
}
