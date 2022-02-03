// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

// Package synchronizer implements a synchronizer for converting sdcore gnmi to json
package synchronizer

import (
	"fmt"
	"sort"
)

// GetIPDomain looks up an IpDomain
func (s *Synchronizer) GetIPDomain(scope *AetherScope, id *string) (*IpDomain, error) {
	if (id == nil) || (*id == "") {
		return nil, fmt.Errorf("IpDomain id is blank")
	}
	ipd, okay := scope.Site.IpDomain[*id]
	if !okay {
		return nil, fmt.Errorf("IpDomain %s not found", *id)
	}
	return ipd, nil
}

// GetUpf looks up a Upf
func (s *Synchronizer) GetUpf(scope *AetherScope, id *string) (*Upf, error) {
	if (id == nil) || (*id == "") {
		return nil, fmt.Errorf("Upf id is blank")
	}
	upf, okay := scope.Site.Upf[*id]
	if !okay {
		return nil, fmt.Errorf("Upf %s not found", *id)
	}
	return upf, nil
}

// GetApplication looks up an application
func (s *Synchronizer) GetApplication(scope *AetherScope, id *string) (*Application, error) {
	if (id == nil) || (*id == "") {
		return nil, fmt.Errorf("Application id is blank")
	}
	app, okay := scope.Enterprise.Application[*id]
	if !okay {
		return nil, fmt.Errorf("Application %s not found", *id)
	}
	return app, nil
}

// GetDeviceGroup looks up a DeviceGroup
func (s *Synchronizer) GetDeviceGroup(scope *AetherScope, id *string) (*DeviceGroup, error) {
	if (id == nil) || (*id == "") {
		return nil, fmt.Errorf("DeviceGroup id is blank")
	}
	dg, okay := scope.Site.DeviceGroup[*id]
	if !okay {
		return nil, fmt.Errorf("DeviceGroup %s not found", *id)
	}
	return dg, nil
}

// GetTrafficClass looks up a TrafficClass
func (s *Synchronizer) GetTrafficClass(scope *AetherScope, id *string) (*TrafficClass, error) {
	if (id == nil) || (*id == "") {
		return nil, fmt.Errorf("Traffic Class id is blank")
	}
	tc, okay := scope.Enterprise.TrafficClass[*id]
	if !okay {
		return nil, fmt.Errorf("TrafficClass %s not found", *id)
	}
	return tc, nil
}

// GetEnterprise looks up an Enterprise
func (s *Synchronizer) GetEnterprise(scope *AetherScope, id *string) (*Enterprise, error) {
	if (id == nil) || (*id == "") {
		return nil, fmt.Errorf("Enterprise id is blank")
	}
	if (scope.RootDevice.Enterprises == nil) || (scope.RootDevice.Enterprises.Enterprise == nil) {
		return nil, fmt.Errorf("No enterprises")
	}
	ent, okay := scope.RootDevice.Enterprises.Enterprise[*id]
	if !okay {
		return nil, fmt.Errorf("Enterprise %s not found", *id)
	}
	return ent, nil
}

// GetSite looks up a Site
func (s *Synchronizer) GetSite(scope *AetherScope, id *string) (*Site, error) {
	if (id == nil) || (*id == "") {
		return nil, fmt.Errorf("Site id is blank")
	}
	site, okay := scope.Enterprise.Site[*id]
	if !okay {
		return nil, fmt.Errorf("Site %s not found", *id)
	}
	return site, nil
}

// GetSlice looks up a Slice
func (s *Synchronizer) GetSlice(scope *AetherScope, id *string) (*Slice, error) {
	if (id == nil) || (*id == "") {
		return nil, fmt.Errorf("Slice id is blank")
	}
	slice, okay := scope.Site.Slice[*id]
	if !okay {
		return nil, fmt.Errorf("Slice %s not found", *id)
	}
	return slice, nil
}

// GetDevice looks up a Device
func (s *Synchronizer) GetDevice(scope *AetherScope, id *string) (*Device, error) {
	if (id == nil) || (*id == "") {
		return nil, fmt.Errorf("Device id is blank")
	}
	device, okay := scope.Site.Device[*id]
	if !okay {
		return nil, fmt.Errorf("Device %s not found", *id)
	}
	return device, nil
}

// GetSimCard looks up a SimCard
func (s *Synchronizer) GetSimCard(scope *AetherScope, id *string) (*SimCard, error) {
	if (id == nil) || (*id == "") {
		return nil, fmt.Errorf("SimCard id is blank")
	}
	simCard, okay := scope.Site.SimCard[*id]
	if !okay {
		return nil, fmt.Errorf("SimCard %s not found", *id)
	}
	return simCard, nil
}

// GetConnectivityService looks up a Connectivity Service
func (s *Synchronizer) GetConnectivityService(scope *AetherScope, id *string) (*ConnectivityService, error) {
	if (id == nil) || (*id == "") {
		return nil, fmt.Errorf("ConnectivityService id is blank")
	}
	if (scope.RootDevice.ConnectivityServices == nil) || (scope.RootDevice.ConnectivityServices.ConnectivityService == nil) {
		return nil, fmt.Errorf("No connectivity services")
	}
	cs, okay := scope.RootDevice.ConnectivityServices.ConnectivityService[*id]
	if !okay {
		return nil, fmt.Errorf("ConnectivityService %s not found", *id)
	}
	return cs, nil
}

// GetSliceDG given a Slice, return the set of DeviceGroup attached to it
func (s *Synchronizer) GetSliceDG(scope *AetherScope, slice *Slice) ([]*DeviceGroup, error) {
	dgList := []*DeviceGroup{}

	// be deterministic...
	dgKeys := []string{}
	for k := range slice.DeviceGroup {
		dgKeys = append(dgKeys, k)
	}
	sort.Strings(dgKeys)

	for _, k := range dgKeys {
		dgLink := slice.DeviceGroup[k]

		dg, okay := scope.Site.DeviceGroup[*dgLink.DeviceGroup]
		if !okay {
			return nil, fmt.Errorf("Slice %s deviceGroup %s not found", *slice.SliceId, *dgLink.DeviceGroup)
		}

		// Only add it to the list if it's enabled
		if *dgLink.Enable {
			dgList = append(dgList, dg)
		}
	}

	return dgList, nil
}

// GetConnectivityServicesForEnterprise given a siteName returns a list of connectivity services
func (s *Synchronizer) GetConnectivityServicesForEnterprise(scope *AetherScope) ([]*ConnectivityService, error) {
	eligibleCS := []*ConnectivityService{}
	for csID, cs := range scope.Enterprise.ConnectivityService {
		if !*cs.Enabled {
			continue
		}
		csModel, err := s.GetConnectivityService(scope, &csID)
		if err != nil {
			return nil, err
		}
		eligibleCS = append(eligibleCS, csModel)
	}

	return eligibleCS, nil
}
