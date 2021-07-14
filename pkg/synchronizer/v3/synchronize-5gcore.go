// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizer implements a synchronizer for converting sdcore gnmi to json
package synchronizerv3

import (
	"encoding/json"
	"fmt"
	"github.com/openconfig/ygot/ygot"
	"strconv"
	"strings"
	"time"

	models "github.com/onosproject/config-models/modelplugin/aether-3.0.0/aether_3_0_0"
	"github.com/onosproject/sdcore-adapter/pkg/synchronizer"
)

// Ideally we would get these from the yang defaults
const (
	DEFAULT_ADMINSTATUS = "ENABLE"
	DEFAULT_MTU         = 1492
	DEFAULT_PROTOCOL    = "TCP"
)

type IpDomain struct {
	Dnn  string `json:"dnn"`
	Pool string `json:"ue-ip-pool"`
	// AdminStatus string `json:"admin-status"`  Dropped from current JSON
	DnsPrimary string `json:"dns-primary"`
	Mtu        uint32 `json:"mtu"`
}

type DeviceGroup struct {
	Imsis        []string `json:"imsis"`
	IpDomainName string   `json:"ip-domain-name"`
	SiteInfo     string   `json:"site-info"`
	IpDomain     IpDomain `json:"ip-domain-expanded"`
}

type SliceId struct {
	Sst uint32 `json:"sst"`
	Sd  uint32 `json:"sd"`
}

type Qos struct {
	Uplink       uint64 `json:"uplink"`
	Downlink     uint64 `json:"downlink"`
	TrafficClass string `json:"traffic-class"`
}

type GNodeB struct {
	Name string `json:"name"`
	Tac  uint32 `json:"tac"`
}

type Plmn struct {
	Mcc string `json:"mcc"`
	Mnc string `json:"mnc"`
}

type Upf struct {
	Name string `json:"upf-name"`
	Port uint32 `json:"upf-port"`
}

type SiteInfo struct {
	SiteName string   `json:"site-name"`
	Plmn     Plmn     `json:"plmn"`
	GNodeBs  []GNodeB `json:"gNodeBs"`
	Upf      Upf      `json:"upf"`
}

type Application struct {
	Name      string `json:"app-name"`
	Endpoint  string `json:"endpoint"`
	StartPort uint32 `json:"start-port"`
	EndPort   uint32 `json:"end-port"`
	Protocol  uint32 `json:"protocol"`
}

type Slice struct {
	Id                SliceId       `json:"slice-id"`
	Qos               Qos           `json:"qos"`
	DeviceGroup       []string      `json:"site-device-group"`
	SiteInfo          SiteInfo      `json:"site-info"`
	DenyApplication   []string      `json:"deny-applications"`
	PermitApplication []string      `json:"permit-applications"`
	Applications      []Application `json:"applications-information"`
}

func ProtoStringToProtoNumber(s string) (uint32, error) {
	n, okay := map[string]uint32{"TCP": 6, "UDP": 17}[s]
	if !okay {
		return 0, fmt.Errorf("Unknown protocol %s", s)
	}
	return n, nil
}

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
	for entId, ent := range device.Enterprise.Enterprise {
		for csId := range ent.ConnectivityService {
			m, okay := csEntMap[csId]
			if !okay {
				m = map[string]bool{}
				csEntMap[csId] = m
			}
			m[entId] = true
		}
	}

	errors := []error{}
	for csId, cs := range device.ConnectivityService.ConnectivityService {
		if (cs.Core_5GEndpoint == nil) || (*cs.Core_5GEndpoint == "") {
			log.Warnf("Skipping connectivity service %s because it has no 5G Endpoint", *cs.Id)
			continue
		}

		// Get the list of valid Enterprises for this CS.
		// Note: This could return an empty map if there is a CS that no
		//   enterprises are linked to . In that case, we can still push models
		//   that are not directly related to an enterprise, such as profiles.
		m := csEntMap[csId]

		tStart := time.Now()
		synchronizer.KpiSynchronizationTotal.WithLabelValues(csId).Inc()

		err := s.SynchronizeConnectivityService(device, cs, m)
		if err != nil {
			synchronizer.KpiSynchronizationFailedTotal.WithLabelValues(csId).Inc()
			// If there are errors, then build a list of them and continue to try
			// to synchronize other connectivity services.
			errors = append(errors, err)
		} else {
			synchronizer.KpiSynchronizationDuration.WithLabelValues(csId).Observe(time.Since(tStart).Seconds())
		}
	}

	if len(errors) == 0 {
		return nil
	} else {
		return fmt.Errorf("synchronization errors: %v", errors)
	}
}

func (s *Synchronizer) SynchronizeConnectivityService(device *models.Device, cs *models.ConnectivityService_ConnectivityService_ConnectivityService, validEnterpriseIds map[string]bool) error {
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

// Lookup a network.
func (s *Synchronizer) GetNetwork(device *models.Device, id *string) (*models.Network_Network_Network, error) {
	if device.Network == nil {
		return nil, fmt.Errorf("Device contains no networks")
	}
	if (id == nil) || (*id == "") {
		return nil, fmt.Errorf("Network id is blank")
	}
	net, okay := device.Network.Network[*id]
	if !okay {
		return nil, fmt.Errorf("Network %s not found", *id)
	}
	return net, nil
}

// Lookup an IpDomain
func (s *Synchronizer) GetIpDomain(device *models.Device, id *string) (*models.IpDomain_IpDomain_IpDomain, error) {
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

// Lookup an ApList
func (s *Synchronizer) GetApList(device *models.Device, id *string) (*models.ApList_ApList_ApList, error) {
	if device.ApList == nil {
		return nil, fmt.Errorf("Device contains no ApLists")
	}
	if (id == nil) || (*id == "") {
		return nil, fmt.Errorf("ApList id is blank")
	}
	apl, okay := device.ApList.ApList[*id]
	if !okay {
		return nil, fmt.Errorf("ApList %s not found", *id)
	}
	return apl, nil
}

// Lookup a UPF
func (s *Synchronizer) GetUpf(device *models.Device, id *string) (*models.Upf_Upf_Upf, error) {
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

// Lookup an Application
func (s *Synchronizer) GetApplication(device *models.Device, id *string) (*models.Application_Application_Application, error) {
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

// Lookup an TrafficClass
func (s *Synchronizer) GetTrafficClass(device *models.Device, id *string) (*models.TrafficClass_TrafficClass_TrafficClass, error) {
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

func (s *Synchronizer) GetDeviceGroupSite(device *models.Device, dg *models.DeviceGroup_DeviceGroup_DeviceGroup) (*models.Site_Site_Site, error) {
	if (dg.Site == nil) || (*dg.Site == "") {
		return nil, fmt.Errorf("DeviceGroup %s has no site.", *dg.Id)
	}
	site, okay := device.Site.Site[*dg.Site]
	if !okay {
		return nil, fmt.Errorf("DeviceGroup %s site %s not found.", *dg.Id, *dg.Site)
	}
	if (site.Enterprise == nil) || (*site.Enterprise == "") {
		return nil, fmt.Errorf("DeviceGroup %s has no enterprise.", *dg.Id)
	}
	return site, nil
}

// Given a VCS, return the set of DeviceGroup attached to it, and the Site.
func (s *Synchronizer) GetVcsDGAndSite(device *models.Device, vcs *models.Vcs_Vcs_Vcs) ([]*models.DeviceGroup_DeviceGroup_DeviceGroup, *models.Site_Site_Site, error) {
	dgList := []*models.DeviceGroup_DeviceGroup_DeviceGroup{}
	for _, dgLink := range vcs.DeviceGroup {
		if !*dgLink.Enable {
			continue
		}
		dg, okay := device.DeviceGroup.DeviceGroup[*dgLink.DeviceGroup]
		if !okay {
			return nil, nil, fmt.Errorf("Vcs %s deviceGroup %s not found.", *vcs.Id, *dgLink.DeviceGroup)
		}
		if (dg.Site == nil) || (*dg.Site == "") {
			return nil, nil, fmt.Errorf("Vcs %s deviceGroup %s has no site.", *vcs.Id, *dgLink.DeviceGroup)
		}

		dgList = append(dgList, dg)

		if *dgList[0].Site != *dg.Site {
			return nil, nil, fmt.Errorf("Vcs %s deviceGroups %s and %s have different sites.", *vcs.Id, *dgList[0].Site, *dg.Site)
		}
	}

	if len(dgList) == 0 {
		return nil, nil, fmt.Errorf("VCS %s has no deviceGroups.", *vcs.Id)
	}

	site, err := s.GetDeviceGroupSite(device, dgList[0])
	if err != nil {
		return nil, nil, err
	}

	return dgList, site, err
}

func (s *Synchronizer) SynchronizeDeviceGroups(device *models.Device, cs *models.ConnectivityService_ConnectivityService_ConnectivityService, validEnterpriseIds map[string]bool) error {
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

		dgCore := DeviceGroup{
			IpDomainName: *dg.Id,
			SiteInfo:     *dg.Site,
		}

		// Latest modeling uses Site.ImsiDefinition to format the IMSI
		var imsiDef *models.Site_Site_Site_ImsiDefinition
		if site.Network == nil {
			if site.ImsiDefinition == nil {
				log.Warnf("DeviceGroup %s site has neither Network nor ImsiDefinition", *dg.Id)
				continue deviceGroupLoop
			}
			err = s.validateSiteImsiDefinition(site.ImsiDefinition)
			if err != nil {
				log.Warnf("DeviceGroup %s unable to determine Site.ImsiDefinition: %s", *dg.Id, err)
				continue deviceGroupLoop
			}
			imsiDef = site.ImsiDefinition
		}

		// populate the imsi list
		for _, imsiBlock := range dg.Imsis {
			if imsiBlock.ImsiRangeFrom == nil {
				log.Infof("imsiBlock has blank ImsiRangeFrom: %v", imsiBlock)
				continue deviceGroupLoop
			}
			var firstImsi uint64
			if imsiDef != nil {
				firstImsi, err = FormatImsiDef(imsiDef, *imsiBlock.ImsiRangeFrom)
				if err != nil {
					log.Infof("Failed to format IMSI in dg %s: %v", *dg.Id, err)
					continue deviceGroupLoop
				}
			} else {
				// DEPRECATED
				firstImsi = *imsiBlock.ImsiRangeFrom
			}
			var lastImsi uint64
			if imsiBlock.ImsiRangeTo == nil {
				lastImsi = firstImsi
			} else {
				if imsiDef != nil {
					lastImsi, err = FormatImsiDef(imsiDef, *imsiBlock.ImsiRangeTo)
					if err != nil {
						log.Infof("Failed to format IMSI in dg %s: %v", *dg.Id, err)
						continue deviceGroupLoop
					}
				} else {
					// DEPRECATED
					lastImsi = *imsiBlock.ImsiRangeTo
				}
			}
			for i := firstImsi; i <= lastImsi; i++ {
				dgCore.Imsis = append(dgCore.Imsis, strconv.FormatUint(i, 10))
			}
		}

		ipd, err := s.GetIpDomain(device, dg.IpDomain)
		if err != nil {
			log.Warnf("DeviceGroup %s failed to get IpDomain: %s", *dg.Id, err)
			continue deviceGroupLoop
		}

		err = s.validateIpDomain(ipd)
		if err != nil {
			log.Warnf("DeviceGroup %s invalid: %s", *dg.Id, err)
			continue deviceGroupLoop
		}

		dgCore.IpDomainName = *ipd.Id
		ipdCore := IpDomain{
			Dnn:  "Internet", // hardcoded
			Pool: *ipd.Subnet,
			// AdminStatus: synchronizer.DerefStrPtr(ipd.AdminStatus, DEFAULT_ADMINSTATUS),   Dropped from current JSON
			DnsPrimary: synchronizer.DerefStrPtr(ipd.DnsPrimary, ""),
			Mtu:        synchronizer.DerefUint32Ptr(ipd.Mtu, DEFAULT_MTU),
		}
		dgCore.IpDomain = ipdCore

		data, err := json.MarshalIndent(dgCore, "", "  ")
		if err != nil {
			log.Warnf("DeviceGroup %s failed to Marshal Json: %s", *dg.Id, err)
			continue deviceGroupLoop
		}

		url := fmt.Sprintf("%s/v1/device-group/%s", *cs.Core_5GEndpoint, *dg.Id)
		err = s.pusher.PushUpdate(url, data)
		if err != nil {
			log.Warnf("DeviceGroup %s failed to Push update: %s", *dg.Id, err)
			continue deviceGroupLoop
		}
	}
	return nil
}

func (s *Synchronizer) SynchronizeVcs(device *models.Device, cs *models.ConnectivityService_ConnectivityService_ConnectivityService, validEnterpriseIds map[string]bool) error {
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

		err = s.validateVcs(vcs)
		if err != nil {
			log.Warnf("Vcs %s is invalid: %s", err)
			continue vcsLoop
		}

		var mcc uint32
		var mnc uint32
		if site.Network != nil {
			// DEPRECATED
			net, err := s.GetNetwork(device, site.Network)
			if err != nil {
				log.Warnf("Vcs %s unable to determine network: %s", *vcs.Id, err)
				continue vcsLoop
			}

			err = s.validateNetwork(net)
			if err != nil {
				log.Warn("Vcs %s Network Invalid: %s", *vcs.Id)
				continue vcsLoop
			}
			mcc = *net.Mcc
			mnc = *net.Mnc
		} else {
			if site.ImsiDefinition == nil {
				log.Warn("Vcs %s has neither Site.Network nor Site.ImsiDefinition", *vcs.Id)
				continue vcsLoop
			}
			err := s.validateSiteImsiDefinition(site.ImsiDefinition)
			if err != nil {
				log.Warnf("Vcs %s unable to determine Site.ImsiDefinition: %s", *vcs.Id, err)
				continue vcsLoop
			}
			mcc = *site.ImsiDefinition.Mcc
			mnc = *site.ImsiDefinition.Mnc
		}

		plmn := Plmn{
			Mcc: strconv.FormatUint(uint64(mcc), 10),
			Mnc: strconv.FormatUint(uint64(mnc), 10),
		}
		siteInfo := SiteInfo{
			SiteName: *site.Id,
			Plmn:     plmn,
		}

		if vcs.Ap != nil {
			apList, err := s.GetApList(device, vcs.Ap)
			if err != nil {
				log.Warnf("Vcs %s unable to determine ap list: %s", *vcs.Id, err)
				continue vcsLoop
			}
			for _, ap := range apList.AccessPoints {
				err = s.validateAccessPoint(ap)
				if err != nil {
					log.Warnf("AccessPointList %s invalid: %s", *apList.Id, err)
					continue vcsLoop
				}
				if *ap.Enable {
					gNodeB := GNodeB{
						Name: *ap.Address,
						Tac:  *ap.Tac,
					}
					siteInfo.GNodeBs = append(siteInfo.GNodeBs, gNodeB)
				}
			}
		}

		if vcs.Upf != nil {
			upf, err := s.GetUpf(device, vcs.Upf)
			if err != nil {
				log.Warnf("Vcs %s unable to determine upf: %s", *vcs.Id, err)
				continue vcsLoop
			}
			err = s.validateUpf(upf)
			if err != nil {
				log.Warnf("Vcs %s Upf is invalid: %s", *vcs.Id, err)
				continue vcsLoop
			}
			siteInfo.Upf = Upf{
				Name: *upf.Address,
				Port: *upf.Port,
			}
		}

		sliceId := SliceId{
			Sst: *vcs.Sst,
			Sd:  *vcs.Sd,
		}

		slice := Slice{
			Id:                sliceId,
			SiteInfo:          siteInfo,
			PermitApplication: []string{},
			DenyApplication:   []string{},
		}

		for _, dg := range dgList {
			slice.DeviceGroup = append(slice.DeviceGroup, *dg.Id)
		}

		// TODO: These should be uint64 in the modeling
		if vcs.Uplink != nil {
			slice.Qos.Uplink = uint64(*vcs.Uplink)
		}
		if vcs.Downlink != nil {
			slice.Qos.Downlink = uint64(*vcs.Downlink)
		}

		if vcs.TrafficClass != nil {
			trafficClass, err := s.GetTrafficClass(device, vcs.TrafficClass)
			if err != nil {
				log.Warnf("Vcs %s unable to determine traffic class: %s", *vcs.Id, err)
				continue vcsLoop
			}
			slice.Qos.TrafficClass = *trafficClass.Id
		}

		for _, appRef := range vcs.Application {
			app, err := s.GetApplication(device, appRef.Application)
			if err != nil {
				log.Warnf("Vcs %s unable to determine application: %s", *vcs.Id, err)
				continue vcsLoop
			}
			if *appRef.Allow {
				slice.PermitApplication = append(slice.PermitApplication, *app.Id)
			} else {
				slice.DenyApplication = append(slice.DenyApplication, *app.Id)
			}
			appCore := Application{
				Name: *app.Id,
			}
			if len(app.Endpoint) > 1 {
				// this is a temporary restriction
				log.Warnf("Vcs %s Application %s has more endpoints than are allowed", *vcs.Id, *app.Id)
				continue vcsLoop
			}
			// there can be at most one at this point...
			for _, endpoint := range app.Endpoint {
				err = s.validateAppEndpoint(endpoint)
				if err != nil {
					log.Warnf("App %s invalid endpoint: %s", *app.Id, err)
					continue vcsLoop
				}
				if strings.Contains(*endpoint.Address, "/") {
					appCore.Endpoint = *endpoint.Address
				} else {
					appCore.Endpoint = *endpoint.Address + "/32"
				}

				appCore.StartPort = *endpoint.PortStart
				if endpoint.PortEnd != nil {
					appCore.EndPort = synchronizer.DerefUint32Ptr(endpoint.PortEnd, 0)
				} else {
					// no EndPort specified -- assume it's a singleton range
					appCore.EndPort = appCore.StartPort
				}

				protoNum, err := ProtoStringToProtoNumber(synchronizer.DerefStrPtr(endpoint.Protocol, DEFAULT_PROTOCOL))
				if err != nil {
					log.Warnf("Vcs %s Application %s unable to determine protocol: %s", *vcs.Id, *app.Id, err)
					continue vcsLoop
				}
				appCore.Protocol = protoNum
			}
			slice.Applications = append(slice.Applications, appCore)
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
			continue vcsLoop
		}
	}

	return nil
}
