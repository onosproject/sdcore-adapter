// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

// Package synchronizer implements a synchronizer for converting sdcore gnmi to json
package synchronizer

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func (s *Synchronizer) mapPriority(i uint8) uint8 {
	return i // // At one point priority was flipped but this was incorrect
}

// SynchronizeSlice synchronizes the VCSes
// Return a count of push-related errors
func (s *Synchronizer) SynchronizeSlice(scope *AetherScope, slice *Slice) (int, error) {
	dgList, err := s.GetSliceDG(scope, slice)
	if err != nil {
		return 0, fmt.Errorf("Slice %s unable to determine site: %s", *slice.SliceId, err)
	}

	err = validateSlice(slice)
	if err != nil {
		return 0, fmt.Errorf("Slice %s is invalid: %s", *slice.SliceId, err)
	}

	if scope.Site.ImsiDefinition == nil {
		return 0, fmt.Errorf("Slice %s Site %s has nil Site.ImsiDefinition", *slice.SliceId, *scope.Site.SiteId)
	}
	err = validateImsiDefinition(scope.Site.ImsiDefinition)
	if err != nil {
		return 0, fmt.Errorf("Slice %s unable to determine Site.ImsiDefinition: %s", *slice.SliceId, err)
	}
	plmn := plmn{
		Mcc: *scope.Site.ImsiDefinition.Mcc,
		Mnc: *scope.Site.ImsiDefinition.Mnc,
	}
	siteInfo := siteInfo{
		SiteName: *scope.Site.SiteId,
		Plmn:     plmn,
	}

	if scope.Site.SmallCell != nil {
		// be deterministic...
		smallCellKeys := []string{}
		for k := range scope.Site.SmallCell {
			smallCellKeys = append(smallCellKeys, k)
		}
		sort.Strings(smallCellKeys)

		for _, k := range smallCellKeys {
			ap := scope.Site.SmallCell[k]
			err = validateSmallCell(ap)
			if err != nil {
				return 0, fmt.Errorf("SmallCell invalid: %s", err)
			}
			if *ap.Enable {
				tac, err := strconv.ParseUint(*ap.Tac, 16, 32)
				if err != nil {
					return 0, fmt.Errorf("SmallCell Failed to convert tac %s to integer: %v", *ap.Tac, err)
				}
				gNodeB := gNodeB{
					Name: *ap.Address,
					Tac:  uint32(tac),
				}
				siteInfo.GNodeBs = append(siteInfo.GNodeBs, gNodeB)
			}
		}
	}

	if slice.Upf != nil {
		aUpf, err := s.GetUpf(scope, slice.Upf)
		if err != nil {
			return 0, fmt.Errorf("Slice %s unable to determine upf: %s", *slice.SliceId, err)
		}
		err = validateUpf(aUpf)
		if err != nil {
			return 0, fmt.Errorf("Slice %s Upf is invalid: %s", *slice.SliceId, err)
		}
		siteInfo.Upf = upf{
			Name: *aUpf.Address,
			Port: *aUpf.Port,
		}
	}

	sliceID := sliceIDStruct{
		Sst: strconv.FormatUint(uint64(*slice.Sst), 10),
	}

	// If the SD is unset, then do not set SD in the output. If it is set,
	// then emit it as a string of six hex digits.
	if slice.Sd != nil {
		sliceID.Sd = fmt.Sprintf("%06X", *slice.Sd)
	}

	coreSlice := coreSlice{
		ID:                        sliceID,
		SiteInfo:                  siteInfo,
		ApplicationFilteringRules: []appFilterRule{},
	}

	for _, dg := range dgList {
		coreSlice.DeviceGroup = append(coreSlice.DeviceGroup, *dg.DeviceGroupId)
	}

	// be deterministic...
	appKeys := []string{}
	for k := range slice.Filter {
		appKeys = append(appKeys, k)
	}
	sort.Strings(appKeys)

	for _, k := range appKeys {
		appRef := slice.Filter[k]
		app, err := s.GetApplication(scope, appRef.Application)
		if err != nil {
			return 0, fmt.Errorf("Slice %s unable to determine application: %s", *slice.SliceId, err)
		}

		if (app.Address == nil) || (*app.Address == "") {
			// this is a temporary restriction
			return 0, fmt.Errorf("Slice %s Application %s has empty address", *slice.SliceId, *app.ApplicationId)
		}

		// be deterministic...
		epKeys := []string{}
		for k := range app.Endpoint {
			epKeys = append(epKeys, k)
		}
		sort.Strings(epKeys)

		for _, epName := range epKeys {
			endpoint := app.Endpoint[epName]
			appCore := appFilterRule{
				Name: fmt.Sprintf("%s-%s", *app.ApplicationId, epName),
			}

			if strings.Contains(*app.Address, "/") {
				appCore.Endpoint = *app.Address
			} else {
				appCore.Endpoint = *app.Address + "/32"
			}

			if endpoint.PortStart != nil {
				appCore.DestPortStart = endpoint.PortStart
				if endpoint.PortEnd != nil {
					appCore.DestPortEnd = endpoint.PortEnd
				} else {
					// no EndPort specified -- assume it's a singleton range
					appCore.DestPortEnd = appCore.DestPortStart
				}
			}

			if endpoint.Protocol != nil {
				protoNum, err := ProtoStringToProtoNumber(*endpoint.Protocol)
				if err != nil {
					return 0, fmt.Errorf("Slice %s Application %s unable to determine protocol: %s", *slice.SliceId, *app.ApplicationId, err)
				}
				appCore.Protocol = &protoNum
			}

			if (appRef.Allow != nil) && (*appRef.Allow) {
				appCore.Action = "permit"
			} else {
				appCore.Action = "deny"
			}

			hasQos := false
			if endpoint.Mbr != nil {
				if endpoint.Mbr.Uplink != nil {
					appCore.Uplink = *endpoint.Mbr.Uplink
					hasQos = true
				}
				if endpoint.Mbr.Downlink != nil {
					appCore.Downlink = *endpoint.Mbr.Downlink
					hasQos = true
				}
			}

			if hasQos {
				appCore.Unit = aStr(DefaultBitrateUnit)
			}

			if endpoint.TrafficClass != nil {
				rocTrafficClass, err := s.GetTrafficClass(scope, endpoint.TrafficClass)
				if err != nil {
					return 0, fmt.Errorf("Slice %s application %s unable to determine traffic class: %s", *slice.SliceId, *app.ApplicationId, err)
				}
				tcCore := &trafficClass{Name: *rocTrafficClass.TrafficClassId,
					PDB:  DerefUint16Ptr(rocTrafficClass.Pdb, 300),
					PELR: uint8(DerefInt8Ptr(rocTrafficClass.Pelr, 6)),
					QCI:  DerefUint8Ptr(rocTrafficClass.Qci, 9),
					ARP:  DerefUint8Ptr(rocTrafficClass.Arp, 9)}
				appCore.TrafficClass = tcCore
			}

			appCore.Priority = s.mapPriority(DerefUint8Ptr(appRef.Priority, 0))
			coreSlice.ApplicationFilteringRules = append(coreSlice.ApplicationFilteringRules, appCore)
		}
	}

	switch *slice.DefaultBehavior {
	case "ALLOW-ALL":
		allowAll := appFilterRule{Name: "ALLOW-ALL", Action: "permit", Priority: s.mapPriority(250), Endpoint: "0.0.0.0/0"}
		coreSlice.ApplicationFilteringRules = append(coreSlice.ApplicationFilteringRules, allowAll)
	case "DENY-ALL":
		denyAll := appFilterRule{Name: "DENY-ALL", Action: "deny", Priority: s.mapPriority(250), Endpoint: "0.0.0.0/0"}
		coreSlice.ApplicationFilteringRules = append(coreSlice.ApplicationFilteringRules, denyAll)
	case "ALLOW-PUBLIC":
		denyClassA := appFilterRule{Name: "DENY-CLASS-A", Action: "deny", Priority: s.mapPriority(250), Endpoint: "10.0.0.0/8"}
		denyClassB := appFilterRule{Name: "DENY-CLASS-B", Action: "deny", Priority: s.mapPriority(251), Endpoint: "172.16.0.0/12"}
		denyClassC := appFilterRule{Name: "DENY-CLASS-C", Action: "deny", Priority: s.mapPriority(252), Endpoint: "192.168.0.0/16"}
		allowAll := appFilterRule{Name: "ALLOW-ALL", Action: "permit", Priority: s.mapPriority(253), Endpoint: "0.0.0.0/0"}
		coreSlice.ApplicationFilteringRules = append(coreSlice.ApplicationFilteringRules, denyClassA)
		coreSlice.ApplicationFilteringRules = append(coreSlice.ApplicationFilteringRules, denyClassB)
		coreSlice.ApplicationFilteringRules = append(coreSlice.ApplicationFilteringRules, denyClassC)
		coreSlice.ApplicationFilteringRules = append(coreSlice.ApplicationFilteringRules, allowAll)
	default:
		return 0, fmt.Errorf("Slice %s has invalid defauilt-behavior %s", *slice.SliceId, *slice.DefaultBehavior)
	}

	if s.partialUpdateEnable && s.CacheCheck(CacheModelSlice, *slice.SliceId, coreSlice) {
		log.Infof("Core Slice %s has not changed", *slice.SliceId)
		return 0, nil
	}

	data, err := json.MarshalIndent(coreSlice, "", "  ")
	if err != nil {
		return 0, fmt.Errorf("Slice %s failed to marshal JSON: %s", *slice.SliceId, err)
	}

	if scope.ConnectivityService.Core_5GEndpoint == nil {
		return 0, fmt.Errorf("Slice %s Connectivity Service %s has no Core Endpoint", *slice.SliceId, *scope.ConnectivityService.ConnectivityServiceId)
	}

	url := fmt.Sprintf("%s/v1/network-slice/%s", *scope.ConnectivityService.Core_5GEndpoint, *slice.SliceId)
	err = s.pusher.PushUpdate(url, data)
	if err != nil {
		return 1, fmt.Errorf("Slice %s failed to push update: %s", *slice.SliceId, err)
	}

	s.CacheUpdate(CacheModelSlice, *slice.SliceId, coreSlice)

	return 0, nil
}
