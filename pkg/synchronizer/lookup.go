// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizer implements a synchronizer for converting sdcore gnmi to json
package synchronizer

import (
	"fmt"
	"sort"

	models "github.com/onosproject/config-models/modelplugin/aether-2.0.0/aether_2_0_0"
)

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

// GetDeviceGroup looks up a DeviceGroup
func (s *Synchronizer) GetDeviceGroup(device *models.Device, id *string) (*models.OnfDeviceGroup_DeviceGroup_DeviceGroup, error) {
	if (id == nil) || (*id == "") {
		return nil, fmt.Errorf("DeviceGroup id is blank")
	}
	if device.DeviceGroup == nil {
		return nil, fmt.Errorf("No DeviceGroups")
	}
	dg, okay := device.DeviceGroup.DeviceGroup[*id]
	if !okay {
		return nil, fmt.Errorf("DeviceGroup %s not found", *id)
	}
	return dg, nil
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

// GetEnterprise looks up an Enterprise
func (s *Synchronizer) GetEnterprise(device *models.Device, id *string) (*models.OnfEnterprise_Enterprise_Enterprise, error) {
	if (id == nil) || (*id == "") {
		return nil, fmt.Errorf("Enterprise id is blank")
	}
	if device.Enterprise == nil {
		return nil, fmt.Errorf("No Enterprises")
	}
	ent, okay := device.Enterprise.Enterprise[*id]
	if !okay {
		return nil, fmt.Errorf("Enterprise %s not found", *id)
	}
	return ent, nil
}

// GetSite looks up a Site
func (s *Synchronizer) GetSite(device *models.Device, id *string) (*models.OnfSite_Site_Site, error) {
	if (id == nil) || (*id == "") {
		return nil, fmt.Errorf("Site id is blank")
	}
	if device.Site == nil {
		return nil, fmt.Errorf("No Sites")
	}
	site, okay := device.Site.Site[*id]
	if !okay {
		return nil, fmt.Errorf("Site %s not found", *id)
	}
	if (site.Enterprise == nil) || (*site.Enterprise == "") {
		return nil, fmt.Errorf("Site %s has no enterprise", *id)
	}
	return site, nil
}

// GetVcs looks up a VCS
func (s *Synchronizer) GetVcs(device *models.Device, id *string) (*models.OnfVcs_Vcs_Vcs, error) {
	if (id == nil) || (*id == "") {
		return nil, fmt.Errorf("Vcs id is blank")
	}
	if device.Vcs == nil {
		return nil, fmt.Errorf("No VCSes")
	}
	vcs, okay := device.Vcs.Vcs[*id]
	if !okay {
		return nil, fmt.Errorf("Vcs %s not found", *id)
	}
	return vcs, nil
}

// GetConnectivityService looks up a Connectivity Service
func (s *Synchronizer) GetConnectivityService(device *models.Device, id *string) (*models.OnfConnectivityService_ConnectivityService_ConnectivityService, error) {
	if (id == nil) || (*id == "") {
		return nil, fmt.Errorf("ConnectivityService id is blank")
	}
	if device.ConnectivityService == nil {
		return nil, fmt.Errorf("No connectivity services")
	}
	cs, okay := device.ConnectivityService.ConnectivityService[*id]
	if !okay {
		return nil, fmt.Errorf("ConnectivityService %s not found", *id)
	}
	return cs, nil
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

	// be deterministic...
	dgKeys := []string{}
	for k := range vcs.DeviceGroup {
		dgKeys = append(dgKeys, k)
	}
	sort.Strings(dgKeys)

	for _, k := range dgKeys {
		dgLink := vcs.DeviceGroup[k]

		dg, okay := device.DeviceGroup.DeviceGroup[*dgLink.DeviceGroup]
		if !okay {
			return nil, fmt.Errorf("Vcs %s deviceGroup %s not found", *vcs.Id, *dgLink.DeviceGroup)
		}
		if (dg.Site == nil) || (*dg.Site == "") {
			return nil, fmt.Errorf("Vcs %s deviceGroup %s has no site", *vcs.Id, *dgLink.DeviceGroup)
		}

		if len(dgList) > 0 && (*dgList[0].Site != *dg.Site) {
			return nil, fmt.Errorf("Vcs %s deviceGroups %s and %s have different sites", *vcs.Id, *dgList[0].Site, *dg.Site)
		}

		// Only add it to the list if it's enabled
		if *dgLink.Enable {
			dgList = append(dgList, dg)
		}
	}

	return dgList, nil
}

// GetConnectivityServiceForSite given a siteName returns a list of connectivity services
func (s *Synchronizer) GetConnectivityServiceForSite(device *models.Device, siteName *string) ([]*models.OnfConnectivityService_ConnectivityService_ConnectivityService, error) {
	site, err := s.GetSite(device, siteName)
	if err != nil {
		return nil, err
	}

	ent, err := s.GetEnterprise(device, site.Enterprise)
	if err != nil {
		return nil, err
	}

	eligibleCS := []*models.OnfConnectivityService_ConnectivityService_ConnectivityService{}
	for csID, cs := range ent.ConnectivityService {
		if !*cs.Enabled {
			continue
		}
		csModel, err := s.GetConnectivityService(device, &csID)
		if err != nil {
			return nil, err
		}
		eligibleCS = append(eligibleCS, csModel)
	}

	return eligibleCS, nil
}
