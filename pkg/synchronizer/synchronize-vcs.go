// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizer implements a synchronizer for converting sdcore gnmi to json
package synchronizer

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	models "github.com/onosproject/config-models/modelplugin/aether-2.0.0/aether_2_0_0"
)

func (s *Synchronizer) mapPriority(i uint8) uint8 {
	return 255 - i // UPF expectation is priority is inverse of ROC's priority
}

// SynchronizeVcsCore synchronizes the VCSes
// Return a count of push-related errors
func (s *Synchronizer) SynchronizeVcsCore(device *models.Device, vcs *models.OnfVcs_Vcs_Vcs, cs *models.OnfConnectivityService_ConnectivityService_ConnectivityService, validEnterpriseIds map[string]bool) (int, error) {
	dgList, err := s.GetVcsDG(device, vcs)
	if err != nil {
		return 0, fmt.Errorf("Vcs %s unable to determine site: %s", *vcs.Id, err)
	}

	site, err := s.GetSite(device, vcs.Site)
	if err != nil {
		return 0, fmt.Errorf("Vcs %s unable to determine site: %s", *vcs.Id, err)
	}

	valid, okay := validEnterpriseIds[*site.Enterprise]
	if (!okay) || (!valid) {
		return 0, fmt.Errorf("VCS %s is not part of ConnectivityService %s", *vcs.Id, *cs.Id)
	}

	err = validateVcs(vcs)
	if err != nil {
		return 0, fmt.Errorf("Vcs %s is invalid: %s", *vcs.Id, err)
	}

	if site.ImsiDefinition == nil {
		return 0, fmt.Errorf("Vcs %s has nnil Site.ImsiDefinition", *vcs.Id)
	}
	err = validateImsiDefinition(site.ImsiDefinition)
	if err != nil {
		return 0, fmt.Errorf("Vcs %s unable to determine Site.ImsiDefinition: %s", *vcs.Id, err)
	}
	plmn := plmn{
		Mcc: *site.ImsiDefinition.Mcc,
		Mnc: *site.ImsiDefinition.Mnc,
	}
	siteInfo := siteInfo{
		SiteName: *site.Id,
		Plmn:     plmn,
	}

	if site.SmallCell != nil {
		// be deterministic...
		smallCellKeys := []string{}
		for k := range site.SmallCell {
			smallCellKeys = append(smallCellKeys, k)
		}
		sort.Strings(smallCellKeys)

		for _, k := range smallCellKeys {
			ap := site.SmallCell[k]
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

	if vcs.Upf != nil {
		aUpf, err := s.GetUpf(device, vcs.Upf)
		if err != nil {
			return 0, fmt.Errorf("Vcs %s unable to determine upf: %s", *vcs.Id, err)
		}
		err = validateUpf(aUpf)
		if err != nil {
			return 0, fmt.Errorf("Vcs %s Upf is invalid: %s", *vcs.Id, err)
		}
		siteInfo.Upf = upf{
			Name: *aUpf.Address,
			Port: *aUpf.Port,
		}
	}

	sliceID := sliceIDStruct{
		Sst: strconv.FormatUint(uint64(*vcs.Sst), 10),
	}

	// If the SD is unset, then do not set SD in the output. If it is set,
	// then emit it as a string of six hex digits.
	if vcs.Sd != nil {
		sliceID.Sd = fmt.Sprintf("%06X", *vcs.Sd)
	}

	slice := slice{
		ID:                        sliceID,
		SiteInfo:                  siteInfo,
		ApplicationFilteringRules: []appFilterRule{},
	}

	for _, dg := range dgList {
		slice.DeviceGroup = append(slice.DeviceGroup, *dg.Id)
	}

	// be deterministic...
	appKeys := []string{}
	for k := range vcs.Filter {
		appKeys = append(appKeys, k)
	}
	sort.Strings(appKeys)

	for _, k := range appKeys {
		appRef := vcs.Filter[k]
		app, err := s.GetApplication(device, appRef.Application)
		if err != nil {
			return 0, fmt.Errorf("Vcs %s unable to determine application: %s", *vcs.Id, err)
		}

		if (app.Address == nil) || (*app.Address == "") {
			// this is a temporary restriction
			return 0, fmt.Errorf("Vcs %s Application %s has empty address", *vcs.Id, *app.Id)
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
				Name: fmt.Sprintf("%s-%s", *app.Id, epName),
			}

			err = validateAppEndpoint(endpoint)
			if err != nil {
				log.Warnf("App %s invalid endpoint: %s", *app.Id, err)
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
					return 0, fmt.Errorf("Vcs %s Application %s unable to determine protocol: %s", *vcs.Id, *app.Id, err)
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
				rocTrafficClass, err := s.GetTrafficClass(device, endpoint.TrafficClass)
				if err != nil {
					return 0, fmt.Errorf("Vcs %s application %s unable to determine traffic class: %s", *vcs.Id, *app.Id, err)
				}
				tcCore := &trafficClass{Name: *rocTrafficClass.Id,
					PDB:  DerefUint16Ptr(rocTrafficClass.Pdb, 300),
					PELR: uint8(DerefInt8Ptr(rocTrafficClass.Pelr, 6)),
					QCI:  DerefUint8Ptr(rocTrafficClass.Qci, 9),
					ARP:  DerefUint8Ptr(rocTrafficClass.Arp, 9)}
				appCore.TrafficClass = tcCore
			}

			appCore.Priority = s.mapPriority(DerefUint8Ptr(appRef.Priority, 0))
			slice.ApplicationFilteringRules = append(slice.ApplicationFilteringRules, appCore)
		}
	}

	switch *vcs.DefaultBehavior {
	case "ALLOW-ALL":
		allowAll := appFilterRule{Name: "ALLOW-ALL", Action: "permit", Priority: s.mapPriority(250), Endpoint: "0.0.0.0/0"}
		slice.ApplicationFilteringRules = append(slice.ApplicationFilteringRules, allowAll)
	case "DENY-ALL":
		denyAll := appFilterRule{Name: "DENY-ALL", Action: "deny", Priority: s.mapPriority(250), Endpoint: "0.0.0.0/0"}
		slice.ApplicationFilteringRules = append(slice.ApplicationFilteringRules, denyAll)
	case "ALLOW-PUBLIC":
		denyClassA := appFilterRule{Name: "DENY-CLASS-A", Action: "deny", Priority: s.mapPriority(250), Endpoint: "10.0.0.0/8"}
		denyClassB := appFilterRule{Name: "DENY-CLASS-B", Action: "deny", Priority: s.mapPriority(251), Endpoint: "172.16.0.0/12"}
		denyClassC := appFilterRule{Name: "DENY-CLASS-C", Action: "deny", Priority: s.mapPriority(252), Endpoint: "192.168.0.0/16"}
		allowAll := appFilterRule{Name: "ALLOW-ALL", Action: "permit", Priority: s.mapPriority(253), Endpoint: "0.0.0.0/0"}
		slice.ApplicationFilteringRules = append(slice.ApplicationFilteringRules, denyClassA)
		slice.ApplicationFilteringRules = append(slice.ApplicationFilteringRules, denyClassB)
		slice.ApplicationFilteringRules = append(slice.ApplicationFilteringRules, denyClassC)
		slice.ApplicationFilteringRules = append(slice.ApplicationFilteringRules, allowAll)
	default:
		return 0, fmt.Errorf("Vcs %s has invalid defauilt-behavior %s", *vcs.Id, *vcs.DefaultBehavior)
	}

	if s.partialUpdateEnable && s.CacheCheck(CacheModelSlice, *vcs.Id, slice) {
		log.Infof("Core Slice %s has not changed", *vcs.Id)
		return 0, nil
	}

	data, err := json.MarshalIndent(slice, "", "  ")
	if err != nil {
		return 0, fmt.Errorf("Vcs %s failed to marshal JSON: %s", *vcs.Id, err)
	}

	url := fmt.Sprintf("%s/v1/network-slice/%s", *cs.Core_5GEndpoint, *vcs.Id)
	err = s.pusher.PushUpdate(url, data)
	if err != nil {
		return 1, fmt.Errorf("Vcs %s failed to push update: %s", *vcs.Id, err)
	}

	s.CacheUpdate(CacheModelSlice, *vcs.Id, slice)

	return 0, nil
}

// SynchronizeAllVcs synchronizes the VCSes
func (s *Synchronizer) SynchronizeAllVcs(device *models.Device, cs *models.OnfConnectivityService_ConnectivityService_ConnectivityService, validEnterpriseIds map[string]bool) int {
	pushFailures := 0
vcsLoop:
	// All errors are treated as nonfatal, logged, and synchronization continues with the next VCS.
	// PushFailures are counted and reported to the caller, who can decide whether to retry.
	for _, vcs := range device.Vcs.Vcs {
		corePushFailures, err := s.SynchronizeVcsCore(device, vcs, cs, validEnterpriseIds)
		pushFailures += corePushFailures
		if err != nil {
			log.Warnf("Vcs %s failed to synchronize Core: %s", *vcs.Id, err)
			continue vcsLoop
		}

		upfPushFailures, err := s.SynchronizeVcsUPF(device, vcs)
		pushFailures += upfPushFailures
		if err != nil {
			log.Warnf("Vcs %s failed to synchronize UPF: %s", *vcs.Id, err)
			continue vcsLoop
		}
	}

	return pushFailures
}
