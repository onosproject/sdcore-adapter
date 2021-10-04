// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizerv4 implements a synchronizer for converting sdcore gnmi to json
package synchronizerv4

import (
	"encoding/json"
	"fmt"
	"github.com/openconfig/ygot/ygot"
	"sort"
	"strconv"
	"strings"
	"time"

	models "github.com/onosproject/config-models/modelplugin/aether-4.0.0/aether_4_0_0"
	"github.com/onosproject/sdcore-adapter/pkg/synchronizer"
)

// Ideally we would get these from the yang defaults
const (
	// DefaultAdminStatus is the default for the AdminStatus Field
	DefaultAdminStatus = "ENABLE"

	// DefaultMTU is the default for the MTU field
	DefaultMTU = 1492

	// DefaultProtocol is the default for the Protocol field
	DefaultProtocol = "TCP"
)

type ipdTrafficClass struct {
	Name string `json:"name"`
	QCI  uint8  `json:"qci"`
	ARP  uint8  `json:"arp"`
	PDB  uint16 `json:"pdb"`
	PELR uint8  `json:"pelr"`
}

type ipdQos struct {
	Uplink       uint64          `json:"dnn-mbr-uplink"`
	Downlink     uint64          `json:"dnn-mbr-downlink"`
	TrafficClass ipdTrafficClass `json:"traffic-class"`
}

type ipDomain struct {
	Dnn          string  `json:"dnn"`
	Pool         string  `json:"ue-ip-pool"`
	DNSPrimary   string  `json:"dns-primary"`
	DNSSecondary string  `json:"dns-secondary,omitempty"`
	Mtu          uint16  `json:"mtu"`
	Qos          *ipdQos `json:"ue-dnn-qos,omitempty"`
}

type deviceGroup struct {
	Imsis        []string `json:"imsis"`
	IPDomainName string   `json:"ip-domain-name"`
	SiteInfo     string   `json:"site-info"`
	IPDomain     ipDomain `json:"ip-domain-expanded"`
}

type sliceIDStruct struct {
	Sst string `json:"sst"`
	Sd  string `json:"sd"`
}

type gNodeB struct {
	Name string `json:"name"`
	Tac  uint32 `json:"tac"`
}

type plmn struct {
	Mcc string `json:"mcc"`
	Mnc string `json:"mnc"`
}

type upf struct {
	Name string `json:"upf-name"`
	Port uint16 `json:"upf-port"`
}

type siteInfo struct {
	SiteName string   `json:"site-name"`
	Plmn     plmn     `json:"plmn"`
	GNodeBs  []gNodeB `json:"gNodeBs"`
	Upf      upf      `json:"upf"`
}

type appFilterRule struct {
	Name          string `json:"rule-name"`
	Priority      uint8  `json:"priority"`
	Action        string `json:"action"`
	DestNetwork   string `json:"dest-network"`
	DestPortStart uint16 `json:"dest-port-start"`
	DestPortEnd   uint16 `json:"dest-port-end"`
	Protocol      uint32 `json:"protocol"`
}

type slice struct {
	ID                        sliceIDStruct   `json:"slice-id"`
	DeviceGroup               []string        `json:"site-device-group"`
	SiteInfo                  siteInfo        `json:"site-info"`
	ApplicationFilteringRules []appFilterRule `json:"application-filtering-rules"`
}

// SynchronizeDevice synchronizes a device
// NOTE: This function is nearly identical with the v2 synchronizer. Refactor?
func (s *Synchronizer) SynchronizeDevice(config ygot.ValidatedGoStruct) error {
	device := config.(*models.Device)

	if device.Enterprise == nil {
		log.Info("No enteprises")
		return nil
	}

	if device.ConnectivityService == nil {
		log.Info("No connectivity services")
		return nil
	}

	// For a given ConnectivityService, we want to know the list of Enterprises
	// that use it. Precompute this so we can pass a list of valid Enterprises
	// along to SynchronizeConnectivityService.
	csEntMap := map[string]map[string]bool{}
	for entID, ent := range device.Enterprise.Enterprise {
		for csID := range ent.ConnectivityService {
			m, okay := csEntMap[csID]
			if !okay {
				m = map[string]bool{}
				csEntMap[csID] = m
			}
			m[entID] = true
		}
	}

	errors := []error{}
	for csID, cs := range device.ConnectivityService.ConnectivityService {
		if (cs.Core_5GEndpoint == nil) || (*cs.Core_5GEndpoint == "") {
			log.Warnf("Skipping connectivity service %s because it has no 5G Endpoint", *cs.Id)
			continue
		}

		// Get the list of valid Enterprises for this CS.
		// Note: This could return an empty map if there is a CS that no
		//   enterprises are linked to . In that case, we can still push models
		//   that are not directly related to an enterprise, such as profiles.
		m := csEntMap[csID]

		tStart := time.Now()
		synchronizer.KpiSynchronizationTotal.WithLabelValues(csID).Inc()

		err := s.SynchronizeConnectivityService(device, cs, m)
		if err != nil {
			synchronizer.KpiSynchronizationFailedTotal.WithLabelValues(csID).Inc()
			// If there are errors, then build a list of them and continue to try
			// to synchronize other connectivity services.
			errors = append(errors, err)
		} else {
			synchronizer.KpiSynchronizationDuration.WithLabelValues(csID).Observe(time.Since(tStart).Seconds())
		}
	}

	if len(errors) == 0 {
		return nil
	}

	return fmt.Errorf("synchronization errors: %v", errors)
}

// SynchronizeConnectivityService synchronizes a connectivity service
func (s *Synchronizer) SynchronizeConnectivityService(device *models.Device, cs *models.OnfConnectivityService_ConnectivityService_ConnectivityService, validEnterpriseIds map[string]bool) error {
	log.Infof("Synchronizing Connectivity Service %s", *cs.Id)

	if device.DeviceGroup != nil {
		err := s.SynchronizeDeviceGroups(device, cs, validEnterpriseIds)
		if err != nil {
			return err
		}
	}
	if device.Vcs != nil {
		err := s.SynchronizeVcs(device, cs, validEnterpriseIds)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetIPDomain looks up an IpDomain
func (s *Synchronizer) GetIPDomain(device *models.Device, id *string) (*models.OnfIpDomain_IpDomain_IpDomain, error) {
	if device.IpDomain == nil {
		return nil, fmt.Errorf("Device contains no IpDomains")
	}
	if (id == nil) || (*id == "") {
		return nil, fmt.Errorf("IpDomain id is blank")
	}
	ipd, okay := device.IpDomain.IpDomain[*id]
	if !okay {
		return nil, fmt.Errorf("IpDomain %s not found", *id)
	}
	return ipd, nil
}

// GetReferencingVCS returns a VCS (the first one it finds) that uses a device group
func (s *Synchronizer) GetReferencingVCS(device *models.Device, dg *models.OnfDeviceGroup_DeviceGroup_DeviceGroup) *models.OnfVcs_Vcs_Vcs {
	if device.Vcs == nil {
		return nil
	}
	for _, vcs := range device.Vcs.Vcs {
		for _, dgLink := range vcs.DeviceGroup {
			if !*dgLink.Enable {
				continue
			}
			if *dgLink.DeviceGroup == *dg.Id {
				return vcs
			}
		}
	}
	return nil
}

// GetUpf looks up a Upf
func (s *Synchronizer) GetUpf(device *models.Device, id *string) (*models.OnfUpf_Upf_Upf, error) {
	if device.Upf == nil {
		return nil, fmt.Errorf("Device contains no Upfs")
	}
	if (id == nil) || (*id == "") {
		return nil, fmt.Errorf("Upf id is blank")
	}
	upf, okay := device.Upf.Upf[*id]
	if !okay {
		return nil, fmt.Errorf("Upf %s not found", *id)
	}
	return upf, nil
}

// GetApplication looks up an application
func (s *Synchronizer) GetApplication(device *models.Device, id *string) (*models.OnfApplication_Application_Application, error) {
	if device.Application == nil {
		return nil, fmt.Errorf("Device contains no Applications")
	}
	if (id == nil) || (*id == "") {
		return nil, fmt.Errorf("Application id is blank")
	}
	app, okay := device.Application.Application[*id]
	if !okay {
		return nil, fmt.Errorf("Application %s not found", *id)
	}
	return app, nil
}

// GetTrafficClass looks up a TrafficClass
func (s *Synchronizer) GetTrafficClass(device *models.Device, id *string) (*models.OnfTrafficClass_TrafficClass_TrafficClass, error) {
	if device.TrafficClass == nil {
		return nil, fmt.Errorf("Device contains no Traffic Classes")
	}
	if (id == nil) || (*id == "") {
		return nil, fmt.Errorf("Traffic Class id is blank")
	}
	tc, okay := device.TrafficClass.TrafficClass[*id]
	if !okay {
		return nil, fmt.Errorf("TrafficClass %s not found", *id)
	}
	return tc, nil
}

// GetDeviceGroupSite gets the site for a DeviceGroup
func (s *Synchronizer) GetDeviceGroupSite(device *models.Device, dg *models.OnfDeviceGroup_DeviceGroup_DeviceGroup) (*models.OnfSite_Site_Site, error) {
	if (dg.Site == nil) || (*dg.Site == "") {
		return nil, fmt.Errorf("DeviceGroup %s has no site", *dg.Id)
	}
	site, okay := device.Site.Site[*dg.Site]
	if !okay {
		return nil, fmt.Errorf("DeviceGroup %s site %s not found", *dg.Id, *dg.Site)
	}
	if (site.Enterprise == nil) || (*site.Enterprise == "") {
		return nil, fmt.Errorf("DeviceGroup %s has no enterprise", *dg.Id)
	}
	return site, nil
}

// GetVcsDG given a VCS, return the set of DeviceGroup attached to it
func (s *Synchronizer) GetVcsDG(device *models.Device, vcs *models.OnfVcs_Vcs_Vcs) ([]*models.OnfDeviceGroup_DeviceGroup_DeviceGroup, error) {
	dgList := []*models.OnfDeviceGroup_DeviceGroup_DeviceGroup{}
	for _, dgLink := range vcs.DeviceGroup {
		if !*dgLink.Enable {
			continue
		}
		dg, okay := device.DeviceGroup.DeviceGroup[*dgLink.DeviceGroup]
		if !okay {
			return nil, fmt.Errorf("Vcs %s deviceGroup %s not found", *vcs.Id, *dgLink.DeviceGroup)
		}
		if (dg.Site == nil) || (*dg.Site == "") {
			return nil, fmt.Errorf("Vcs %s deviceGroup %s has no site", *vcs.Id, *dgLink.DeviceGroup)
		}

		dgList = append(dgList, dg)

		if *dgList[0].Site != *dg.Site {
			return nil, fmt.Errorf("Vcs %s deviceGroups %s and %s have different sites", *vcs.Id, *dgList[0].Site, *dg.Site)
		}
	}

	if len(dgList) == 0 {
		return nil, fmt.Errorf("VCS %s has no deviceGroups", *vcs.Id)
	}

	return dgList, nil
}

// GetVcsDGAndSite given a VCS, return the set of DeviceGroup attached to it, and the Site.
func (s *Synchronizer) GetVcsDGAndSite(device *models.Device, vcs *models.OnfVcs_Vcs_Vcs) ([]*models.OnfDeviceGroup_DeviceGroup_DeviceGroup, *models.OnfSite_Site_Site, error) {
	dgList, err := s.GetVcsDG(device, vcs)
	if err != nil {
		return nil, nil, err
	}

	site, err := s.GetDeviceGroupSite(device, dgList[0])
	if err != nil {
		return nil, nil, err
	}

	return dgList, site, nil
}

// SynchronizeDeviceGroups synchronizes the device groups
func (s *Synchronizer) SynchronizeDeviceGroups(device *models.Device, cs *models.OnfConnectivityService_ConnectivityService_ConnectivityService, validEnterpriseIds map[string]bool) error {
	pushFailures := 0
deviceGroupLoop:
	for _, dg := range device.DeviceGroup.DeviceGroup {
		site, err := s.GetDeviceGroupSite(device, dg)
		if err != nil {
			log.Warnf("DeviceGroup %s unable to determine site: %s", *dg.Id, err)
			continue deviceGroupLoop
		}
		valid, okay := validEnterpriseIds[*site.Enterprise]
		if (!okay) || (!valid) {
			log.Infof("DeviceGroup %s is not part of ConnectivityService %s.", *dg.Id, *cs.Id)
			continue deviceGroupLoop
		}

		dgCore := deviceGroup{
			IPDomainName: *dg.Id,
			SiteInfo:     *dg.Site,
		}

		if site.ImsiDefinition == nil {
			log.Warnf("DeviceGroup %s site has nil ImsiDefinition", *dg.Id)
			continue deviceGroupLoop
		}
		err = validateImsiDefinition(site.ImsiDefinition)
		if err != nil {
			log.Warnf("DeviceGroup %s unable to determine Site.ImsiDefinition: %s", *dg.Id, err)
			continue deviceGroupLoop
		}

		// populate the imsi list
		for _, imsiBlock := range dg.Imsis {
			if imsiBlock.ImsiRangeFrom == nil {
				log.Infof("imsiBlock has blank ImsiRangeFrom: %v", imsiBlock)
				continue deviceGroupLoop
			}
			var firstImsi uint64
			firstImsi, err = FormatImsiDef(site.ImsiDefinition, *imsiBlock.ImsiRangeFrom)
			if err != nil {
				log.Infof("Failed to format IMSI in dg %s: %v", *dg.Id, err)
				continue deviceGroupLoop
			}
			var lastImsi uint64
			if imsiBlock.ImsiRangeTo == nil {
				lastImsi = firstImsi
			} else {
				lastImsi, err = FormatImsiDef(site.ImsiDefinition, *imsiBlock.ImsiRangeTo)
				if err != nil {
					log.Infof("Failed to format IMSI in dg %s: %v", *dg.Id, err)
					continue deviceGroupLoop
				}

			}
			for i := firstImsi; i <= lastImsi; i++ {
				dgCore.Imsis = append(dgCore.Imsis, fmt.Sprintf("%015d", i))
			}
		}

		ipd, err := s.GetIPDomain(device, dg.IpDomain)
		if err != nil {
			log.Warnf("DeviceGroup %s failed to get IpDomain: %s", *dg.Id, err)
			continue deviceGroupLoop
		}

		err = validateIPDomain(ipd)
		if err != nil {
			log.Warnf("DeviceGroup %s invalid: %s", *dg.Id, err)
			continue deviceGroupLoop
		}

		dgCore.IPDomainName = *ipd.Id
		ipdCore := ipDomain{
			Dnn:          synchronizer.DerefStrPtr(ipd.Dnn, "internet"),
			Pool:         *ipd.Subnet,
			DNSPrimary:   synchronizer.DerefStrPtr(ipd.DnsPrimary, ""),
			DNSSecondary: synchronizer.DerefStrPtr(ipd.DnsSecondary, ""),
			Mtu:          synchronizer.DerefUint16Ptr(ipd.Mtu, DefaultMTU),
		}
		dgCore.IPDomain = ipdCore

		// TODO: This reflects that per-ue limits are modeled as part of the VCS
		// rather than part of the DG. So we go off and look for VCS that uses
		// this DG, and grabs its QOS settings. This will be revised.
		vcs := s.GetReferencingVCS(device, dg)
		if vcs != nil {
			dgCore.IPDomain.Qos = &ipdQos{}
			if vcs.Device != nil {
				if vcs.Device.Mbr != nil {
					if vcs.Device.Mbr.Uplink != nil {
						dgCore.IPDomain.Qos.Uplink = *vcs.Device.Mbr.Uplink
					}
					if vcs.Device.Mbr.Downlink != nil {
						dgCore.IPDomain.Qos.Downlink = *vcs.Device.Mbr.Downlink
					}
				}
			}

			if vcs.TrafficClass != nil {
				trafficClass, err := s.GetTrafficClass(device, vcs.TrafficClass)
				if err != nil {
					log.Warnf("Vcs %s unable to determine traffic class: %s", *vcs.Id, err)
					continue deviceGroupLoop
				}
				dgCore.IPDomain.Qos.TrafficClass.Name = *trafficClass.Id
				dgCore.IPDomain.Qos.TrafficClass.PDB = 300 // synchronizer.DerefUint16Ptr(trafficClass.PDB,300)
				dgCore.IPDomain.Qos.TrafficClass.PELR = 6  // synchronizer.DerefUint8Ptr(trafficClass.PELR,6)
				dgCore.IPDomain.Qos.TrafficClass.QCI = synchronizer.DerefUint8Ptr(trafficClass.Qci, 9)
				dgCore.IPDomain.Qos.TrafficClass.ARP = synchronizer.DerefUint8Ptr(trafficClass.Arp, 9)
			}
		}

		data, err := json.MarshalIndent(dgCore, "", "  ")
		if err != nil {
			log.Warnf("DeviceGroup %s failed to Marshal Json: %s", *dg.Id, err)
			continue deviceGroupLoop
		}

		url := fmt.Sprintf("%s/v1/device-group/%s", *cs.Core_5GEndpoint, *dg.Id)
		err = s.pusher.PushUpdate(url, data)
		if err != nil {
			log.Warnf("DeviceGroup %s failed to Push update: %s", *dg.Id, err)
			pushFailures++
			continue deviceGroupLoop
		}
	}
	if pushFailures > 0 {
		return fmt.Errorf("%d errors while pushing DeviceGroups", pushFailures)
	}
	return nil
}

// SynchronizeVcs synchronizes the VCSes
func (s *Synchronizer) SynchronizeVcs(device *models.Device, cs *models.OnfConnectivityService_ConnectivityService_ConnectivityService, validEnterpriseIds map[string]bool) error {
	pushFailures := 0
vcsLoop:
	for _, vcs := range device.Vcs.Vcs {
		dgList, site, err := s.GetVcsDGAndSite(device, vcs)
		if err != nil {
			log.Warnf("Vcs %s unable to determine site: %s", *vcs.Id, err)
			continue vcsLoop
		}
		valid, okay := validEnterpriseIds[*site.Enterprise]
		if (!okay) || (!valid) {
			log.Infof("VCS %s is not part of ConnectivityService %s.", *vcs.Id, *cs.Id)
			continue vcsLoop
		}

		err = validateVcs(vcs)
		if err != nil {
			log.Warnf("Vcs %s is invalid: %s", err)
			continue vcsLoop
		}

		if site.ImsiDefinition == nil {
			log.Warn("Vcs %s has nnil Site.ImsiDefinition", *vcs.Id)
			continue vcsLoop
		}
		err = validateImsiDefinition(site.ImsiDefinition)
		if err != nil {
			log.Warnf("Vcs %s unable to determine Site.ImsiDefinition: %s", *vcs.Id, err)
			continue vcsLoop
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
			for _, ap := range site.SmallCell {
				err = validateSmallCell(ap)
				if err != nil {
					log.Warnf("SmallCell invalid: %s", err)
					continue vcsLoop
				}
				if *ap.Enable {
					tac, err := strconv.ParseUint(*ap.Tac, 16, 32)
					if err != nil {
						log.Warnf("SmallCell Failed to convert tac %s to integer: %v", *ap.Tac, err)
						continue vcsLoop
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
				log.Warnf("Vcs %s unable to determine upf: %s", *vcs.Id, err)
				continue vcsLoop
			}
			err = validateUpf(aUpf)
			if err != nil {
				log.Warnf("Vcs %s Upf is invalid: %s", *vcs.Id, err)
				continue vcsLoop
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
		for k, _ := range vcs.Filter {
			appKeys = append(appKeys, k)
		}
		sort.Strings(appKeys)

		for _, k := range appKeys {
			appRef := vcs.Filter[k]
			app, err := s.GetApplication(device, appRef.Application)
			if err != nil {
				log.Warnf("Vcs %s unable to determine application: %s", *vcs.Id, err)
				continue vcsLoop
			}
			appCore := appFilterRule{
				Name: *app.Id,
			}

			/*
				type appFilterRule struct {
					Name          string `json:"rule-name"`
					Priority      uint8  `json:"priority"`
					Action        string `json:"action"`
					DestNetwork   string `json:"dest-network"`
					DestPortStart uint16 `json:"dest-port-start"`
					DestPortEnd   uint16 `json:"dest-port-end"`
					Protocol      uint32 `json:"protocol"`
				}
			*/

			if (app.Address == nil) || (*app.Address == "") {
				// this is a temporary restriction
				log.Warnf("Vcs %s Application %s has empty address", *vcs.Id, *app.Id)
				continue vcsLoop
			}
			if len(app.Endpoint) > 1 {
				// this is a temporary restriction
				log.Warnf("Vcs %s Application %s has more endpoints than are allowed", *vcs.Id, *app.Id)
				continue vcsLoop
			}
			// there can be at most one at this point...
			for _, endpoint := range app.Endpoint {
				err = validateAppEndpoint(endpoint)
				if err != nil {
					log.Warnf("App %s invalid endpoint: %s", *app.Id, err)
					continue vcsLoop
				}
				if strings.Contains(*app.Address, "/") {
					appCore.DestNetwork = *app.Address
				} else {
					appCore.DestNetwork = *app.Address + "/32"
				}

				appCore.DestPortStart = *endpoint.PortStart
				if endpoint.PortEnd != nil {
					appCore.DestPortEnd = synchronizer.DerefUint16Ptr(endpoint.PortEnd, 0)
				} else {
					// no EndPort specified -- assume it's a singleton range
					appCore.DestPortEnd = appCore.DestPortStart
				}

				protoNum, err := ProtoStringToProtoNumber(synchronizer.DerefStrPtr(endpoint.Protocol, DefaultProtocol))
				if err != nil {
					log.Warnf("Vcs %s Application %s unable to determine protocol: %s", *vcs.Id, *app.Id, err)
					continue vcsLoop
				}
				appCore.Protocol = protoNum

				if (appRef.Allow != nil) && (*appRef.Allow) {
					appCore.Action = "permit"
				} else {
					appCore.Action = "deny"
				}

				appCore.Priority = synchronizer.DerefUint8Ptr(appRef.Priority, 0)
			}
			slice.ApplicationFilteringRules = append(slice.ApplicationFilteringRules, appCore)
		}

		data, err := json.MarshalIndent(slice, "", "  ")
		if err != nil {
			log.Warnf("Vcs %s failed to marshal JSON: %s", *vcs.Id, err)
			continue vcsLoop
		}

		url := fmt.Sprintf("%s/v1/network-slice/%s", *cs.Core_5GEndpoint, *vcs.Id)
		err = s.pusher.PushUpdate(url, data)
		if err != nil {
			log.Warnf("Vcs %s failed to push update: %s", *vcs.Id, err)
			pushFailures++
			continue vcsLoop
		}
	}
	if pushFailures > 0 {
		return fmt.Errorf("%d errors while pushing VCS", pushFailures)
	}

	return nil
}
