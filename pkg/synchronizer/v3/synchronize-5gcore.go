// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizer implements a synchronizer for converting sdcore gnmi to json
package synchronizerv3

import (
	//"bytes"
	"encoding/json"
	//"errors"
	"fmt"
	"github.com/openconfig/ygot/ygot"
	//"net/http"
	//"os"
	"strconv"
	"time"

	models "github.com/onosproject/config-models/modelplugin/aether-3.0.0/aether_3_0_0"
)

type IpDomain struct {
	Dnn         string `json:"dnn"`
	Pool        string `json:"ue-ip-pool"`
	AdminStatus string `json:"admin-status"`
	DnsPrimary  string `json:"dns-primary"`
	Mtu         uint32 `json:"mtu"`
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
	Uplink   uint64 `json:"uplink"`
	Downlink uint64 `json:"downlink"`
}

type GNodeB struct {
	Name string `json:"name"`
	Tac  uint32 `json:"tac"`
}

type Plmn struct {
	Mcc uint32 `json:"mcc"`
	Mnc uint32 `json:"mnc"`
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
	DeviceGroup       string        `json:"device-group"`
	SiteInfo          SiteInfo      `json:"site-info"`
	DenyApplication   []string      `json:"deny-application"`
	PermitApplication []string      `json:"permitted-applications"`
	Applications      []Application `json:"application-information"`
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
		// Get the list of valid Enterprises for this CS.
		// Note: This could return an empty map if there is a CS that no
		//   enterprises are linked to . In that case, we can still push models
		//   that are not directly related to an enterprise, such as profiles.
		m := csEntMap[csId]

		tStart := time.Now()
		synchronizationTotal.WithLabelValues(csId).Inc()

		err := s.SynchronizeConnectivityService(device, cs, m)
		if err != nil {
			synchronizationFailedTotal.WithLabelValues(csId).Inc()
			// If there are errors, then build a list of them and continue to try
			// to synchronize other connectivity services.
			errors = append(errors, err)
		} else {
			synchronizationDuration.WithLabelValues(csId).Observe(time.Since(tStart).Seconds())
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
		s.SynchronizeDeviceGroups(device, cs, validEnterpriseIds)
	}
	if device.Vcs != nil {
		//SynchronizeVCS(device.Vcs)
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
		return nil, fmt.Errorf("Network %s not found", id)
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
		return nil, fmt.Errorf("IpDomain %s not found", id)
	}
	return ipd, nil
}

func (s *Synchronizer) GetDeviceGroupSite(device *models.Device, dg *models.DeviceGroup_DeviceGroup_DeviceGroup) (*models.Site_Site_Site, error) {
	if (dg.Site == nil) || (*dg.Site == "") {
		return nil, fmt.Errorf("DeviceGroup %s has no site.", *dg.Id)
		return nil, nil
	}
	site, okay := device.Site.Site[*dg.Site]
	if !okay {
		return nil, fmt.Errorf("DeviceGroup %s site %s not found.", *dg.Id, *dg.Site)
	}
	if (site.Enterprise == nil) || (*site.Enterprise == "") {
		return nil, fmt.Errorf("DeviceGroup %s has no enterprise.", *dg.Id)
		return nil, nil
	}
	return site, nil
}

func (s *Synchronizer) GetVcsSite(device *models.Device, vcs *models.Vcs_Vcs_Vcs) (*models.DeviceGroup_DeviceGroup_DeviceGroup, *models.Site_Site_Site, error) {
	if (vcs.DeviceGroup == nil) || (*vcs.DeviceGroup == "") {
		return nil, nil, fmt.Errorf("VCS %s has no deviceGroup.", *vcs.Id)
	}
	dg, okay := device.DeviceGroup.DeviceGroup[*vcs.DeviceGroup]
	if !okay {
		return nil, nil, fmt.Errorf("Vcs %s deviceGroup %s not found.", *vcs.Id, *vcs.DeviceGroup)
	}
	site, err := s.GetDeviceGroupSite(device, dg)
	if err != nil {
		return nil, nil, err
	}
	return dg, site, err
}

func (s *Synchronizer) SynchronizeDeviceGroups(device *models.Device, cs *models.ConnectivityService_ConnectivityService_ConnectivityService, validEnterpriseIds map[string]bool) error {
	for _, dg := range device.DeviceGroup.DeviceGroup {
		site, err := s.GetDeviceGroupSite(device, dg)
		if err != nil {
			log.Warnf("DeviceGroup %s unable to determine site: %s", *dg.Id, err)
			continue
		}
		valid, okay := validEnterpriseIds[*site.Enterprise]
		if (!okay) || (!valid) {
			log.Infof("DeviceGroup %s is not part of ConnectivityService %s.", *dg.Id, *cs.Id)
			continue
		}

		dgCore := DeviceGroup{
			IpDomainName: *dg.Id,
			SiteInfo:     *dg.Site,
		}

		// populate the imsi list
		for _, imsiBlock := range dg.Imsis {
			var lastImsi uint64
			if imsiBlock.ImsiRangeFrom == nil {
				// print error?
				continue
			}
			if imsiBlock.ImsiRangeTo == nil {
				lastImsi = *imsiBlock.ImsiRangeFrom
			}
			for i := *imsiBlock.ImsiRangeFrom; i <= lastImsi; i++ {
				dgCore.Imsis = append(dgCore.Imsis, strconv.FormatUint(i, 10))
			}
		}

		ipd, err := s.GetIpDomain(device, dg.IpDomain)
		if err != nil {
			log.Warnf("DeviceGroup %s unable to determine ipDomain: %s", *dg.Id, err)
			continue
		}

		dgCore.IpDomainName = *ipd.Id
		ipdCore := IpDomain{
			Dnn:         "Internet", // hardcoded
			Pool:        *ipd.Subnet,
			AdminStatus: *ipd.AdminStatus,
			DnsPrimary:  *ipd.DnsPrimary,
			Mtu:         *ipd.Mtu,
		}
		dgCore.IpDomain = ipdCore

		data, err := json.MarshalIndent(dgCore, "", "  ")
		if err != nil {
			return err
		}

		log.Infof("Put IpDomain: %v", string(data))
	}
	return nil
}

func (s *Synchronizer) SynchronizeVcs(device *models.Device, cs *models.ConnectivityService_ConnectivityService_ConnectivityService, validEnterpriseIds map[string]bool) error {
	for _, vcs := range device.Vcs.Vcs {
		dg, site, err := s.GetVcsSite(device, vcs)
		if err != nil {
			log.Warnf("Vcs %s unable to determine site", *vcs.Id)
			continue
		}
		valid, okay := validEnterpriseIds[*site.Enterprise]
		if (!okay) || (!valid) {
			log.Infof("VCS %s is not part of ConnectivityService %s.", *vcs.Id, *cs.Id)
			continue
		}

		net, err := s.GetNetwork(device, site.Network)
		if err != nil {
			log.Warnf("Vcs %s unable to determine network: %s", *vcs.Id, err)
			continue
		}

		plmn := Plmn{
			Mcc: *net.Mcc,
			Mnc: *net.Mnc,
		}
		siteInfo := SiteInfo{
			SiteName: *site.Id,
			Plmn:     plmn,
		}
		sliceId := SliceId{
			Sst: *vcs.Sst,
			Sd:  *vcs.Sd,
		}
		slice := Slice{
			Id:          sliceId,
			DeviceGroup: *dg.Id,
			SiteInfo:    siteInfo}
	}

	/*
		SiteName string   `json:"site-name"`
		Plmn     Plmn     `json:"plmn"`
		GNodeBs  []GNodeB `json:"gNodeBs"`
		Upf      Upf      `json:"upf"`

			type Slice struct {
			Id                SliceId       `json:"slice-id"`
			Qos               Qos           `json:"qos"`
			DeviceGroup       string        `json:"device-group"`
			SiteInfo          SiteInfo      `json:"site-info"`
			DenyApplication   []string      `json:"deny-application"`
			PermitApplication []string      `json:"permitted-applications"`
			Applications      []Application `json:"application-information"`
	*/

	return nil
}
