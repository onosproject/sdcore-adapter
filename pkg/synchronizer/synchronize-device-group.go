// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

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
		return 0, fmt.Errorf("DeviceGroup %s failed validation: %v", *dg.DeviceGroupId, err)
	}

	dgCore := deviceGroup{
		IPDomainName: *dg.DeviceGroupId,
		SiteInfo:     *scope.Site.SiteId,
	}

	if scope.Site.ImsiDefinition == nil {
		return 0, fmt.Errorf("DeviceGroup %s site has nil ImsiDefinition", *dg.DeviceGroupId)
	}
	err = validateImsiDefinition(scope.Site.ImsiDefinition)
	if err != nil {
		return 0, fmt.Errorf("DeviceGroup %s unable to determine Site.ImsiDefinition: %s", *dg.DeviceGroupId, err)
	}

	// be deterministic...
	deviceLinkKeys := []string{}
	for k := range dg.Device {
		deviceLinkKeys = append(deviceLinkKeys, k)
	}
	sort.Strings(deviceLinkKeys)

	// Make sure no devices gets passed as an empty list rather than None.
	// SD-Core would ignore the None, but it will act on the empty list.
	dgCore.Imsis = []string{}

	// populate the imsi list
	for _, k := range deviceLinkKeys {
		deviceID := dg.Device[k].DeviceId

		// The default for Enable, if unspecified, is True
		if (dg.Device[k].Enable != nil) && (!*dg.Device[k].Enable) {
			continue
		}

		device, err := s.GetDevice(scope, deviceID)
		if err != nil {
			return 0, fmt.Errorf("DeviceGroup %s failed to get Device: %s", *dg.DeviceGroupId, err)
		}

		if (device.SimCard == nil) || (*device.SimCard == "") {
			continue
		}

		simCard, err := s.GetSimCard(scope, device.SimCard)
		if err != nil {
			return 0, fmt.Errorf("DeviceGroup %s failed to get SimCard: %s", *dg.DeviceGroupId, err)
		}

		imsi, err := FormatImsiDef(scope.Site.ImsiDefinition, *simCard.Imsi)
		if err != nil {
			return 0, fmt.Errorf("Failed to format IMSI in dg %s: %v", *dg.DeviceGroupId, err)
		}
		dgCore.Imsis = append(dgCore.Imsis, fmt.Sprintf("%015d", imsi))
	}

	ipd, err := s.GetIPDomain(scope, dg.IpDomain)
	if err != nil {
		return 0, fmt.Errorf("DeviceGroup %s failed to get IpDomain: %s", *dg.DeviceGroupId, err)
	}

	err = validateIPDomain(ipd)
	if err != nil {
		return 0, fmt.Errorf("DeviceGroup %s IPDomain %s is invalid: %s", *dg.DeviceGroupId, *ipd.IpDomainId, err)
	}

	dgCore.IPDomainName = *ipd.IpDomainId
	ipdCore := ipDomain{
		Dnn:          DerefStrPtr(ipd.Dnn, "internet"),
		Pool:         *ipd.Subnet,
		DNSPrimary:   DerefStrPtr(ipd.DnsPrimary, ""),
		DNSSecondary: DerefStrPtr(ipd.DnsSecondary, ""),
		Mtu:          DerefUint16Ptr(ipd.Mtu, DefaultMTU),
		Qos:          &ipdQos{Uplink: *dg.Mbr.Uplink, Downlink: *dg.Mbr.Downlink, Unit: aStr(DefaultBitrateUnit)},
	}
	dgCore.IPDomain = ipdCore

	rocTrafficClass, err := s.GetTrafficClass(scope, dg.TrafficClass)
	if err != nil {
		return 0, fmt.Errorf("DG %s unable to determine traffic class: %s", *dg.DeviceGroupId, err)
	}
	tcCore := &trafficClass{Name: *rocTrafficClass.TrafficClassId,
		PDB:  DerefUint16Ptr(rocTrafficClass.Pdb, 300),
		PELR: uint8(DerefInt8Ptr(rocTrafficClass.Pelr, 6)),
		QCI:  DerefUint8Ptr(rocTrafficClass.Qci, 9),
		ARP:  DerefUint8Ptr(rocTrafficClass.Arp, 9)}
	dgCore.IPDomain.Qos.TrafficClass = tcCore

	if s.partialUpdateEnable && s.CacheCheck(CacheModelDeviceGroup, *dg.DeviceGroupId, dgCore) {
		log.Infof("Core Device-Group %s has not changed", *dg.DeviceGroupId)
		return 0, nil
	}

	data, err := json.MarshalIndent(dgCore, "", "  ")
	if err != nil {
		return 0, fmt.Errorf("DeviceGroup %s failed to Marshal Json: %s", *dg.DeviceGroupId, err)
	}

	if scope.CoreEndpoint == nil {
		return 0, fmt.Errorf("Device Group %s found no Core Endpoint", *dg.DeviceGroupId)
	}

	url := fmt.Sprintf("%s/v1/device-group/%s", *scope.CoreEndpoint, *dg.DeviceGroupId)
	log.Infof("Push Device-Group to %s", url)
	err = s.pusher.PushUpdate(url, data)
	if err != nil {
		return 1, fmt.Errorf("DeviceGroup %s failed to Push update: %s", *dg.DeviceGroupId, err)
	}

	s.CacheUpdate(CacheModelDeviceGroup, *dg.DeviceGroupId, dgCore)

	return 0, nil
}
